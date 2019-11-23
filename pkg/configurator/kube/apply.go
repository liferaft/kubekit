package kube

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/resource"
)

// Apply creates a resource with the given content
func (c *Client) Apply(name string, content []byte) error {
	// DEBUG:
	// c.ui.Log.Debugf("applying resource with content: %s", string(content))

	c.ui.Log.Debugf("applying resource template: %q", name)
	r := c.ResultForContent(name, content, true)
	return c.ApplyResource(r)
}

// ApplyFile creates a resource in the given local filename or HTTP URL
func (c *Client) ApplyFile(filename string) error {
	c.ui.Log.Debugf("applying resource from file: %q", filename)
	filenames := []string{filename}
	r := c.ResultForFilenameParam(filenames, true)
	return c.ApplyResource(r)
}

// ApplyResource creates a resource with the resource.Result
func (c *Client) ApplyResource(r *resource.Result) error {
	return r.Visit(func(info *resource.Info, err error) error {
		var resKind string
		if info.Mapping != nil {
			resKind = info.Mapping.GroupVersionKind.Kind + " "
		}
		if err != nil {
			c.ui.Log.Debugf("cannot apply object %s%q on namespace %s, received error: %s", resKind, info.Name, info.Namespace, err)
			return err
		}
		c.ui.Log.Debugf("applying object %s%q on namespace %s", resKind, info.Name, info.Namespace)

		// if err := info.Get(); err != nil {
		originalObj, err := resource.NewHelper(info.Client, info.Mapping).Get(info.Namespace, info.Name, info.Export)
		if err != nil {
			if !errors.IsNotFound(err) {
				c.ui.Log.Errorf("retrieving current configuration of resource %s. %s", info.String(), err)
				return fmt.Errorf("retrieving current configuration of resource %s. %s", info.String(), err)
			}

			c.ui.Log.Debugf("creating object %s%q because was not found in namespace %s", resKind, info.Name, info.Namespace)
			return create(info)
		}

		c.ui.Log.Debugf("object %s%q found in namespace %s", resKind, info.Name, info.Namespace)
		return c.patch(info, originalObj)
	})

}

// ExistsResource returns true if the resources are found, otherwise returns
// false and no error. If failed to get the resources, it returns false and the error.
func (c *Client) ExistsResource(r *resource.Result) (bool, error) {
	err := r.Visit(func(info *resource.Info, err error) error {
		var resKind string
		if info.Mapping != nil {
			resKind = info.Mapping.GroupVersionKind.Kind + " "
		}
		if err != nil {
			c.ui.Log.Debugf("cannot get object %s%q on namespace %s, received error: %s", resKind, info.Name, info.Namespace, err)
			return err
		}
		c.ui.Log.Debugf("getting object %s%q on namespace %s", resKind, info.Name, info.Namespace)

		if err := info.Get(); err != nil {
			if !errors.IsNotFound(err) {
				// flase, failed to get the resource
				return fmt.Errorf("retrieving current configuration of resource %s. %s", info.String(), err)
			}
			// false, does not exists
			return err
		}
		c.ui.Log.Debugf("object %s%q found in namespace %s", resKind, info.Name, info.Namespace)
		// true, exists
		return nil
	})

	if err == nil {
		return true, nil
	}
	if errors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}
