// helpers to parse labels
package internal

import (
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// Helper to clean up metadata and set labels
func UpdateResourceMeta(obj interface{}, labels map[string]string) {
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		o.Namespace = ""
		o.Labels = labels
		o.UID = ""
		o.ResourceVersion = ""
		o.CreationTimestamp = v1.Time{}
		o.ManagedFields = nil
		if o.Annotations != nil {
			delete(o.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
		}
	case *corev1.Secret:
		o.Namespace = ""
		o.Labels = labels
		o.UID = ""
		o.ResourceVersion = ""
		o.CreationTimestamp = v1.Time{}
		o.ManagedFields = nil
		if o.Annotations != nil {
			delete(o.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
		}
	}
}

// Helper to clean up and add managed labels, and parse targets/exclude/strategy
func PrepareLabels(labels map[string]string, namespace, name string) (finalLabels map[string]string, targets, exclude, strategy string) {
	cleanLabels := make(map[string]string)
	for k, v := range labels {
		if !strings.HasPrefix(k, "mirrorverse.dev/") {
			cleanLabels[k] = v
		}
	}
	for key, value := range labels {
		if key == "mirrorverse.dev/targets" {
			targets = value
		}
		if key == "mirrorverse.dev/exclude" {
			exclude = value
		}
		if key == "mirrorverse.dev/strategy" {
			strategy = value
		}
	}
	timeStr := time.Now().Format("2006-01-02T15-04-05Z07.00")
	timeStr = strings.ReplaceAll(timeStr, "+", "Z")
	managedLabels := map[string]string{
		"mirrorverse.dev/sync-replica":    "true",
		"mirrorverse.dev/sync-source-ref": fmt.Sprintf("%s.%s", namespace, name),
		"mirrorverse.dev/last-synced":     timeStr,
	}
	if strategy != "" {
		managedLabels["mirrorverse.dev/strategy"] = strategy
	}
	for k, v := range managedLabels {
		cleanLabels[k] = v
	}
	
	return cleanLabels, targets, exclude, strategy
}
