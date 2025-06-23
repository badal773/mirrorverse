// helpers to parse labels
package internal

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Returns true if the object has the sync-source label set to "true"
func HasSyncSourceLabel(obj interface{}) bool {
	labels := GetLabels(obj)
	if labels == nil {
		return false
	}
	return labels["mirrorverse.dev/sync-source"] == "true"
}

func IsMirrorverseReplica(obj interface{}) bool {
	labels := GetLabels(obj)
	if labels == nil {
		return false
	}
	return labels["mirrorverse.dev/sync-replica"] == "true"
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
		"mirrorverse.dev/sync-source-ref": fmt.Sprintf("%s.%s", name, namespace),
		"mirrorverse.dev/last-synced":     timeStr,
	}
	// Add strategy label if it is empty
	if strategy != "" {
		managedLabels["mirrorverse.dev/strategy"] = "patch"
	}
	for k, v := range managedLabels {
		cleanLabels[k] = v
	}
	
	return cleanLabels, targets, exclude, strategy
}


// Helper to update the last-synced label
func UpdateLabelsLastSynced(obj interface{}, clientset *kubernetes.Clientset) {
	labels := GetLabels(obj)
	if labels == nil {
		return
	}
	timeStr := time.Now().Format("2006-01-02T15-04-05Z07.00")
	timeStr = strings.ReplaceAll(timeStr, "+", "Z")

	labels["mirrorverse.dev/last-synced"] = timeStr
	UpdateLabels(obj, clientset, labels)
}

// update labels on the object
func UpdateLabels(obj interface{}, clientset *kubernetes.Clientset, labels map[string]string) {
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		o.Labels = labels
		_, err := clientset.CoreV1().ConfigMaps(GetNamespace(obj)).Update(context.TODO(), obj.(*corev1.ConfigMap), v1.UpdateOptions{})
		if err != nil {
			fmt.Printf("Failed to update last-synced label: %v\n", err)
		} else {
			fmt.Printf("Updated last-synced label for %s/%s\n", GetNamespace(obj), GetName(obj))
		}
	case *corev1.Secret:
		o.Labels = labels
		_, err := clientset.CoreV1().Secrets(GetNamespace(obj)).Update(context.TODO(), obj.(*corev1.Secret), v1.UpdateOptions{})
		if err != nil {
			fmt.Printf("Failed to update last-synced label: %v\n", err)
		} else {
			fmt.Printf("Updated last-synced label for %s/%s\n", GetNamespace(obj), GetName(obj))
		}
	}

}