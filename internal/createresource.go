// logic to create/update/patch
package internal

import (
	"context"
	"fmt"
	"strings"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

// GetTargetNamespaces returns the final list of namespaces to apply, giving priority to excludeNamespaces
func GetTargetNamespaces(targets, exclude string) []string {
	targetNamespaces := []string{}
	if targets != "" {
		targetNamespaces = strings.Split(targets, "_")
	}
	excludeNamespaces := map[string]bool{}
	if exclude != "" {
		for _, ns := range strings.Split(exclude, "_") {
			excludeNamespaces[strings.TrimSpace(ns)] = true
		}
	}
	finalNamespaces := []string{}
	for _, ns := range targetNamespaces {
		ns = strings.TrimSpace(ns)
		if ns == "" || excludeNamespaces[ns] {
			continue // skip if excluded or empty
		}
		finalNamespaces = append(finalNamespaces, ns)
	}
	return finalNamespaces
}

func CreateResource(clientset *k8s.Clientset, obj interface{}) {
	labels := GetLabels(obj)
	var name, namespace string
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

	finalLabels, targets, exclude, strategy := PrepareLabels(labels, namespace, name)
	UpdateResourceMeta(obj, finalLabels)

	// Get final target namespaces (exclude takes priority)
	finalNamespaces := GetTargetNamespaces(targets, exclude)

	// Create in each target namespace
	for _, targetNS := range finalNamespaces {
		// Set the target namespace for the object
		var objtype string
		switch o := obj.(type) {
		case *corev1.ConfigMap:
			o.Namespace = targetNS
			objtype = "configMap"
		case *corev1.Secret:
			o.Namespace = targetNS
			objtype = "Secret"
		}
		// Try to create, update if already exists
		fmt.Printf("creating %s '%s' in namespace '%s'...\n", objtype, name, targetNS)
		createOrUpdateResource(clientset, obj, strategy, targetNS, name)
	}
}

// createOrUpdateResource tries to create, and updates if already exists
func createOrUpdateResource(clientset *k8s.Clientset, obj interface{}, strategy, namespace, name string) error {
	var err error
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		_, err = clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), o, v1.CreateOptions{})
		if err != nil && apierrors.IsAlreadyExists(err) {
			fmt.Printf("ConfigMap '%s' already exists in namespace '%s', updating.\n", name, namespace)
			UpdateResource(clientset, o, strategy, namespace, name)
			return nil // Return early after update
		}
	case *corev1.Secret:
		_, err = clientset.CoreV1().Secrets(namespace).Create(context.TODO(), o, v1.CreateOptions{})
		if err != nil && apierrors.IsAlreadyExists(err) {
			fmt.Printf("Secret '%s' already exists in namespace '%s', updating.\n", name, namespace)
			UpdateResource(clientset, o, strategy, namespace, name)
			return nil // Return early after update

		}
	}
	if err != nil {
		fmt.Printf("Failed to create resource '%s' in namespace '%s': %v\n", name, namespace, err)
	} else {
		fmt.Printf("Created resource '%s' in namespace '%s'\n", name, namespace)
	}
	return err
}
