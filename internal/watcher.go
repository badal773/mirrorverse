package internal

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	k8s "k8s.io/client-go/kubernetes"
)

func CreateWatcher(k8sClient *k8s.Clientset) {
	fmt.Println("Creating a watcher for ConfigMaps and Secrets...")

	stopCh := make(chan struct{})
	go watchResource(k8sClient, "configmaps", stopCh)
	go watchResource(k8sClient, "secrets", stopCh)

	// Keep the main goroutine alive
	select {}          
}

func watchResource(clientset *k8s.Clientset, resource string, stopCh <-chan struct{}) {
	var watcher watch.Interface
	var err error

	switch resource {
	case "configmaps":
		watcher, err = clientset.CoreV1().ConfigMaps("").Watch(context.TODO(), metav1.ListOptions{})
	case "secrets":
		watcher, err = clientset.CoreV1().Secrets("").Watch(context.TODO(), metav1.ListOptions{})
	default:
		fmt.Printf("Unknown resource: %s\n", resource)
		return
	}

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
				fmt.Printf("%s watcher channel closed, restarting...\n", resource)
				time.Sleep(2 * time.Second)
				go watchResource(clientset, resource, stopCh)
				return
			}
			switch event.Type {
			case watch.Added:
				fmt.Printf("%s created: %s\n", resource, getName(event.Object))
				if CheckIfLabelsPresent(resource, getName(event.Object), event.Object) {
					CreateResource(clientset,event.Object)
				}

			case watch.Modified:
				fmt.Printf("%s updated: %s\n", resource, getName(event.Object))
			case watch.Deleted:
				fmt.Printf("%s deleted: %s\n", resource, getName(event.Object))
			case watch.Error:
				fmt.Printf("Error event for %s: %v\n", resource, event.Object)
			}
		case <-stopCh:
			fmt.Printf("Stopping watcher for %s\n", resource)
			return
		}
	}
}

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
