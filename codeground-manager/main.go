package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func int32Ptr(i int32) *int32 {
	return &i
}


func main(){
	var kubeconfig string
	home:=homedir.HomeDir()
	if home!=""{
		kubeconfig=filepath.Join(home,".kube","config")
	}

	config,err:=clientcmd.BuildConfigFromFlags("",kubeconfig)
	if err!=nil{
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	clientset,err:=kubernetes.NewForConfig(config)
	if err!=nil{
		log.Fatalf("Error building kubernetes clientset: %v", err)
	}
	log.Printf("Successfully connected to Kubernetes cluster")

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nodejs-pod",
			Labels: map[string]string{
				"app": "nodejs-app",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nodejs-container",
					Image: "rushikeshg25/node-app:1.0.0", 
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 8090,
						},
					},
				},
			},
		},
	}
	log.Printf("Creating pod %v", pod.Name)




	createPod,err:=clientset.CoreV1().Pods("default").Create(context.TODO(),pod,metav1.CreateOptions{})
	if err!=nil{
		log.Printf("Error creating pod %v", err)
	}
	log.Printf("Created pod %v", createPod.Name)
	fmt.Println("Listing Pods...")
	pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing Pods: %v", err)
	}
	for _, pod := range pods.Items {
		fmt.Printf("Pod: %s\n", pod.Name)
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "react-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1), // Number of replicas (pods)
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "react-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "react-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "react-container",
							Image: "rushikeshg25/react-app:1.0.0",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 5173, // Port exposed by the container
								},
							},
						},
					},
				},
			},
		},
	}

	// Step 4: Create the Deployment
	fmt.Println("Creating Deployment...")
	createdDeployment, err := clientset.AppsV1().Deployments("default").Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Error creating Deployment: %v", err)
	}
	fmt.Printf("Created Deployment: %s\n", createdDeployment.Name)

	// Step 5: List all Deployments in the default namespace
	fmt.Println("Listing Deployments...")
	deployments, err := clientset.AppsV1().Deployments("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing Deployments: %v", err)
	}
	for _, deployment := range deployments.Items {
		fmt.Printf("Deployment: %s (Replicas: %d)\n", deployment.Name, *deployment.Spec.Replicas)
	}
}