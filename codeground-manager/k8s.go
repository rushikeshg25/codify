package main

import (
	"context"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type K8s struct {
	clientset *kubernetes.Clientset
	namespace string
}


func NewK8s(clientset *kubernetes.Clientset, namespace string) *K8s {
	return &K8s{
		clientset: clientset,
		namespace: namespace,
	}
}



func (k *K8s) CreateDeployment(codeground Codeground) error {
	deploymentClient := k.clientset.AppsV1().Deployments(k.namespace)

	_, err := deploymentClient.Get(context.TODO(), "codeground-deployment", metav1.GetOptions{})
	if err == nil {
		log.Println("Deployment already exists, skipping creation.")
		return nil
	}

	var image string
	if codeground.CodegroundType == "react" {
		image = "rushikeshg25/reactjs-codeground:1.1.0"
	} else {
		image = "rushikeshg25/nodejs-codeground:1.1.0"
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "codeground-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "codeground"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "codeground"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "react-app",
							Image: image,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 5173},
								{ContainerPort: 9000},
							},
						},
					},
				},
			},
		},
	}

	_, err = deploymentClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	return err
}

func int32Ptr(i int32) *int32 { return &i }

func (k *K8s)CreateService(codeground Codeground) error {
	serviceClient := k.clientset.CoreV1().Services(k.namespace)

	_, err := serviceClient.Get(context.TODO(), "codeground-service", metav1.GetOptions{})
	if err == nil {
		log.Println("Service already exists")
		return nil 
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "codeground-service",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "codeground"},
			Ports: []corev1.ServicePort{
				{Name: "react", Port: 5173, TargetPort: intstrPtr(5173)},
				{Name: "websocket", Port: 9000, TargetPort: intstrPtr(9000)},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	_, err = serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
	return err
}


func (k *K8s)CreateIngress(codeground Codeground) error {
	ingressClient := k.clientset.NetworkingV1().Ingresses(k.namespace)

	
	_, err := ingressClient.Get(context.TODO(), "codeground-ingress", metav1.GetOptions{})
	if err == nil {
		log.Println("Ingress already exists")
		return nil 
	}

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: "codeground-ingress",
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "app.rushikesh.localhost",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "codeground-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 5173,
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Host: "api.rushikesh.localhost",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "codeground-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 9000,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = ingressClient.Create(context.TODO(), ingress, metav1.CreateOptions{})
	return err
}

var pathTypePrefix = networkingv1.PathTypePrefix


func intstrPtr(i int32) intstr.IntOrString {
	return intstr.IntOrString{Type: intstr.Int, IntVal: i}
}