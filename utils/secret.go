package utils

import (
	"fmt"
	"io/ioutil"
	"os"

	cordav1 "github.com/orangesys/corda-node-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("controller_corda")

//GenCertsSecret ...
func GenCertsSecret(cr *cordav1.CordaNode) *corev1.Secret {
	src := fmt.Sprintf("http://a4d5963d66b7146db818b39bda4813a9-230794927.ap-northeast-1.elb.amazonaws.com:8080/v1/corda/certs?myLegalName=%s&p2pAddress=%s", cr.Spec.MyLegalName, cr.Status.ExternalIP+":10200")
	basePath := "/tmp/" + cr.ObjectMeta.Name
	dst := basePath + "/certs.zip"
	Download(src, dst)
	os.MkdirAll(basePath, os.ModePerm)
	UnZip(basePath, dst)
	f1, _ := ioutil.ReadFile(basePath + "/nodekeystore.jks")
	f2, _ := ioutil.ReadFile(basePath + "/truststore.jks")
	f3, _ := ioutil.ReadFile(basePath + "/sslkeystore.jks")

	secret := &corev1.Secret{
		TypeMeta: GenMetaInfo("Secret", "v1"),
		ObjectMeta: GenObjMetaInfo(cr.ObjectMeta.Name, cr.Namespace, map[string]string{
			"app": cr.ObjectMeta.Name,
		}),
		Data: map[string][]byte{
			"nodekeystore.jks": f1,
			"truststore.jks":   f2,
			"sslkeystore.jks":  f3,
		},
	}

	AddOwnerRefToObj(secret, AsOwner(cr))
	return secret
}

//CreateCertsSecret ...
func CreateCertsSecret(cr *cordav1.CordaNode) {
	log := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.ObjectMeta.Name)
	new := GenCertsSecret(cr)
	old, err := GetClientSet().CoreV1().Secrets(cr.Namespace).Get(cr.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		log.Info("Creating certs secrets", "Secret.Name", cr.ObjectMeta.Name)
		GetClientSet().CoreV1().Secrets(cr.Namespace).Create(new)
	} else if new != old {
		log.Info("Reconciling certs secrets", "Secret.Name", cr.ObjectMeta.Name)
		GetClientSet().CoreV1().Secrets(cr.Namespace).Update(new)
	} else {
		log.Info("All certs secrets in sync", "Secret.Name", cr.ObjectMeta.Name)
	}
}
