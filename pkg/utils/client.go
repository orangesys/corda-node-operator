package utils

import (
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

//GetClientSet ...
func GetClientSet() *kubernetes.Clientset {
	config := ctrl.GetConfigOrDie()
	clientset, _ := kubernetes.NewForConfig(config)
	return clientset
}
