package internal

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8s "k8s.io/client-go/kubernetes"
)

func UpdateResource(clientset *k8s.Clientset, obj interface{}, strategy string, namespace, name string) {
	var err error
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		if strategy == "replace" {
			_, err = clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), o, v1.UpdateOptions{})
		} else if strategy == "patch" {
			_, err = clientset.CoreV1().ConfigMaps(namespace).Patch(context.TODO(), name, types.MergePatchType, []byte("{}"), v1.PatchOptions{})
		} else {
			fmt.Printf("Unknown strategy '%s' for ConfigMap '%s' in namespace '%s'\n", strategy, name, namespace)
			return
		}
		if err != nil {
			fmt.Printf("Failed to update ConfigMap '%s' in namespace '%s' with strategy '%s': %v\n", name, namespace,strategy, err)
		} else {
			fmt.Printf("Updated ConfigMap '%s' in namespace '%s' with strategy '%s' \n", name, namespace,strategy)
		}
	case *corev1.Secret:
		if strategy == "replace" {
			_, err = clientset.CoreV1().Secrets(namespace).Update(context.TODO(), o, v1.UpdateOptions{})
		} else if strategy == "patch" {
			// Example: patch with empty merge (customize as needed)
			_, err = clientset.CoreV1().Secrets(namespace).Patch(context.TODO(), name, types.MergePatchType, []byte("{}"), v1.PatchOptions{})
		} else {
			fmt.Printf("Unknown strategy '%s' for Secret '%s' in namespace '%s'\n", strategy, name, namespace)
			return
		}
		if err != nil {
			fmt.Printf("Failed to update Secret '%s' in namespace '%s' with strategy '%s': %v\n", name, namespace,strategy, err)
		} else {
			fmt.Printf("Updated Secret '%s' in namespace '%s' with strategy '%s'\n", name, namespace,strategy)
		}
	}
}
