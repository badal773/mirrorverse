package internal

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DeleteResource(clientset *kubernetes.Clientset, obj interface{}) {
	// Implement the logic to delete the resource using the clientset
	labels := GetLabels(obj)
	var targets, exclude string

	//check if cleanup is needed
	// Implement the cleanup logic here
	for key, value := range labels {
		if key == "mirrorverse.dev/targets" {
			targets = value
		}
		if key == "mirrorverse.dev/exclude" {
			exclude = value
		}
	}
	finalNamespaces := GetTargetNamespaces(targets, exclude)
	objectName := GetName(obj)
	if labels["mirrorverse.dev/cleanup"] == "true" {
		// If cleanup is true, delete the resource from all target namespaces
		if len(finalNamespaces) == 0 {
			fmt.Println("No target namespaces specified for deletion")
			return
		}
		for _, namespace := range finalNamespaces {
			switch obj.(type) {
			case *corev1.ConfigMap:
				err := clientset.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), objectName, metav1.DeleteOptions{})
				if err != nil {
					fmt.Printf("Error deleting ConfigMap %s in namespace %s: %v\n", objectName, namespace, err)
				} else {
					fmt.Printf("Deleted ConfigMap %s in namespace %s\n", objectName, namespace)
				}
			case *corev1.Secret:
				err := clientset.CoreV1().Secrets(namespace).Delete(context.TODO(), objectName, metav1.DeleteOptions{})
				if err != nil {
					fmt.Printf("Error deleting Secret %s in namespace %s: %v\n", objectName, namespace, err)
				} else {
					fmt.Printf("Deleted Secret %s in namespace %s\n", objectName, namespace)
				}
			default:
				fmt.Println("Unsupported resource type for deletion")
			}
		}
	} else {
		// If cleanup is false, just print the message
		fmt.Printf("Cleanup is false for %T %s, skipping deletion\n", obj, objectName)
		// Add mirrorverse.dev/stale label to all target objects
		if len(finalNamespaces) == 0 {
			fmt.Println("No target namespaces specified for marking as stale")
			return
		}
		for _, namespace := range finalNamespaces {
			// Prepare labels for each target
			staleLabels := make(map[string]string)
			for k, v := range labels {
				staleLabels[k] = v
			}
			staleLabels["mirrorverse.dev/stale"] = "true"
			UpdateLabels(GetSyncSourceObject(clientset,GetName(obj),namespace),clientset, staleLabels)
			fmt.Printf("Marked %T %s in namespace %s as stale\n", obj, objectName, namespace)
		}
	}
}
