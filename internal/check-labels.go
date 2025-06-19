// helpers to parse labels
package internal

import (
	corev1 "k8s.io/api/core/v1"
)

// Returns true if the object has the sync-source label set to "true"
func HasSyncSourceLabel(obj interface{}) bool {
	labels := GetLabels(obj)
	if labels == nil {
		return false
	}
	return labels["mirrorverse.dev/sync-source"] == "true"
}

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
