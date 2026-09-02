// Package fine implémente la détection de drift "fine" : exécution d'un
// `tofu plan -refresh-only` dans un sandbox temporaire à partir des fichiers
// .tf source et du tfstate, pour obtenir un diff générique multi-cloud.
package fine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"

	"github.com/modulops/terraform-service/internal/client"
)

const defaultTofuPath = "/usr/local/bin/tofu"

// runMu sérialise les exécutions de plans au sein du service.
//
// Un plan de rafraîchissement lance un processus de fournisseur par ressource
// et consomme plusieurs centaines de mégaoctets : en laisser tourner deux de
// front sur une machine modeste conduit le noyau à en tuer un (constaté en
// production, « signal: killed » sur un OOM de cgroup).
//
// Le verrou est au niveau du paquet, et non du Runner, car un Runner est créé
// à chaque détection : un verrou d'instance ne protégerait rien.
var runMu sync.Mutex

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
	// TFFiles contient les fichiers .tf (et éventuels fichiers annexes
	// référencés via file("${path.module}/...")), avec des chemins relatifs
	// à la racine du dépôt (ex: "terraform/main.tf", "kura/cloud-init.yaml").
	TFFiles []client.TFFile
	// ModuleDir est le chemin (relatif à la racine du dépôt) du répertoire
	// contenant la configuration .tf principale (ex: "terraform"). C'est dans
	// ce répertoire que `tofu` est exécuté, afin que les références
	// "${path.module}/.." résolvent correctement vers les autres répertoires
	// du dépôt.
	ModuleDir string
	StateJSON []byte
	EnvCreds  map[string]string
	// Outputs contient les valeurs de sortie du tfstate, utilisées pour
	// renseigner automatiquement les variables déclarées sans valeur par
	// défaut (par correspondance de nom).
	Outputs map[string]interface{}
}

// Run écrit les fichiers .tf (et fichiers annexes) ainsi que le tfstate dans
// un répertoire temporaire en conservant leur structure relative au dépôt,
// exécute `tofu init -backend=false` puis `tofu plan -refresh-only` depuis le
// répertoire du module, et retourne le plan au format JSON structuré.
func (r *Runner) Run(ctx context.Context, input RunInput) (*tfjson.Plan, error) {
	runMu.Lock()
	defer runMu.Unlock()

	// L'attente du verrou a pu consommer tout le délai imparti : inutile de
	// démarrer un plan condamné à être interrompu.
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("attente d'un plan concurrent: %w", err)
	}

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
		if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
			return nil, fmt.Errorf("création répertoire pour %s: %w", f.Path, err)
		}
		content := []byte(f.Content)
		if strings.HasSuffix(f.Path, ".tf") {
			content = stripBackendBlock(content, f.Path)
		}
		if err := os.WriteFile(dest, content, 0o600); err != nil {
			return nil, fmt.Errorf("écriture de %s: %w", f.Path, err)
		}
	}

	moduleDir := workDir
	if input.ModuleDir != "" {
		moduleDir = filepath.Join(workDir, filepath.FromSlash(input.ModuleDir))
		if !strings.HasPrefix(filepath.Clean(moduleDir), filepath.Clean(workDir)) {
			return nil, fmt.Errorf("chemin de module invalide: %s", input.ModuleDir)
		}
		if err := os.MkdirAll(moduleDir, 0o700); err != nil {
			return nil, fmt.Errorf("création du répertoire de module: %w", err)
		}
	}

	statePath := filepath.Join(moduleDir, "terraform.tfstate")
	if err := os.WriteFile(statePath, input.StateJSON, 0o600); err != nil {
		return nil, fmt.Errorf("écriture du tfstate: %w", err)
	}

	moduleFiles := filesInDir(input.TFFiles, input.ModuleDir)
	if vars := declaredVariables(moduleFiles); len(vars) > 0 {
		tfvarsJSON, err := buildTFVarsJSON(vars, input.Outputs, gcpProjectID(input.EnvCreds))
		if err != nil {
			return nil, fmt.Errorf("génération de terraform.tfvars.json: %w", err)
		}
		if tfvarsJSON != nil {
			tfvarsPath := filepath.Join(moduleDir, "terraform.tfvars.json")
			if err := os.WriteFile(tfvarsPath, tfvarsJSON, 0o600); err != nil {
				return nil, fmt.Errorf("écriture de terraform.tfvars.json: %w", err)
			}
		}
	}

	// Cache de plugins propre à cette exécution, amorcé par copie du cache
	// partagé.
	//
	// Un cache unique partagé produisait des « text file busy » : après un
	// plan, les processus de provider lancés par tofu survivent brièvement et
	// tiennent leur binaire ouvert. L'exécution suivante voulait réécrire ce
	// même fichier, et Linux refuse d'ouvrir en écriture un exécutable en cours
	// d'exécution. Sérialiser les plans n'y suffit pas : ces processus se
	// terminent après le retour du plan, donc après la libération du verrou.
	//
	// Chaque exécution travaille donc sur sa propre copie, détruite avec le
	// répertoire de travail. Le cache partagé n'est plus écrit qu'une seule
	// fois, lors du tout premier amorçage, ce qui évite de retélécharger le
	// fournisseur à chaque détection.
	pluginCacheDir := filepath.Join(workDir, "plugin-cache")
	if err := os.MkdirAll(pluginCacheDir, 0o700); err != nil {
		return nil, fmt.Errorf("création du cache de plugins: %w", err)
	}
	sharedCacheDir := filepath.Join(os.TempDir(), "tofu-plugin-cache")
	seeded, err := copyTree(sharedCacheDir, pluginCacheDir)
	if err != nil {
		return nil, fmt.Errorf("amorçage du cache de plugins: %w", err)
	}

	tf, err := tfexec.NewTerraform(moduleDir, r.tofuPath)
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

	// Premier amorçage : le cache partagé était vide, on l'alimente avec ce que
	// cet init vient de télécharger. Les exécutions suivantes repartiront de
	// cette copie au lieu d'interroger le registre.
	if !seeded {
		if _, err := copyTree(pluginCacheDir, sharedCacheDir); err != nil {
			// Sans cache partagé, les prochains plans retéléchargeront le
			// fournisseur : c'est plus lent, mais ce n'est pas un échec.
			log.Printf("[tofu] amorçage du cache partagé impossible: %v", err)
		}
	}

	planPath := filepath.Join(moduleDir, "plan.tfplan")
	if _, err := tf.Plan(ctx, tfexec.RefreshOnly(true), tfexec.Out(planPath)); err != nil {
		return nil, fmt.Errorf("tofu plan -refresh-only: %w", err)
	}

	plan, err := tf.ShowPlanFile(ctx, planPath)
	if err != nil {
		return nil, fmt.Errorf("tofu show -json: %w", err)
	}

	return plan, nil
}

// gcpProjectID extrait le project_id du JSON de credentials GCP (clé
// GOOGLE_CREDENTIALS), utilisé pour renseigner automatiquement les variables
// de type "project" (ex: gcp_project) sans valeur par défaut.
func gcpProjectID(envCreds map[string]string) string {
	raw, ok := envCreds["GOOGLE_CREDENTIALS"]
	if !ok || raw == "" {
		return ""
	}
	var creds struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return ""
	}
	return creds.ProjectID
}

// filesInDir filtre les fichiers dont le chemin (relatif à la racine du dépôt)
// se trouve directement dans dir (le répertoire du module), en retournant leur
// chemin relatif à dir. Permet de n'analyser que les .tf de la configuration
// principale (et pas les fichiers annexes référencés ailleurs dans le dépôt).
func filesInDir(files []client.TFFile, dir string) []client.TFFile {
	prefix := ""
	if dir != "" {
		prefix = dir + "/"
	}
	var result []client.TFFile
	for _, f := range files {
		rel := strings.TrimPrefix(f.Path, prefix)
		if rel == f.Path && prefix != "" {
			continue // pas dans dir
		}
		if strings.Contains(rel, "/") {
			continue // sous-répertoire de dir, pas le module lui-même
		}
		result = append(result, client.TFFile{Path: rel, Content: f.Content})
	}
	return result
}

// copyTree recopie le contenu de src dans dst en préservant les permissions.
//
// Le bit exécutable des binaires de fournisseur doit impérativement être
// conservé, faute de quoi tofu les retéléchargerait — ou échouerait à les
// lancer.
//
// Retourne false si src n'existe pas : ce n'est pas une erreur, c'est le cas
// du tout premier amorçage.
func copyTree(src, dst string) (bool, error) {
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !info.IsDir() {
		return false, fmt.Errorf("%s n'est pas un répertoire", src)
	}

	if err := os.MkdirAll(dst, 0o700); err != nil {
		return false, err
	}

	// Accès scopés à src et dst : toute opération ci-dessous porte sur un
	// chemin relatif (rel) résolu par le Root lui-même, qui refuse de sortir
	// du répertoire même via un lien symbolique substitué entre le parcours
	// et l'ouverture (TOCTOU) — gosec G122 sur un os.Open/os.OpenFile classique
	// dans un callback WalkDir.
	srcRoot, err := os.OpenRoot(src)
	if err != nil {
		return false, err
	}
	defer srcRoot.Close()

	dstRoot, err := os.OpenRoot(dst)
	if err != nil {
		return false, err
	}
	defer dstRoot.Close()

	empty := true
	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		if d.IsDir() {
			return dstRoot.Mkdir(rel, 0o700)
		}
		// Les liens symboliques du cache ne sont pas suivis : tofu n'en crée
		// pas, et les recopier à l'aveugle ferait sortir du répertoire.
		if !d.Type().IsRegular() {
			return nil
		}

		entryInfo, err := d.Info()
		if err != nil {
			return err
		}
		in, err := srcRoot.Open(rel)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := dstRoot.OpenFile(rel, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, entryInfo.Mode().Perm())
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err := io.Copy(out, in); err != nil {
			return err
		}
		empty = false
		return out.Close()
	})
	if err != nil {
		return false, err
	}
	return !empty, nil
}
