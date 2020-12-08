package utils

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cordav1 "orangesys.io/cordanode/api/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("controller_corda")

//GenCertsSecret ...
func GenCertsSecret(cr *cordav1.CordaNode) *corev1.Secret {
	secret := &corev1.Secret{
		TypeMeta: GenMetaInfo("Secret", "v1"),
		ObjectMeta: GenObjMetaInfo(cr.ObjectMeta.Name, cr.Namespace, map[string]string{
			"app": cr.ObjectMeta.Name,
		}),
		Data: map[string][]byte{
			"nodekeystore.jks": []byte("nodekeystore.jks"),
			"truststore.jks":   []byte("truststore.jks"),
			"sslkeystore.jks":  []byte("sslkeystore.jks"),
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
