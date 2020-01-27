package kube

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Nodes returns the list of nodes
func (c *Client) Nodes() (*v1.NodeList, error) {
	return c.clientset.CoreV1().Nodes().List(metav1.ListOptions{})
}

// NodesReady returns the number of nodes ready
func (c *Client) NodesReady() (int, int, error) {
	nodes, err := c.Nodes()
	if err != nil {
		return 0, 0, err
	}
	if len(nodes.Items) == 0 {
		return 0, 0, nil
	}
	readyCount := 0
	for _, n := range nodes.Items {
		for _, c := range n.Status.Conditions {
			if c.Type == "Ready" && c.Status == "True" {
				readyCount++
				break
			}
		}
	}

	return readyCount, len(nodes.Items), nil
}
