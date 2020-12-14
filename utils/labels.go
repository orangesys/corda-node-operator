package utils

import (
	cordav1 "github.com/orangesys/corda-node-operator/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GenMetaInfo ...
func GenMetaInfo(resourceKind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       resourceKind,
		APIVersion: apiVersion,
	}
}

//GenObjMetaInfo ...
func GenObjMetaInfo(name, namespace string, labels map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    labels,
	}
}

//AddOwnerRefToObj ...
func AddOwnerRefToObj(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

//AsOwner ...
func AsOwner(cr *cordav1.CordaNode) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: cr.APIVersion,
		Kind:       cr.Kind,
		Name:       cr.Name,
		UID:        cr.UID,
		Controller: &trueVar,
	}
}
