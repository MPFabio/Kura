// Package fine implémente la détection de drift "fine" : exécution d'un
// `tofu plan -refresh-only` dans un sandbox temporaire à partir des fichiers
// .tf source et du tfstate, pour obtenir un diff générique multi-cloud.
package fine

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"

	"github.com/modulops/terraform-service/internal/client"
)

const defaultTofuPath = "/usr/local/bin/tofu"

// Runner exécute des plans OpenTofu dans un répertoire temporaire isolé.
type Runner struct {
	tofuPath string
}

// NewRunner crée un Runner. Si tofuPath est vide, utilise le chemin par défaut
// du binaire `tofu` bundlé dans l'image.
func NewRunner(tofuPath string) *Runner {
	if tofuPath == "" {
		tofuPath = defaultTofuPath
	}
	return &Runner{tofuPath: tofuPath}
}

// RunInput contient les entrées nécessaires à l'exécution d'un plan refresh-only.
type RunInput struct {
	TFFiles   []client.TFFile
	StateJSON []byte
	EnvCreds  map[string]string
	// Outputs contient les valeurs de sortie du tfstate, utilisées pour
	// renseigner automatiquement les variables déclarées sans valeur par
	// défaut (par correspondance de nom).
	Outputs map[string]interface{}
}

// Run écrit les fichiers .tf et le tfstate dans un répertoire temporaire,
// exécute `tofu init -backend=false` puis `tofu plan -refresh-only`, et
// retourne le plan au format JSON structuré.
func (r *Runner) Run(ctx context.Context, input RunInput) (*tfjson.Plan, error) {
	workDir, err := os.MkdirTemp("", "tofu-fine-drift-*")
	if err != nil {
		return nil, fmt.Errorf("création répertoire temporaire: %w", err)
	}
	defer os.RemoveAll(workDir)

	for _, f := range input.TFFiles {
		dest := filepath.Join(workDir, filepath.FromSlash(f.Path))
		if !strings.HasPrefix(filepath.Clean(dest), filepath.Clean(workDir)) {
			return nil, fmt.Errorf("chemin de fichier invalide: %s", f.Path)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return nil, fmt.Errorf("création répertoire pour %s: %w", f.Path, err)
		}
		content := []byte(f.Content)
		if strings.HasSuffix(f.Path, ".tf") {
			content = stripBackendBlock(content, f.Path)
		}
		if err := os.WriteFile(dest, content, 0o644); err != nil {
			return nil, fmt.Errorf("écriture de %s: %w", f.Path, err)
		}
	}

	statePath := filepath.Join(workDir, "terraform.tfstate")
	if err := os.WriteFile(statePath, input.StateJSON, 0o644); err != nil {
		return nil, fmt.Errorf("écriture du tfstate: %w", err)
	}

	if vars := declaredVariables(input.TFFiles); len(vars) > 0 {
		tfvarsJSON, err := buildTFVarsJSON(vars, input.Outputs)
		if err != nil {
			return nil, fmt.Errorf("génération de terraform.tfvars.json: %w", err)
		}
		if tfvarsJSON != nil {
			tfvarsPath := filepath.Join(workDir, "terraform.tfvars.json")
			if err := os.WriteFile(tfvarsPath, tfvarsJSON, 0o644); err != nil {
				return nil, fmt.Errorf("écriture de terraform.tfvars.json: %w", err)
			}
		}
	}

	pluginCacheDir := filepath.Join(os.TempDir(), "tofu-plugin-cache")
	if err := os.MkdirAll(pluginCacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("création du cache de plugins: %w", err)
	}

	tf, err := tfexec.NewTerraform(workDir, r.tofuPath)
	if err != nil {
		return nil, fmt.Errorf("initialisation tofu-exec: %w", err)
	}
	tf.SetStdout(os.Stderr)
	tf.SetStderr(os.Stderr)
	tf.SetLogger(log.New(os.Stderr, "[tofu] ", 0))

	env := map[string]string{
		"TF_PLUGIN_CACHE_DIR": pluginCacheDir,
		"HOME":                os.TempDir(),
	}
	for k, v := range input.EnvCreds {
		env[k] = v
	}
	if err := tf.SetEnv(env); err != nil {
		return nil, fmt.Errorf("configuration de l'environnement tofu: %w", err)
	}

	if err := tf.Init(ctx, tfexec.Backend(false)); err != nil {
		return nil, fmt.Errorf("tofu init: %w", err)
	}

	planPath := filepath.Join(workDir, "plan.tfplan")
	if _, err := tf.Plan(ctx, tfexec.RefreshOnly(true), tfexec.Out(planPath)); err != nil {
		return nil, fmt.Errorf("tofu plan -refresh-only: %w", err)
	}

	plan, err := tf.ShowPlanFile(ctx, planPath)
	if err != nil {
		return nil, fmt.Errorf("tofu show -json: %w", err)
	}

	return plan, nil
}
