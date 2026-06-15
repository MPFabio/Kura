package k8s

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

// ArgoCDNamespace est le namespace dans lequel ArgoCD est installé.
const ArgoCDNamespace = "argocd"

// argoCDInstallManifestURL pointe vers les manifests officiels d'installation d'ArgoCD.
const argoCDInstallManifestURL = "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"

// maxManifestSize limite la taille du manifest téléchargé (10 Mo) par sécurité.
const maxManifestSize = 10 << 20

// InstallArgoCD installe ArgoCD dans le namespace "argocd" du cluster ciblé par restConfig,
// en appliquant les manifests officiels.
func InstallArgoCD(ctx context.Context, restConfig *rest.Config, clientset *kubernetes.Clientset) error {
	// 1. Créer le namespace argocd s'il n'existe pas.
	_, err := clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ArgoCDNamespace},
	}, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("création du namespace %s: %w", ArgoCDNamespace, err)
	}

	// 2. Télécharger le manifest officiel.
	manifest, err := fetchManifest(ctx)
	if err != nil {
		return fmt.Errorf("téléchargement du manifest ArgoCD: %w", err)
	}

	// 3. Construire le dynamic client et le RESTMapper.
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("création du dynamic client: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("création du discovery client: %w", err)
	}

	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return fmt.Errorf("découverte des ressources API: %w", err)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	// 4. Découper le manifest multi-documents et appliquer chaque ressource.
	objects, err := splitYAMLDocuments(manifest)
	if err != nil {
		return fmt.Errorf("analyse du manifest ArgoCD: %w", err)
	}

	var errs []string
	for _, obj := range objects {
		if obj == nil || len(obj.Object) == 0 {
			continue
		}
		if err := applyObject(ctx, dynamicClient, mapper, obj); err != nil {
			errs = append(errs, fmt.Sprintf("%s/%s (%s): %v", obj.GetNamespace(), obj.GetName(), obj.GetKind(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("certaines ressources n'ont pas pu être appliquées: %s", strings.Join(errs, "; "))
	}

	// Le manifest officiel ne définit ni requests CPU/mémoire ni probes tolérantes
	// pour argocd-repo-server. Sur un nœud sous pression CPU (ex: pool spot),
	// l'initialisation (génération du trousseau GPG, ~30s) dépasse le délai du
	// liveness probe et le pod est tué/redémarré en boucle, faisant échouer la
	// génération des manifests Helm (ComparisonError "connection refused") pour
	// toute Application un peu lourde (ex: kube-prometheus-stack). On corrige
	// ça une fois pour toutes ici plutôt qu'au cas par cas par Application.
	if err := patchRepoServerResilience(ctx, clientset); err != nil {
		return fmt.Errorf("ajustement de argocd-repo-server: %w", err)
	}

	// Le manifest officiel ne pose pas le label app.kubernetes.io/instance=argocd
	// sur les pods (seulement name/component/part-of). Quand ArgoCD s'auto-gère
	// ensuite via le chart Helm argo-cd (qui ajoute ce label aux Services pour
	// cibler ses propres pods), les Deployments/StatefulSet du manifest officiel
	// ne sont alors plus sélectionnés par ces Services (endpoints vides,
	// repo-server inatteignable). On pose ce label dès l'installation pour que
	// la bascule vers l'auto-gestion Helm soit transparente.
	if err := labelOfficialManifestForSelfManagement(ctx, clientset); err != nil {
		return fmt.Errorf("préparation à l'auto-gestion ArgoCD: %w", err)
	}

	return nil
}

// labelOfficialManifestForSelfManagement ajoute app.kubernetes.io/instance=argocd
// au pod template des Deployments/StatefulSet installés par le manifest officiel,
// label requis par les Services générés par le chart Helm argo-cd lors de
// l'auto-gestion ("app of apps").
func labelOfficialManifestForSelfManagement(ctx context.Context, clientset *kubernetes.Clientset) error {
	patch := []byte(`{"spec":{"template":{"metadata":{"labels":{"app.kubernetes.io/instance":"argocd"}}}}}`)

	deployments := []string{
		"argocd-repo-server",
		"argocd-server",
		"argocd-redis",
		"argocd-dex-server",
		"argocd-applicationset-controller",
		"argocd-notifications-controller",
	}
	for _, name := range deployments {
		_, err := clientset.AppsV1().Deployments(ArgoCDNamespace).Patch(
			ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{FieldManager: "modulops"},
		)
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("étiquetage du déploiement %s: %w", name, err)
		}
	}

	_, err := clientset.AppsV1().StatefulSets(ArgoCDNamespace).Patch(
		ctx, "argocd-application-controller", types.StrategicMergePatchType, patch, metav1.PatchOptions{FieldManager: "modulops"},
	)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("étiquetage du statefulset argocd-application-controller: %w", err)
	}

	return nil
}

// patchRepoServerResilience donne à argocd-repo-server des requests CPU/mémoire
// garanties et des probes plus tolérantes au démarrage lent.
func patchRepoServerResilience(ctx context.Context, clientset *kubernetes.Clientset) error {
	patch := []byte(`{
		"spec": {
			"template": {
				"spec": {
					"containers": [{
						"name": "argocd-repo-server",
						"resources": {
							"requests": {"cpu": "100m", "memory": "256Mi"},
							"limits": {"memory": "512Mi"}
						},
						"livenessProbe": {
							"httpGet": {"path": "/healthz?full=true", "port": 8084, "scheme": "HTTP"},
							"initialDelaySeconds": 90,
							"periodSeconds": 30,
							"timeoutSeconds": 10,
							"failureThreshold": 5
						},
						"readinessProbe": {
							"httpGet": {"path": "/healthz", "port": 8084, "scheme": "HTTP"},
							"initialDelaySeconds": 60,
							"periodSeconds": 10,
							"timeoutSeconds": 5,
							"failureThreshold": 5
						}
					}]
				}
			}
		}
	}`)

	_, err := clientset.AppsV1().Deployments(ArgoCDNamespace).Patch(
		ctx, "argocd-repo-server", types.StrategicMergePatchType, patch, metav1.PatchOptions{FieldManager: "modulops"},
	)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

// fetchManifest télécharge le manifest d'installation officiel d'ArgoCD.
func fetchManifest(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, argoCDInstallManifestURL, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("statut HTTP inattendu: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxManifestSize))
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// splitYAMLDocuments découpe un manifest YAML multi-documents en objets unstructured.
func splitYAMLDocuments(manifest string) ([]*unstructured.Unstructured, error) {
	reader := utilyaml.NewYAMLReader(bufio.NewReader(strings.NewReader(manifest)))

	var objects []*unstructured.Unstructured
	for {
		doc, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		doc = []byte(strings.TrimSpace(string(doc)))
		if len(doc) == 0 {
			continue
		}

		jsonDoc, err := utilyaml.ToJSON(doc)
		if err != nil {
			return nil, err
		}
		if string(jsonDoc) == "null" {
			continue
		}

		obj := &unstructured.Unstructured{}
		if err := obj.UnmarshalJSON(jsonDoc); err != nil {
			return nil, err
		}
		if obj.GetKind() == "" {
			continue
		}

		objects = append(objects, obj)
	}

	return objects, nil
}

// applyObject applique une ressource unstructured via server-side apply, avec
// fallback Create -> Update en cas de conflit avec un objet existant non géré par modulops.
func applyObject(ctx context.Context, dynamicClient dynamic.Interface, mapper meta.RESTMapper, obj *unstructured.Unstructured) error {
	gvk := obj.GroupVersionKind()
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("résolution du mapping REST: %w", err)
	}

	var resourceClient dynamic.ResourceInterface
	if mapping.Scope.Name() == "namespace" {
		namespace := obj.GetNamespace()
		if namespace == "" {
			namespace = ArgoCDNamespace
			obj.SetNamespace(namespace)
		}
		resourceClient = dynamicClient.Resource(mapping.Resource).Namespace(namespace)
	} else {
		resourceClient = dynamicClient.Resource(mapping.Resource)
	}

	data, err := obj.MarshalJSON()
	if err != nil {
		return fmt.Errorf("sérialisation de la ressource: %w", err)
	}

	_, err = resourceClient.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "modulops",
		Force:        boolPtr(true),
	})
	if err == nil {
		return nil
	}

	// Fallback : certains clusters/CRDs ne supportent pas le server-side apply -> create puis update.
	if _, createErr := resourceClient.Create(ctx, obj, metav1.CreateOptions{FieldManager: "modulops"}); createErr == nil {
		return nil
	} else if !apierrors.IsAlreadyExists(createErr) {
		return fmt.Errorf("apply échoué (%v) et create échoué: %w", err, createErr)
	}

	existing, getErr := resourceClient.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if getErr != nil {
		return fmt.Errorf("apply échoué (%v) et lecture de la ressource existante échouée: %w", err, getErr)
	}
	obj.SetResourceVersion(existing.GetResourceVersion())
	if _, updateErr := resourceClient.Update(ctx, obj, metav1.UpdateOptions{FieldManager: "modulops"}); updateErr != nil {
		return fmt.Errorf("apply échoué (%v) et update échoué: %w", err, updateErr)
	}

	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
