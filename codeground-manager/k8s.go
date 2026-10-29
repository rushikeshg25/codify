package main

import (
	"context"
	"fmt"
	"log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

/*
codegroundLabels stamp every resource with who it belongs to.

Names already encode the same thing, but parsing names is fragile and gives no
way to list. With these labels the manager can enumerate everything it owns and
compare that against the database, which is what makes orphan cleanup possible
at all.
*/
func codegroundLabels(codeground Codeground) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "codify",
		"codify/codeground":            codeground.Id,
		"codify/user":                  fmt.Sprint(codeground.UserId),
	}
}

// managedSelector matches every resource this manager created, regardless of
// which codeground it belongs to.
const managedSelector = "app.kubernetes.io/managed-by=codify"

// secretName holds the codeground's derived signing key. Kept out of the
// Deployment spec so the key is not readable by anyone with plain read access
// to workloads.
func secretName(codeground Codeground) string {
	return fmt.Sprintf("codeground-secret-%d-%s", codeground.UserId, codeground.Id)
}

func (k *K8s) CreateSecret(codeground Codeground) error {
	secretClient := k.clientset.CoreV1().Secrets(k.namespace)

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	name := secretName(codeground)
	_, err := secretClient.Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		log.Println("Secret already exists:", name)
		return nil
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: codegroundLabels(codeground)},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"token-key": codegroundKey(codeground.Id),
		},
	}

	_, err = secretClient.Create(ctx, secret, metav1.CreateOptions{})
	return err
}

func (k *K8s) DeleteSecret(codeground Codeground) error {
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return k.clientset.CoreV1().Secrets(k.namespace).Delete(ctx, secretName(codeground), metav1.DeleteOptions{})
}

func (k *K8s) CreateDeployment(codeground Codeground) error {
	deploymentName := fmt.Sprintf("codeground-deployment-%d-%s", codeground.UserId, codeground.Id)

	deploymentClient := k.clientset.AppsV1().Deployments(k.namespace)

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	_, err := deploymentClient.Get(ctx, deploymentName, metav1.GetOptions{})
	if err == nil {
		log.Println("Deployment already exists:", deploymentName)
		return nil
	}
	var containerPort int32
	image := "rushikeshg25/nodejs-codeground:1.1.0"
	containerPort = 8090
	if codeground.CodegroundType == "REACT" {
		image = "rushikeshg25/reactjs-codeground:1.1.0"
		containerPort = 5173
	}

	podLabels := codegroundLabels(codeground)
	podLabels["app"] = deploymentName

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Labels: codegroundLabels(codeground)},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": deploymentName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: podLabels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app-container",
							Image: image,
							Ports: []corev1.ContainerPort{
								{ContainerPort: containerPort},
								{ContainerPort: 9000},
							},
							// The agent refuses to start without these, which is
							// deliberate: a pod with no key would otherwise serve
							// an unauthenticated shell.
							Env: []corev1.EnvVar{
								{Name: "CODEGROUND_ID", Value: codeground.Id},
								{
									Name: "CODEGROUND_TOKEN_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: secretName(codeground)},
											Key:                  "token-key",
										},
									},
								},
								{Name: "CORS_ALLOW_ORIGIN", Value: corsOrigin()},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:              resource.MustParse("100m"),
									corev1.ResourceMemory:           resource.MustParse("128Mi"),
									corev1.ResourceEphemeralStorage: resource.MustParse("256Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
									// Without this one `dd if=/dev/zero` fills the
									// node disk and evicts neighbouring pods.
									corev1.ResourceEphemeralStorage: resource.MustParse("2Gi"),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								// The image installs into /app as root at build
								// time; at run time nothing needs privilege.
								RunAsNonRoot:             boolPtr(true),
								RunAsUser:                int64Ptr(codegroundUID),
								RunAsGroup:               int64Ptr(codegroundGID),
								AllowPrivilegeEscalation: boolPtr(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
								// Not read-only: the whole point of a codeground
								// is that the user writes files and runs installs.
								// Escape surface is cut by the capability drop and
								// the seccomp profile below instead.
								ReadOnlyRootFilesystem: boolPtr(false),
							},
						},
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot:   boolPtr(true),
						RunAsUser:      int64Ptr(codegroundUID),
						RunAsGroup:     int64Ptr(codegroundGID),
						FSGroup:        int64Ptr(codegroundGID),
						SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
					},
					// User code has no business talking to the API server, and
					// a mounted token is the shortest path from RCE in a pod to
					// control of the cluster.
					AutomountServiceAccountToken: boolPtr(false),
				},
			},
		},
	}

	_, err = deploymentClient.Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func (k *K8s) CreateService(codeground Codeground) error {
	serviceName := fmt.Sprintf("codeground-service-%d-%s", codeground.UserId, codeground.Id)

	serviceClient := k.clientset.CoreV1().Services(k.namespace)
	var servicePort int32 = 5173
	name := "react"
	if codeground.CodegroundType == "NODE" {
		name = "node"
		servicePort = 8090
	}

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	_, err := serviceClient.Get(ctx, serviceName, metav1.GetOptions{})
	if err == nil {
		log.Println("Service already exists:", serviceName)
		return nil
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: serviceName, Labels: codegroundLabels(codeground)},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": fmt.Sprintf("codeground-deployment-%d-%s", codeground.UserId, codeground.Id)},
			Ports: []corev1.ServicePort{
				{Name: name, Port: servicePort, TargetPort: intstrPtr(servicePort)},
				{Name: "websocket", Port: 9000, TargetPort: intstrPtr(9000)},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	_, err = serviceClient.Create(ctx, service, metav1.CreateOptions{})
	return err
}

func (k *K8s) CreateIngress(codeground Codeground) error {
	ingressName := fmt.Sprintf("codeground-ingress-%d-%s", codeground.UserId, codeground.Id)
	serviceName := fmt.Sprintf("codeground-service-%d-%s", codeground.UserId, codeground.Id)

	ingressClient := k.clientset.NetworkingV1().Ingresses(k.namespace)

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	_, err := ingressClient.Get(ctx, ingressName, metav1.GetOptions{})
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
			Name:   ingressName,
			Labels: codegroundLabels(codeground),
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
	_, err = ingressClient.Create(ctx, ingress, metav1.CreateOptions{})
	return err
}

/*
CreateNetworkPolicy locks down what a codeground pod can talk to.

Without a policy the pod can reach MySQL, RabbitMQ, the API server, sibling
codegrounds and — most dangerously — the cloud metadata endpoint at
169.254.169.254, which hands out the node's IAM credentials. Egress is
therefore restricted to DNS plus the public internet with the private and
link-local ranges carved out, which still leaves package installs working.
*/
func (k *K8s) CreateNetworkPolicy(codeground Codeground) error {
	policyName := fmt.Sprintf("codeground-netpol-%d-%s", codeground.UserId, codeground.Id)
	deploymentName := fmt.Sprintf("codeground-deployment-%d-%s", codeground.UserId, codeground.Id)

	policyClient := k.clientset.NetworkingV1().NetworkPolicies(k.namespace)

	ctx, cancel := ctxWithTimeout()
	defer cancel()

	_, err := policyClient.Get(ctx, policyName, metav1.GetOptions{})
	if err == nil {
		log.Println("NetworkPolicy already exists:", policyName)
		return nil
	}

	tcp := corev1.ProtocolTCP
	udp := corev1.ProtocolUDP
	dnsPort := intstrPtr(53)
	appPort := intstrPtr(appPortFor(codeground))
	wssPort := intstrPtr(wssPort)

	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: policyName, Labels: codegroundLabels(codeground)},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"app": deploymentName},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			// Reachable only on the app port and the agent port. Narrowing the
			// source to the ingress controller's namespace is the next step;
			// it needs that namespace to carry a known label.
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{Protocol: &tcp, Port: &appPort},
						{Protocol: &tcp, Port: &wssPort},
					},
				},
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{Protocol: &udp, Port: &dnsPort},
						{Protocol: &tcp, Port: &dnsPort},
					},
				},
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							IPBlock: &networkingv1.IPBlock{
								CIDR: "0.0.0.0/0",
								Except: []string{
									"10.0.0.0/8",
									"172.16.0.0/12",
									"192.168.0.0/16",
									// Cloud instance metadata; node IAM credentials.
									"169.254.0.0/16",
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = policyClient.Create(ctx, policy, metav1.CreateOptions{})
	return err
}

func (k *K8s) DeleteNetworkPolicy(codeground Codeground) error {
	policyName := fmt.Sprintf("codeground-netpol-%d-%s", codeground.UserId, codeground.Id)
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return k.clientset.NetworkingV1().NetworkPolicies(k.namespace).Delete(ctx, policyName, metav1.DeleteOptions{})
}

func (k *K8s) DeleteDeployment(codeground Codeground) error {
	deploymentName := fmt.Sprintf("codeground-deployment-%d-%s", codeground.UserId, codeground.Id)
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return k.clientset.AppsV1().Deployments(k.namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{})
}

func (k *K8s) DeleteService(codeground Codeground) error {
	serviceName := fmt.Sprintf("codeground-service-%d-%s", codeground.UserId, codeground.Id)
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return k.clientset.CoreV1().Services(k.namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
}

func (k *K8s) DeleteIngress(codeground Codeground) error {
	ingressName := fmt.Sprintf("codeground-ingress-%d-%s", codeground.UserId, codeground.Id)
	ctx, cancel := ctxWithTimeout()
	defer cancel()
	return k.clientset.NetworkingV1().Ingresses(k.namespace).Delete(ctx, ingressName, metav1.DeleteOptions{})
}

// codegroundUID/GID are the unprivileged user the boilerplate images create.
const (
	codegroundUID = 1000
	codegroundGID = 1000
)

// wssPort is the in-pod agent (terminal + file API) port, the same for both
// codeground types.
const wssPort int32 = 9000

// corsOrigin is the browser origin the in-pod agent will accept. It has to
// match the frontend the user actually loads.
func corsOrigin() string {
	return envOr("CORS_ALLOW_ORIGIN", "http://localhost:3000")
}

// appPortFor returns the port the user's own dev server listens on.
func appPortFor(codeground Codeground) int32 {
	if codeground.CodegroundType == "NODE" {
		return 8090
	}
	return 5173
}

func ctxWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

var pathTypePrefix = networkingv1.PathTypePrefix

func intstrPtr(i int32) intstr.IntOrString {
	return intstr.IntOrString{Type: intstr.Int, IntVal: i}
}

func int32Ptr(i int32) *int32 { return &i }

func int64Ptr(i int64) *int64 { return &i }

func boolPtr(b bool) *bool { return &b }
