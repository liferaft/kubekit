package kube

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListNodes returns the list of nodes
func (c *Client) ListNodes() (*corev1.NodeList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.CoreV1().Nodes().List(metav1.ListOptions{})
}

// ListDaemonSets returns a list of daemonsets for a namespace, set to empty string to get from all
func (c *Client) ListDaemonSets(namespace string) (*appsv1.DaemonSetList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.AppsV1().DaemonSets(namespace).List(metav1.ListOptions{})
}

// ListStatefulSets returns a list of statefulsets for a namespace, set to empty string to get from all
func (c *Client) ListStatefulSets(namespace string) (*appsv1.StatefulSetList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{})
}

// ListDeployments returns a list of deployments for a namespace, set to empty string to get from all
func (c *Client) ListDeployments(namespace string) (*appsv1.DeploymentList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
}

// ListReplicaSets returns a list of deployments for a namespace, set to empty string to get from all
func (c *Client) ListReplicaSets(namespace string) (*appsv1.ReplicaSetList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.AppsV1().ReplicaSets(namespace).List(metav1.ListOptions{})
}

// ListServices returns a list of services for a namespace, set to empty string to get from all
func (c *Client) ListServices(namespace string) (*corev1.ServiceList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.CoreV1().Services(namespace).List(metav1.ListOptions{})
}

// ListEndpoints returns a list of endpoints for a namespace, set to empty string to get from all
func (c *Client) ListEndpoints(namespace string) (*corev1.EndpointsList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.CoreV1().Endpoints(namespace).List(metav1.ListOptions{})
}

// ListReplicationControllers returns a list of replication controllers for a namespace, set to empty string to get from all
func (c *Client) ListReplicationControllers(namespace string) (*corev1.ReplicationControllerList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.CoreV1().ReplicationControllers(namespace).List(metav1.ListOptions{})
}

// ListPods returns a list of pods for a namespace, set to empty string to get from all
func (c *Client) ListPods(namespace string) (*corev1.PodList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
}

// ListConfigMaps returns a list of configmaps for a namespace, set to empty string to get from all
func (c *Client) ListConfigMaps(namespace string) (*corev1.ConfigMapList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{})
}

// ListSecrets returns a list of secrets for a namespace, set to empty string to get from all
func (c *Client) ListSecrets(namespace string) (*corev1.SecretList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
}

// ListServiceAccounts returns a list of service accounts for a namespace, set to empty string to get from all
func (c *Client) ListServiceAccounts(namespace string) (*corev1.ServiceAccountList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.CoreV1().ServiceAccounts(namespace).List(metav1.ListOptions{})
}

// ListPersistentVolumeClaims returns a list of persistent volume claims for a namespace, set to empty string to get from all
func (c *Client) ListPersistentVolumeClaims(namespace string) (*corev1.PersistentVolumeClaimList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
}

// ListPersistentVolumes returns a list of persistent volumes
func (c *Client) ListPersistentVolumes() (*corev1.PersistentVolumeList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
}

// ListNamespaces returns a list of namespaces
func (c *Client) ListNamespaces() (*corev1.NamespaceList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
}

// ListJobs returns a list of jobs for a namespace, set to empty string to get from all
func (c *Client) ListJobs(namespace string) (*batchv1.JobList, error) {
	err := c.ClientSet()
	if err != nil {
		return nil, err
	}
	return c.clientset.BatchV1().Jobs(namespace).List(metav1.ListOptions{})
}
