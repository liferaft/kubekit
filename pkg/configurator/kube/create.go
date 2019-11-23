package kube

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/resource"
)

// Create creates a resource with the given content
func (c *Client) Create(name string, content []byte) error {
	r := c.ResultForContent(name, content, true)
	return c.create(r)
}

// CreateFile creates a resource in the given local filename or HTTP URL
func (c *Client) CreateFile(filename string) error {
	filenames := []string{filename}
	r := c.ResultForFilenameParam(filenames, true)
	return c.create(r)
}

func (c *Client) create(r *resource.Result) error {
	return r.Visit(func(info *resource.Info, err error) error {
		var resKind string
		if info.Mapping != nil {
			resKind = info.Mapping.GroupVersionKind.Kind + " "
		}
		if err != nil {
			c.ui.Log.Debugf("cannot create object %s%q on namespace %s, received error: %s", resKind, info.Name, info.Namespace, err)
			return err
		}
		c.ui.Log.Debugf("creating object %s%q on namespace %s", resKind, info.Name, info.Namespace)

		// if err := kubectl.CreateApplyAnnotation(info.Object, unstructured.UnstructuredJSONScheme); err != nil {
		// 	return fmt.Errorf("creating %s. %s", info.String(), err)
		// }

		return create(info)
	})
}

func create(info *resource.Info) error {
	options := metav1.CreateOptions{}
	obj, err := resource.NewHelper(info.Client, info.Mapping).Create(info.Namespace, true, info.Object, &options)
	if err != nil {
		return fmt.Errorf("creating %s. %s", info.String(), err)
	}
	info.Refresh(obj, true)
	return nil
}

func (c *Client) reCreate(info *resource.Info) error {
	// TODO: this method is to delete and create the resource. Requires the
	// implementation of a delete method
	return nil
}
