package main

import (
	"fmt"
	"k8s-syncer/client"
	"k8s-syncer/internal"

)	

func main() {
	fmt.Println("Starting the k8s-syncer controller...")

	k8sClient := client.GetKubeClient()

	// set up the watcher
	internal.CreateWatcher(k8sClient)


}
