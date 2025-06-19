package internal

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// CreateWatcher starts watchers for ConfigMaps and Secrets in all namespaces.
func CreateWatcher(clientset *kubernetes.Clientset) {
	fmt.Println("Watching ConfigMaps and Secrets...")
	stopCh := make(chan struct{})
	go watchResource(clientset, "configmaps", stopCh)
	go watchResource(clientset, "secrets", stopCh)
	select {} // Keep running
}

// watchResource watches a specific resource type and handles events.
func watchResource(clientset *kubernetes.Clientset, resource string, stopCh <-chan struct{}) {
	watcher, err := getWatcher(clientset, resource)
	if err != nil {
		fmt.Printf("Error creating watcher for %s: %v\n", resource, err)
		return
	}
	defer watcher.Stop()

	fmt.Printf("Started watching %s...\n", resource)
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				time.Sleep(2 * time.Second)
				go watchResource(clientset, resource, stopCh)
				return
			}
			handleEvent(event, resource, clientset)
		case <-stopCh:
			return
		}
	}
}

// getWatcher returns a watcher for the given resource type.
func getWatcher(clientset *kubernetes.Clientset, resource string) (watch.Interface, error) {
	switch resource {
	case "configmaps":
		return clientset.CoreV1().ConfigMaps("").Watch(context.TODO(), metav1.ListOptions{})
	case "secrets":
		return clientset.CoreV1().Secrets("").Watch(context.TODO(), metav1.ListOptions{})
	default:
		return nil, fmt.Errorf("unknown resource: %s", resource)
	}
}

// handleEvent processes a watch event for a resource.
func handleEvent(event watch.Event, resource string, clientset *kubernetes.Clientset) {
	name := getName(event.Object)
	switch event.Type {
	case watch.Added:
		fmt.Printf("%s created: %s\n", resource, name)
		if HasSyncSourceLabel(event.Object) {
			CreateResource(clientset, event.Object)
		}
	case watch.Modified:
		fmt.Printf("%s updated: %s\n", resource, name)
		if HasSyncSourceLabel(event.Object) {
			CreateResource(clientset, event.Object)
		}
	case watch.Deleted:
		fmt.Printf("%s deleted: %s\n", resource, name)
		// Optionally handle deletion logic here
	case watch.Error:
		fmt.Printf("Error event for %s: %v\n", resource, event.Object)
	}
}

// getName returns the name of a ConfigMap or Secret.
func getName(obj interface{}) string {
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		return o.Name
	case *corev1.Secret:
		return o.Name
	default:
		return "unknown"
	}
}
