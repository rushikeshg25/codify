package main

import (
	"context"
	"fmt"
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
	deploymentName := fmt.Sprintf("codeground-deployment-%d-%s", codeground.UserId, codeground.Id)

	deploymentClient := k.clientset.AppsV1().Deployments(k.namespace)

	_, err := deploymentClient.Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err == nil {
		log.Println("Deployment already exists:", deploymentName)
		return nil
	}
	var containerPort int32
	image := "rushikeshg25/nodejs-codeground:1.1.0"
	containerPort=8090
	if codeground.CodegroundType == "REACT" {
		image = "rushikeshg25/reactjs-codeground:1.1.0"
		containerPort=5173
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: deploymentName},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": deploymentName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": deploymentName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app-container",
							Image: image,
							Ports: []corev1.ContainerPort{
								{ContainerPort: containerPort},
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

func (k *K8s) CreateService(codeground Codeground) error {
	serviceName := fmt.Sprintf("codeground-service-%d-%s", codeground.UserId, codeground.Id)

	serviceClient := k.clientset.CoreV1().Services(k.namespace)
	var servicePort int32=5173;
	name:="react"
	if codeground.CodegroundType == "NODE" {
		name="node"
		servicePort=8090
	}


	_, err := serviceClient.Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err == nil {
		log.Println("Service already exists:", serviceName)
		return nil
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: serviceName},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": fmt.Sprintf("codeground-deployment-%d-%s", codeground.UserId, codeground.Id)},
			Ports: []corev1.ServicePort{
				{Name: name, Port: servicePort, TargetPort: intstrPtr(servicePort)},
				{Name: "websocket", Port: 9000, TargetPort: intstrPtr(9000)},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	_, err = serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
	return err
}

func (k *K8s) CreateIngress(codeground Codeground) error {
	ingressName := fmt.Sprintf("codeground-ingress-%d-%s", codeground.UserId, codeground.Id)
	serviceName := fmt.Sprintf("codeground-service-%d-%s", codeground.UserId, codeground.Id)

	ingressClient := k.clientset.NetworkingV1().Ingresses(k.namespace)

	_, err := ingressClient.Get(context.TODO(), ingressName, metav1.GetOptions{})
	if err == nil {
		log.Println("Ingress already exists:", ingressName)
		return nil
	}

	var port int32 = 5173
	if codeground.CodegroundType == "NODE" {
		port = 8090
	}

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: ingressName,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: fmt.Sprintf("app-%s.codify.localhost", codeground.Id),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{Number: port},
										},
									},
								},
							},
						},
					},
				},
				{
					Host: fmt.Sprintf("api-%s.codify.localhost", codeground.Id),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName, // Use the correct service name
											Port: networkingv1.ServiceBackendPort{Number: 9000},
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
	fmt.Printf("%+v\n", ingress)

	_, err = ingressClient.Create(context.TODO(), ingress, metav1.CreateOptions{})
	return err
}

func (k *K8s) DeleteDeployment(codeground Codeground) error {
	deploymentName := fmt.Sprintf("codeground-deployment-%d-%s", codeground.UserId, codeground.Id)
	return k.clientset.AppsV1().Deployments(k.namespace).Delete(context.TODO(), deploymentName, metav1.DeleteOptions{})
}

func (k *K8s) DeleteService(codeground Codeground) error {
	serviceName := fmt.Sprintf("codeground-service-%d-%s", codeground.UserId, codeground.Id)
	return k.clientset.CoreV1().Services(k.namespace).Delete(context.TODO(), serviceName, metav1.DeleteOptions{})
}

func (k *K8s) DeleteIngress(codeground Codeground) error {
	ingressName := fmt.Sprintf("codeground-ingress-%d-%s", codeground.UserId, codeground.Id)
	return k.clientset.NetworkingV1().Ingresses(k.namespace).Delete(context.TODO(), ingressName, metav1.DeleteOptions{})
}

var pathTypePrefix = networkingv1.PathTypePrefix

func intstrPtr(i int32) intstr.IntOrString {
	return intstr.IntOrString{Type: intstr.Int, IntVal: i}
}

func int32Ptr(i int32) *int32 { return &i }
