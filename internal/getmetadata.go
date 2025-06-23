package internal

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Returns the labels of a ConfigMap or Secret, or nil otherwise
func GetLabels(obj interface{}) map[string]string {
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		return o.Labels
	case *corev1.Secret:
		return o.Labels
	default:
		return nil
	}
}

// Returns the name of a ConfigMap or Secret, or "unknown" if not found
func GetName(obj interface{}) string {
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		return o.Name
	case *corev1.Secret:
		return o.Name
	default:
		return "unknown"
	}
}

// Returns the namespace of a ConfigMap or Secret, or "unknown" if not found
func GetNamespace(obj interface{}) string {
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		return o.Namespace
	case *corev1.Secret:
		return o.Namespace
	default:
		return "unknown"
	}
}
// Returns the sync source reference label from a ConfigMap or Secret
func GetSyncSourceRef(obj interface{}) (name string, namespace string) {
	labels := GetLabels(obj)
	if labels == nil {
		return "", ""
	}

	ref, ok := labels["mirrorverse.dev/sync-source-ref"]
	if !ok || ref == "" {
		return "", ""
	}

	parts := strings.SplitN(ref, ".", 2)
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}

//get strategy from labels
func GetStrategy(obj interface{}) string {
	labels := GetLabels(obj)
	if labels == nil {
		return ""
	}
	return labels["mirrorverse.dev/strategy"]
}

// get object
func GetSyncSourceObject(clientset *kubernetes.Clientset, name string, namespace string) (obj interface{}) {
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return configMap
	}
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return secret

	}
	return nil
}

// Returns if the object is stale
func IsMarkedAsStale(obj interface{}) bool {
	labels := GetLabels(obj)
	if labels == nil {
		return false
	}
	return labels["mirrorverse.dev/stale"] == "true"
}