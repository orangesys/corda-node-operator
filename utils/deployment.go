package utils

import (
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cordav1 "orangesys.io/cordanode/api/v1"
)

//GenCordaNodeDeployment ...
func GenCordaNodeDeployment(cr *cordav1.CordaNode) *appv1.Deployment {
	parser := NewNodeInfoParser(cr)
	p2pPort, err := parser.GetP2PAddressPort()
	if err != nil {
		log.Error(err, "Parsing p2p port error, will use default 11002", "Request.Namespace", cr.Namespace, "Request.Name", cr.ObjectMeta.Name)
		p2pPort = 11002
	}
	deployment := &appv1.Deployment{
		TypeMeta:   GenMetaInfo("Deployment", "apps/v1"),
		ObjectMeta: GenObjMetaInfo(cr.ObjectMeta.Name, cr.Namespace, map[string]string{}),
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: cr.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: cr.Labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            "web",
						Image:           cr.Spec.WebImage,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 10055,
							Name:          "rest",
							Protocol:      corev1.ProtocolTCP,
						}},
					}, {
						Name:            "app",
						Image:           cr.Spec.APPImage,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports: []corev1.ContainerPort{{
							ContainerPort: p2pPort,
							Name:          "p2p",
							Protocol:      corev1.ProtocolTCP,
						}, {
							ContainerPort: 8080,
							Name:          "metrics",
							Protocol:      corev1.ProtocolTCP,
						}},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "certs",
								MountPath: "/opt/corda/persistence",
							}, {
								Name:      "nodeconf",
								MountPath: "/etc/corda",
							},
						},
					}},
					Volumes: []corev1.Volume{{
						Name: "certs",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: cr.ObjectMeta.Name,
							},
						},
					}, {
						Name: "nodeconf",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.ObjectMeta.Name,
								},
							},
						},
					}},
				},
			},
		},
	}
	AddOwnerRefToObj(deployment, AsOwner(cr))
	return deployment
}

//CreateCordaNodeDeployment ...
func CreateCordaNodeDeployment(cr *cordav1.CordaNode) {
	log := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.ObjectMeta.Name)
	new := GenCordaNodeDeployment(cr)
	old, err := GetClientSet().AppsV1().Deployments(cr.Namespace).Get(cr.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		log.Info("Creating deployment", "Deployment.Name", cr.ObjectMeta.Name)
		_, err := GetClientSet().AppsV1().Deployments(cr.Namespace).Create(new)
		if err != nil {
			log.Error(err, "Create deployment failed", "Deployment.Name", cr.ObjectMeta.Name)
		}
	} else if new != old {
		log.Info("Reconciling deployment", "Deployment.Name", cr.ObjectMeta.Name)
		GetClientSet().AppsV1().Deployments(cr.Namespace).Update(new)
	} else {
		log.Info("All deployment in sync", "Deployment.Name", cr.ObjectMeta.Name)
	}
}
