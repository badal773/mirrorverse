package internal

import (
	"context"
	"fmt"
	"time"

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
	name := GetName(event.Object)
	switch event.Type {
	case watch.Added:
		fmt.Printf("%s created: %s", resource, name)
		if HasSyncSourceLabel(event.Object) {
			fmt.Printf(" - found the source labels...\n")
			CreateResource(clientset, event.Object)
		}
	case watch.Modified:
		fmt.Printf("%s updated: %s\n", resource, name)
		if HasSyncSourceLabel(event.Object) {
			fmt.Printf(" - found the source labels...\n")
			CreateResource(clientset, event.Object)
		} else if IsMirrorverseReplica(event.Object) && !IsMarkedAsStale(event.Object) && HasSyncSourceRef(event.Object) {
			sourceName, sourceNamespace := GetSyncSourceRef(event.Object)
			strategy := GetStrategy(event.Object)
			sourceObj := GetSyncSourceObject(clientset, sourceName, sourceNamespace)
			if NeedsSync(event.Object, sourceObj) { // <-- Only update if needed
				fmt.Printf(" - found the mirrorverse replica and needs sync...\n")
				UpdateResource(clientset, sourceObj, strategy, GetNamespace(event.Object), GetName(event.Object))
				UpdateLabelsLastSynced(event.Object, clientset)
			} else {
				fmt.Printf(" - found the mirrorverse replica but no sync needed. as no changes detected\n")
			}
		}
	case watch.Deleted:
		fmt.Printf("%s deleted: %s\n", resource, name)
		if HasSyncSourceLabel(event.Object) {
			fmt.Printf(" - found the source labels...\n")
			DeleteResource(clientset, event.Object)
		}
	case watch.Error:
		fmt.Printf("Error event for %s: %v\n", resource, event.Object)
	}
}

