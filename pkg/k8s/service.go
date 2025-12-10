package k8s

import (
	"context"
	"fmt"
	"ys-cloud/internal/config"

	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/sirupsen/logrus"
)

type K8sService struct {
	clientset *kubernetes.Clientset
	config    *config.K8sConfig
	logger    *logrus.Logger
}

type DeploymentOptions struct {
	Name        string
	Namespace   string
	Image       string
	Tag         string
	Replicas    int32
	Port        int32
	EnvVars     []corev1.EnvVar
	Resources   *corev1.ResourceRequirements
	Labels      map[string]string
	Annotations map[string]string
}

type ServiceOptions struct {
	Name        string
	Namespace   string
	Selector    map[string]string
	Port        int32
	TargetPort  int32
	Type        corev1.ServiceType
	Labels      map[string]string
	Annotations map[string]string
}

type IngressOptions struct {
	Name        string
	Namespace   string
	Host        string
	ServiceName string
	ServicePort int32
	Labels      map[string]string
	Annotations map[string]string
}

func NewK8sService(cfg *config.Config) (*K8sService, error) {
	var k8sConfig *rest.Config
	var err error

	if cfg.K8s.Kubeconfig != "" && cfg.K8s.Kubeconfig != "default" {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", cfg.K8s.Kubeconfig)
	} else {
		// Try in-cluster config first, then fallback to kubeconfig
		k8sConfig, err = rest.InClusterConfig()
		if err != nil {
			home := homedir.HomeDir()
			if home == "" {
				return nil, fmt.Errorf("failed to get home directory")
			}
			kubeconfig := filepath.Join(home, ".kube", "config")
			k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return &K8sService{
		clientset: clientset,
		config:    &cfg.K8s,
		logger:    logrus.New(),
	}, nil
}

// getDefaultResourceRequirements returns default resource requirements if none are specified
func getDefaultResourceRequirements() *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
}

// getOrDefaultLabels returns labels or default empty map if nil
func getOrDefaultLabels(labels map[string]string, name string) map[string]string {
	if labels == nil {
		return map[string]string{
			"app":     name,
			"version": "v1",
		}
	}

	// Ensure essential labels exist
	if _, exists := labels["app"]; !exists {
		labels["app"] = name
	}
	if _, exists := labels["version"]; !exists {
		labels["version"] = "v1"
	}

	return labels
}

func (s *K8sService) CreateNamespace(name string) error {
	ctx := context.Background()

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"name": name,
				"app":  "ys-cloud",
			},
		},
	}

	_, err := s.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	s.logger.WithField("namespace", name).Info("Kubernetes namespace created")
	return nil
}

func (s *K8sService) Deploy(opts DeploymentOptions) error {
	ctx := context.Background()

	// Validate required fields
	if opts.Name == "" {
		return fmt.Errorf("deployment name is required")
	}
	if opts.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if opts.Image == "" {
		return fmt.Errorf("image is required")
	}
	if opts.Tag == "" {
		return fmt.Errorf("tag is required")
	}

	// Set defaults
	if opts.Replicas <= 0 {
		opts.Replicas = 1
	}
	if opts.Port <= 0 {
		opts.Port = 8080
	}

	// Use default resources if none specified
	resources := opts.Resources
	if resources == nil {
		resources = getDefaultResourceRequirements()
	}

	// Use default labels if none specified
	labels := getOrDefaultLabels(opts.Labels, opts.Name)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Namespace:   opts.Namespace,
			Labels:      labels,
			Annotations: opts.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &opts.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": opts.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  opts.Name,
							Image: fmt.Sprintf("%s:%s", opts.Image, opts.Tag),
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: opts.Port,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env:       opts.EnvVars,
							Resources: *resources,
							// Add health checks
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(int(opts.Port)),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(int(opts.Port)),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}

	_, err := s.clientset.AppsV1().Deployments(opts.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"deployment": opts.Name,
		"namespace":  opts.Namespace,
		"image":      fmt.Sprintf("%s:%s", opts.Image, opts.Tag),
		"replicas":   opts.Replicas,
	}).Info("Kubernetes deployment created")

	return nil
}

func (s *K8sService) UpdateDeployment(opts DeploymentOptions) error {
	ctx := context.Background()

	// Validate required fields
	if opts.Name == "" {
		return fmt.Errorf("deployment name is required")
	}
	if opts.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if opts.Image == "" {
		return fmt.Errorf("image is required")
	}
	if opts.Tag == "" {
		return fmt.Errorf("tag is required")
	}

	deployment, err := s.clientset.AppsV1().Deployments(opts.Namespace).Get(ctx, opts.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	// Update container image
	containerIndex := -1
	for i, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == opts.Name {
			containerIndex = i
			break
		}
	}

	if containerIndex == -1 {
		return fmt.Errorf("container %s not found in deployment", opts.Name)
	}

	deployment.Spec.Template.Spec.Containers[containerIndex].Image = fmt.Sprintf("%s:%s", opts.Image, opts.Tag)

	// Update environment variables if provided
	if opts.EnvVars != nil {
		deployment.Spec.Template.Spec.Containers[containerIndex].Env = opts.EnvVars
	}

	// Update resources if provided
	if opts.Resources != nil {
		deployment.Spec.Template.Spec.Containers[containerIndex].Resources = *opts.Resources
	}

	_, err = s.clientset.AppsV1().Deployments(opts.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"deployment": opts.Name,
		"namespace":  opts.Namespace,
		"image":      fmt.Sprintf("%s:%s", opts.Image, opts.Tag),
	}).Info("Kubernetes deployment updated")

	return nil
}

func (s *K8sService) CreateService(opts ServiceOptions) error {
	ctx := context.Background()

	// Validate required fields
	if opts.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if opts.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if opts.Port <= 0 {
		return fmt.Errorf("port is required")
	}
	if opts.TargetPort <= 0 {
		opts.TargetPort = opts.Port
	}

	// Set default selector if none provided
	if opts.Selector == nil {
		opts.Selector = map[string]string{
			"app": opts.Name,
		}
	}

	// Use default labels if none specified
	labels := getOrDefaultLabels(opts.Labels, opts.Name)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Namespace:   opts.Namespace,
			Labels:      labels,
			Annotations: opts.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Selector: opts.Selector,
			Ports: []corev1.ServicePort{
				{
					Port:       opts.Port,
					TargetPort: intstr.FromInt(int(opts.TargetPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: opts.Type,
		},
	}

	_, err := s.clientset.CoreV1().Services(opts.Namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"service":   opts.Name,
		"namespace": opts.Namespace,
		"port":      opts.Port,
		"type":      opts.Type,
	}).Info("Kubernetes service created")

	return nil
}

func (s *K8sService) CreateIngress(opts IngressOptions) error {
	ctx := context.Background()

	// Validate required fields
	if opts.Name == "" {
		return fmt.Errorf("ingress name is required")
	}
	if opts.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if opts.Host == "" {
		return fmt.Errorf("host is required")
	}
	if opts.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}
	if opts.ServicePort <= 0 {
		return fmt.Errorf("service port is required")
	}

	// Use default labels if none specified
	labels := getOrDefaultLabels(opts.Labels, opts.Name)

	pathType := networkingv1.PathTypePrefix

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Namespace:   opts.Namespace,
			Labels:      labels,
			Annotations: opts.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: opts.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: opts.ServiceName,
											Port: networkingv1.ServiceBackendPort{
												Number: opts.ServicePort,
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

	_, err := s.clientset.NetworkingV1().Ingresses(opts.Namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ingress: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"ingress":   opts.Name,
		"namespace": opts.Namespace,
		"host":      opts.Host,
	}).Info("Kubernetes ingress created")

	return nil
}

func (s *K8sService) GetDeploymentStatus(namespace, name string) (*appsv1.DeploymentStatus, error) {
	ctx := context.Background()

	deployment, err := s.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	return &deployment.Status, nil
}

func (s *K8sService) GetPodLogs(namespace, podName, containerName string) (string, error) {
	ctx := context.Background()

	// Validate required fields
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if podName == "" {
		return "", fmt.Errorf("pod name is required")
	}
	if containerName == "" {
		containerName = podName // Use pod name as default container name
	}

	req := s.clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: containerName,
		Follow:    false,
		Previous:  false,
		TailLines: func() *int64 { n := int64(1000); return &n }(),
	})

	logs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer logs.Close()

	buf := make([]byte, 1024)
	var result string
	for {
		n, err := logs.Read(buf)
		if err != nil {
			break
		}
		result += string(buf[:n])
	}

	return result, nil
}

func (s *K8sService) RollbackDeployment(namespace, deploymentName string) error {
	ctx := context.Background()

	// Validate required fields
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if deploymentName == "" {
		return fmt.Errorf("deployment name is required")
	}

	s.logger.WithFields(logrus.Fields{
		"deployment": deploymentName,
		"namespace":  namespace,
	}).Info("Rollback deployment requested")

	// Get current deployment
	deployment, err := s.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment for rollback: %w", err)
	}

	// Simple rollback: restart the deployment by updating annotations
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}

	// Add timestamp to force restart
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = fmt.Sprintf("%d", metav1.Now().Unix())

	_, err = s.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to rollback deployment: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"deployment": deploymentName,
		"namespace":  namespace,
	}).Info("Deployment rollback completed")

	return nil
}

func (s *K8sService) DeleteDeployment(namespace, name string) error {
	ctx := context.Background()

	// Validate required fields
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if name == "" {
		return fmt.Errorf("deployment name is required")
	}

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	err := s.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, deleteOptions)
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"deployment": name,
		"namespace":  namespace,
	}).Info("Kubernetes deployment deleted")

	return nil
}

// ScaleDeployment scales a deployment to the specified number of replicas
func (s *K8sService) ScaleDeployment(namespace, name string, replicas int32) error {
	ctx := context.Background()

	// Validate required fields
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if name == "" {
		return fmt.Errorf("deployment name is required")
	}
	if replicas < 0 {
		return fmt.Errorf("replicas cannot be negative")
	}

	scale, err := s.clientset.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment scale: %w", err)
	}

	scale.Spec.Replicas = replicas

	_, err = s.clientset.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"deployment": name,
		"namespace":  namespace,
		"replicas":   replicas,
	}).Info("Deployment scaled successfully")

	return nil
}
