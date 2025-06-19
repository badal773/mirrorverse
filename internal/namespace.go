package internal

import (
	"fmt"
	"context"

	k8s "k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

)

func GetNsList(k8sClient  *k8s.Clientset) []string  { 
	
	namespacelist, err := k8sClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing namespaces: %v\n", err)
		return nil
	}
	var  nsNames []string ;
	fmt.Println("Current namespaces in the cluster:")
	for _, ns := range namespacelist.Items {
		nsNames = append(nsNames, ns.Name)
	}

	return nsNames
}
