package utils

import (
	cordav1 "github.com/orangesys/corda-node-operator/api/v1"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GenCordaNodeDeployment ...
func GenCordaNodeDeployment(cr *cordav1.CordaNode) *appv1.Deployment {
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
						Name:            "app",
						Image:           "corda-node:latest",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("2Gi"),
								corev1.ResourceCPU:    resource.MustParse("2"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("2Gi"),
								corev1.ResourceCPU:    resource.MustParse("2"),
							},
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 10200,
							Name:          "p2p",
							Protocol:      corev1.ProtocolTCP,
						}, {
							ContainerPort: 8080,
							Name:          "braid",
							Protocol:      corev1.ProtocolTCP,
						}, {
							ContainerPort: 2222,
							Name:          "ssh",
							Protocol:      corev1.ProtocolTCP,
						}},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "certificates",
								MountPath: "/opt/corda/certificates",
							}, {
								Name:      "nodeconf",
								MountPath: "/etc/corda",
							},
						},
					}},
					Volumes: []corev1.Volume{{
						Name: "certificates",
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
