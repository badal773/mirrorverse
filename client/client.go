package client

import (
	"fmt"
	"os"

	k8s "k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
)

func GetKubeClient() *k8s.Clientset { // Capital G to export the function
	// Try in-cluster config first, fall back to kubeconfig if not found
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG") // Set this env var to your kubeconfig path
		if kubeconfig == "" {
			home, _ := os.UserHomeDir()
			kubeconfig = home + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			fmt.Printf("cannot load kubeconfig: %v\n", err)
			os.Exit(1)
		}
	}

	k8sClient, err := k8s.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return k8sClient
}
