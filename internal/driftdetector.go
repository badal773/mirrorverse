package internal

import corev1 "k8s.io/api/core/v1"

// NeedsSync returns true if any key in the source's Data is missing or different in the replica.
// For merge/patch strategy: extra keys in the replica are ignored.
func NeedsSync(replicaObj, sourceObj interface{}) bool {
	// Handle ConfigMap comparison
	switch src := sourceObj.(type) {
	case *corev1.ConfigMap:
		replica, _ := replicaObj.(*corev1.ConfigMap)
		// For each key in the source's Data, check if replica has the same key and value
		for srcKey, srcVal := range src.Data {
			replicaVal, exists := replica.Data[srcKey]
			if !exists || replicaVal != srcVal {
				// Key missing or value different: needs sync
				return true
			}
		}
		// All source keys/values are present and match in replica
		return false
	case *corev1.Secret:
		replica, _ := replicaObj.(*corev1.Secret)
		for srcKey, srcVal := range src.Data {
			replicaVal, exists := replica.Data[srcKey]
			if !exists || string(replicaVal) != string(srcVal) {
				return true
			}
		}
		return false
	default:
		// Unknown type: always sync to be safe
		return false
	}
}
