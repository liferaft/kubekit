package kube

// NodesReady returns the number of nodes ready
func (c *Client) NodesReady() (int, int, error) {
	nodes, err := c.ListNodes()
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
