package k8s

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// NewDynamicClient crée un client dynamique permettant de manipuler des ressources
// arbitraires (CRDs incluses) à partir d'une configuration REST.
func NewDynamicClient(restConfig *rest.Config) (dynamic.Interface, error) {
	return dynamic.NewForConfig(restConfig)
}
