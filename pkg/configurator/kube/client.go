package kube

import (
	"bytes"
	"io"

	"github.com/kraken/ui"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/kubectl/validation"
)

// Client is a kubernetes client, like `kubectl`
type Client struct {
	Config           *Config
	validator        validation.Schema
	builder          *resource.Builder
	namespace        string
	enforceNamespace bool
	clientset        *kubernetes.Clientset
	ui               *ui.UI
}

// NewClientE creates a kubernetes client, returns an error if fail
func NewClientE(context, kubeconfig string, ui *ui.UI) (*Client, error) {
	config := NewConfig(context, kubeconfig)

	// validator, _ := config.Validator(true)

	namespace, enforceNamespace, err := config.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return nil, err
	}
	clientset, err := config.KubernetesClientSet()
	if err != nil {
		return nil, err
	}

	return &Client{
		Config:           config,
		validator:        validation.NullSchema{},
		namespace:        namespace,
		enforceNamespace: enforceNamespace,
		clientset:        clientset,
		ui:               ui,
	}, nil
}

// NewClient creates a kubernetes client
func NewClient(context, kubeconfig string, ui *ui.UI) *Client {
	client, _ := NewClientE(context, kubeconfig, ui)
	return client
}

// ClientSet sets the client set in the config if its not already set
func (c *Client) ClientSet() error {
	if c.clientset == nil {
		clientset, err := c.Config.KubernetesClientSet()
		if err != nil {
			return err
		}
		c.clientset = clientset
	}
	return nil
}

// UnstructuredBuilder creates an unstructure builder for the given namespace
func (c *Client) UnstructuredBuilder() *resource.Builder {
	return c.Config.
		NewBuilder().
		Unstructured().
		Schema(c.validator).
		ContinueOnError().
		NamespaceParam(c.namespace).DefaultNamespace()
}

// Builder creates a builder for the given namespace
func (c *Client) Builder() *resource.Builder {
	return c.Config.
		NewBuilder().
		Schema(c.validator).
		ContinueOnError().
		NamespaceParam(c.namespace).DefaultNamespace()
}

// ResultForFilenameParam returns the builder results for the given list of files or URLs
func (c *Client) ResultForFilenameParam(filenames []string, unstructured bool) *resource.Result {
	filenameOptions := &resource.FilenameOptions{
		Recursive: false,
		Filenames: filenames,
	}

	var b *resource.Builder
	if unstructured {
		b = c.UnstructuredBuilder()
	} else {
		b = c.Builder()
	}

	return b.
		FilenameParam(c.enforceNamespace, filenameOptions).
		Flatten().
		Do()
}

// ResultForReader returns the builder results for the given reader
func (c *Client) ResultForReader(name string, r io.Reader, unstructured bool) *resource.Result {
	var b *resource.Builder
	if unstructured {
		b = c.UnstructuredBuilder()
	} else {
		b = c.Builder()
	}

	return b.
		Stream(r, name).
		Flatten().
		Do()
}

// ResultForContent returns the builder results for the given content
func (c *Client) ResultForContent(name string, content []byte, unstructured bool) *resource.Result {
	b := bytes.NewBuffer(content)
	return c.ResultForReader(name, b, unstructured)
}
