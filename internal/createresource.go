// logic to create/update/patch
package internal

import (
	"fmt"
	"strings"
	"time"

	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func CreateResource(clientset *k8s.Clientset, obj interface{}) {
	labels := GetLabels(obj)
	var name, namespace, strategy, targets, exclude string

	// Extract namespace from manifest
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		namespace = o.Namespace
		name = o.Name
	case *corev1.Secret:
		namespace = o.Namespace
		name = o.Name
	default:
		fmt.Println("Unsupported resource type")
		return
	}

	// Parse targets, exclude, and strategy
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

	// Remove all labels starting with mirrorverse.dev/
	cleanLabels := make(map[string]string)
	for k, v := range labels {
		if !strings.HasPrefix(k, "mirrorverse.dev/") {
			cleanLabels[k] = v
		}
	}

	// Add managed-by-me labels with valid values
	managedLabels := map[string]string{
		"mirrorverse.dev/sync-replica":    "true",
		"mirrorverse.dev/sync-source-ref": fmt.Sprintf("%s-%s", namespace, name), // use - instead of /
		"mirrorverse.dev/last-synced":     time.Now().Format("2006-01-02T15:04:05Z07:00") ,  // compact, valid label value
	}
	if strategy != "" {
		managedLabels["mirrorverse.dev/strategy"] = strategy
	}
	for k, v := range managedLabels {
		cleanLabels[k] = v
	}

	// Remove namespace from manifest and set new labels
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		o.Namespace = ""
		o.Labels = cleanLabels
		// Remove unwanted metadata fields
		o.UID = ""
		o.ResourceVersion = ""
		o.CreationTimestamp = v1.Time{}
		o.ManagedFields = nil
		if o.Annotations != nil {
			delete(o.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
		}
	case *corev1.Secret:
		o.Namespace = ""
		o.Labels = cleanLabels
		// Remove unwanted metadata fields
		o.UID = ""
		o.ResourceVersion = ""
		o.CreationTimestamp = v1.Time{}
		o.ManagedFields = nil
		if o.Annotations != nil {
			delete(o.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
		}
	}

	// Prepare target and exclude namespaces
	targetNamespaces := []string{}
	if targets != "" {
		targetNamespaces = strings.Split(targets, "-")
	}
	excludeNamespaces := map[string]bool{}
	if exclude != "" {
		for _, ns := range strings.Split(exclude, "-") {
			excludeNamespaces[strings.TrimSpace(ns)] = true
		}
	}

	// Create in each target namespace except those in exclude
	for _, targetNS := range targetNamespaces {
		targetNS = strings.TrimSpace(targetNS)
		if targetNS == "" || excludeNamespaces[targetNS] {
			continue
		}
		fmt.Printf("Creating %T '%s' in namespace '%s'...\n", obj, name, targetNS)
		// Print the manifest as YAML before applying
		var yamlBytes []byte
		var err error
		switch o := obj.(type) {
		case *corev1.ConfigMap:
			o.Namespace = targetNS
			yamlBytes, err = yaml.Marshal(o)
			if err == nil {
				fmt.Printf("---\n%s\n---\n", string(yamlBytes))
			} else {
				fmt.Printf("Failed to marshal ConfigMap to YAML: %v\n", err)
			}
			_, err = clientset.CoreV1().ConfigMaps(targetNS).Create(context.TODO(), o, v1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create ConfigMap '%s' in namespace '%s': %v\n", name, targetNS, err)
			} else {
				fmt.Printf("Created ConfigMap '%s' in namespace '%s'\n", name, targetNS)
			}
		case *corev1.Secret:
			o.Namespace = targetNS
			yamlBytes, err = yaml.Marshal(o)
			if err == nil {
				fmt.Printf("---\n%s\n---\n", string(yamlBytes))
			} else {
				fmt.Printf("Failed to marshal Secret to YAML: %v\n", err)
			}
			_, err = clientset.CoreV1().Secrets(targetNS).Create(context.TODO(), o, v1.CreateOptions{})
			if err != nil {
				fmt.Printf("Failed to create Secret '%s' in namespace '%s': %v\n", name, targetNS, err)
			} else {
				fmt.Printf("Created Secret '%s' in namespace '%s'\n", name, targetNS)
			}
		}
	}
}
