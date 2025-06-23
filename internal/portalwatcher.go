package internal

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// =====================
// Mirrorverse Watcher: Watches for changes to ConfigMaps and Secrets in all namespaces.
// This is the "event loop" that powers the whole controller.
//
// If you're new to Kubernetes controllers, see:
//   - https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes
//   - https://pkg.go.dev/k8s.io/client-go/tools/cache#NewSharedInformer
//   - https://book.kubebuilder.io/cronjob-tutorial/controller-implementation.html
// =====================

// CreateWatcher is the entry point for starting the Mirrorverse watcher system.
//
// What it does:
//   - Creates a channel (stopCh) that could be used to signal the watchers to stop (not used here, but good practice for future extensibility).
//   - Starts two goroutines (background threads in Go):
//     1. One to watch all ConfigMaps in all namespaces
//     2. One to watch all Secrets in all namespaces
//   - Each watcher runs independently and will react to add/update/delete events.
//   - The final 'select {}' line blocks forever, so the program doesn't exit and the watchers keep running.
//
// Why goroutines? In Go, goroutines are lightweight threads. This lets us watch both resource types in parallel without blocking each other.
//
// For more on goroutines: https://gobyexample.com/goroutines
// For more on channels:   https://gobyexample.com/channels
func CreateWatcher(clientset *kubernetes.Clientset) {
	fmt.Println("Watching ConfigMaps and Secrets...")
	stopCh := make(chan struct{}) // Channel to signal stopping (not used here, but good practice)
	// Start a watcher goroutine for each resource type
	go watchResource(clientset, "configmaps", stopCh)
	go watchResource(clientset, "secrets", stopCh)
	select {} // Block forever so the watchers keep running
}

// watchResource watches a specific resource type (ConfigMap or Secret) and handles events.
// It will automatically restart itself if the watch channel closes (e.g., due to network issues).
//
// For beginners: This is a loop that keeps listening for changes (add, update, delete)
// to a specific resource type. When something happens, it calls handleEvent to react.
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
				// If the channel is closed, restart the watcher after a short delay
				time.Sleep(2 * time.Second)
				go watchResource(clientset, resource, stopCh)
				return
			}
			handleEvent(event, resource, clientset)
		case <-stopCh:
			// If stopCh is closed, exit the watcher
			return
		}
	}
}

// getWatcher returns a Kubernetes watcher for the given resource type (ConfigMap or Secret).
// It watches all namespaces by passing an empty string as the namespace.
//
// For more on how "watch" works in Kubernetes:
//
//	https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes
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

// handleEvent processes a watch event for a resource (ConfigMap or Secret).
// It determines what kind of event happened (Added, Modified, Deleted, Error)
// and triggers the appropriate Mirrorverse sync logic.
//
// For beginners: This is the "brain" that decides what to do when something changes.
// If a new source is created, it triggers sync. If a replica is updated, it checks if it
// needs to be re-synced. If a source is deleted, it cleans up replicas.
func handleEvent(event watch.Event, resource string, clientset *kubernetes.Clientset) {
	name := GetName(event.Object)
	switch event.Type {
	case watch.Added:
		fmt.Printf("%s created: %s", resource, name)
		// If this is a source resource (has the sync label), trigger sync logic
		if HasSyncSourceLabel(event.Object) {
			fmt.Printf(" - found the source labels...\n")
			CreateResource(clientset, event.Object)
		}
	case watch.Modified:
		fmt.Printf("%s updated: %s\n", resource, name)
		if HasSyncSourceLabel(event.Object) {
			// If the source was updated, trigger sync logic
			fmt.Printf(" - found the source labels...\n")
			CreateResource(clientset, event.Object)
		} else if IsMirrorverseReplica(event.Object) && !IsMarkedAsStale(event.Object) && HasSyncSourceRef(event.Object) {
			// If a managed replica was updated, check if it needs to be re-synced
			sourceName, sourceNamespace := GetSyncSourceRef(event.Object)
			strategy := GetStrategy(event.Object)
			sourceObj := GetSyncSourceObject(clientset, sourceName, sourceNamespace)
			if NeedsSync(event.Object, sourceObj) { // Only update if needed
				fmt.Printf(" - found the mirrorverse replica and needs sync...\n")
				UpdateResource(clientset, sourceObj, strategy, GetNamespace(event.Object), GetName(event.Object))
				UpdateLabelsLastSynced(event.Object, clientset)
			} else {
				fmt.Printf(" - found the mirrorverse replica but no sync needed. as no changes detected\n")
			}
		}
	case watch.Deleted:
		fmt.Printf("%s deleted: %s\n", resource, name)
		// If a source is deleted, trigger cleanup logic
		if HasSyncSourceLabel(event.Object) {
			fmt.Printf(" - found the source labels...\n")
			DeleteResource(clientset, event.Object)
		}
	case watch.Error:
		fmt.Printf("Error event for %s: %v\n", resource, event.Object)
	}
}
