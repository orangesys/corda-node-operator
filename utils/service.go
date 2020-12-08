package utils

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	cordav1 "orangesys.io/cordanode/api/v1"
)

//GenService ...
func GenService(cr *cordav1.CordaNode) *corev1.Service {
	parser := NewNodeInfoParser(cr)
	p2pPort, err := parser.GetP2PAddressPort()
	if err != nil {
		log.Error(err, "Parsing p2p port error, will use default 11002", "Request.Namespace", cr.Namespace, "Request.Name", cr.ObjectMeta.Name)
		p2pPort = 11002
	}
	service := &corev1.Service{
		TypeMeta: GenMetaInfo("Service", "core/v1"),
		ObjectMeta: GenObjMetaInfo(cr.ObjectMeta.Name, cr.Namespace, map[string]string{
			"app.kubernetes.io/name":    fmt.Sprintf("%s-metrics", cr.Name),
			"app.kubernetes.io/part-of": cr.Name,
			"app":                       "corda",
		}),
		Spec: corev1.ServiceSpec{
			Selector: cr.Labels,
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "p2p",
					Protocol:   corev1.ProtocolTCP,
					Port:       p2pPort,
					TargetPort: intstr.FromInt(int(p2pPort)),
				}, {
					Name:       "rest",
					Protocol:   corev1.ProtocolTCP,
					Port:       10055,
					TargetPort: intstr.FromInt(int(10055)),
				},
				{
					Name:       "metrics",
					Protocol:   corev1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(int(8080)),
				},
			},
		},
	}
	AddOwnerRefToObj(service, AsOwner(cr))
	return service
}

//CreateService ...
func CreateService(cr *cordav1.CordaNode) {
	log := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.ObjectMeta.Name)
	new := GenService(cr)
	old, err := GetClientSet().CoreV1().Services(cr.Namespace).Get(cr.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		log.Info("Creating service", "Service.Name", cr.ObjectMeta.Name)
		GetClientSet().CoreV1().Services(cr.Namespace).Create(new)
	} else if new != old {
		log.Info("Reconciling service", "Service.Name", cr.ObjectMeta.Name)
		GetClientSet().CoreV1().Services(cr.Namespace).Update(new)
	} else {
		log.Info("All service in sync", "Service.Name", cr.ObjectMeta.Name)
	}
}
