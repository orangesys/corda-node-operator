package utils

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cordav1 "orangesys.io/cordanode/api/v1"
)

//GenNodeInfoConfigMap ...
func GenNodeInfoConfigMap(cr *cordav1.CordaNode) *corev1.ConfigMap {
	configMap := &corev1.ConfigMap{
		TypeMeta: GenMetaInfo("ConfigMap", "v1"),
		ObjectMeta: GenObjMetaInfo(cr.ObjectMeta.Name, cr.Namespace, map[string]string{
			"app": cr.ObjectMeta.Name,
		}),
		Data: map[string]string{
			"node.info": cr.Spec.NodeInfo,
		},
	}
	AddOwnerRefToObj(configMap, AsOwner(cr))
	return configMap
}

//CreateNodeInfoConfigMap ...
func CreateNodeInfoConfigMap(cr *cordav1.CordaNode) {
	log := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.ObjectMeta.Name)
	new := GenNodeInfoConfigMap(cr)
	old, err := GetClientSet().CoreV1().ConfigMaps(cr.Namespace).Get(cr.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		log.Info("Creating configMap", "ConfigMap.Name", cr.ObjectMeta.Name)
		GetClientSet().CoreV1().ConfigMaps(cr.Namespace).Create(new)
	} else if new != old {
		log.Info("Reconciling  configMap", "ConfigMap.Name", cr.ObjectMeta.Name)
		GetClientSet().CoreV1().ConfigMaps(cr.Namespace).Update(new)
	} else {
		log.Info("All configMap in sync", "ConfigMap.Name", cr.ObjectMeta.Name)
	}
}
