package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"

	cordav1 "github.com/orangesys/corda-node-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GenNodeInfoConfigMap ...
func GenNodeInfoConfigMap(cr *cordav1.CordaNode) (*corev1.ConfigMap, error) {
	src := fmt.Sprintf("http://a4d5963d66b7146db818b39bda4813a9-230794927.ap-northeast-1.elb.amazonaws.com:8080/v1/corda/nodeconf?myLegalName=%s&p2pAddress=%s", cr.Spec.MyLegalName, cr.Status.ExternalIP+":10200")

	fmt.Println("url\n", src)
	resp, err := http.Get(src)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	configMap := &corev1.ConfigMap{
		TypeMeta: GenMetaInfo("ConfigMap", "v1"),
		ObjectMeta: GenObjMetaInfo(cr.ObjectMeta.Name, cr.Namespace, map[string]string{
			"app": cr.ObjectMeta.Name,
		}),
		Data: map[string]string{
			"node.conf": string(bytes),
		},
	}
	AddOwnerRefToObj(configMap, AsOwner(cr))
	return configMap, nil
}

//CreateNodeInfoConfigMap ...
func CreateNodeInfoConfigMap(cr *cordav1.CordaNode) error {
	log := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.ObjectMeta.Name)
	new, err := GenNodeInfoConfigMap(cr)
	if err != nil {
		log.Error(err, "Generage configMap error", "ConfigMap.Name", cr.ObjectMeta.Name)
		return err
	}
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
	return nil
}
