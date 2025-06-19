// helpers to parse labels
package internal

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

func CheckIfLabelsPresent(resource string,name string,obj interface{}) bool {
	
    fmt.Printf("Checking labels for %s %s \n", name,resource)
	labels := GetLabels(obj)
	if labels == nil {
		fmt.Printf("No labels found for %s %s\n", resource, name)
		return false
	}
	for key, value := range labels {
		if key == "mirrorverse.dev/sync-source" && value == "true" {
			fmt.Printf("Label %s=%s found for %s %s\n", key, value, resource, name)
			return true
		}
	}
	return false;
}



func GetLabels(obj interface{}) map[string]string {
	switch o := obj.(type) {
	case *corev1.ConfigMap:
		return o.GetObjectMeta().GetLabels()
	case *corev1.Secret:
		return o.GetObjectMeta().GetLabels()
	default:
		return nil
	}
}
