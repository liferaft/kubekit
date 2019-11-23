package kube

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
)

// Update creates a resource with the given content
func (c *Client) Update(name string, content []byte) error {
	r := c.ResultForContent(name, content, true)
	return c.update(r)
}

// UpdateFile creates a resource in the given local filename or HTTP URL
func (c *Client) UpdateFile(filename string) error {
	filenames := []string{filename}
	r := c.ResultForFilenameParam(filenames, true)
	return c.update(r)
}

func (c *Client) update(r *resource.Result) error {
	return r.Visit(func(info *resource.Info, err error) error {
		var resKind string
		if info.Mapping != nil {
			resKind = info.Mapping.GroupVersionKind.Kind + " "
		}
		if err != nil {
			c.ui.Log.Debugf("cannot update object %s%q on namespace %s, received error: %s", resKind, info.Name, info.Namespace, err)
			return err
		}
		c.ui.Log.Debugf("updating object %s%q on namespace %s", resKind, info.Name, info.Namespace)

		originalObj, err := resource.NewHelper(info.Client, info.Mapping).Get(info.Namespace, info.Name, info.Export)
		if err != nil {
			c.ui.Log.Errorf("retrieving current configuration of resource %s. %s", info.String(), err)
			return fmt.Errorf("retrieving current configuration of resource %s. %s", info.String(), err)
		}
		return c.patch(info, originalObj)
	})
}

func (c *Client) patch(info *resource.Info, current runtime.Object) error {
	// TODO: force will be a parameter when `reCreate()` is implemented
	force := false

	var resKind string
	if info.Mapping != nil {
		resKind = info.Mapping.GroupVersionKind.Kind + " "
	}
	patch, patchType, err := createPatch(info, current)
	if err != nil {
		return fmt.Errorf("creating patch. %s", err)
	}
	if patch == nil {
		c.ui.Log.Infof("there is nothing to update on %s%q", resKind, info.Name)
		return nil
	}

	obj, err := resource.NewHelper(info.Client, info.Mapping).Patch(info.Namespace, info.Name, patchType, patch, nil)
	if err != nil {
		c.ui.Log.Debugf("cannot patch object %s%q on namespace %s, received error: %s", resKind, info.Name, info.Namespace, err)
		if force {
			return c.reCreate(info)
		}
		return err
	}

	info.Refresh(obj, true)
	return nil
}

func createPatch(info *resource.Info, current runtime.Object) ([]byte, types.PatchType, error) {
	oldData, err := json.Marshal(current)
	if err != nil {
		return nil, types.StrategicMergePatchType, fmt.Errorf("serializing current configuration: %s", err)
	}
	newData, err := json.Marshal(info.Object)
	if err != nil {
		return nil, types.StrategicMergePatchType, fmt.Errorf("serializing info configuration: %s", err)
	}

	// While different objects need different merge types, the parent function
	// that calls this does not try to create a patch when the data (first
	// returned object) is nil. We can skip calculating the merge type as
	// the returned merge type is ignored.
	if equality.Semantic.DeepEqual(oldData, newData) {
		return nil, types.StrategicMergePatchType, nil
	}

	converter := runtime.ObjectConvertor(scheme.Scheme)
	groupVersioner := runtime.GroupVersioner(schema.GroupVersions(scheme.Scheme.PrioritizedVersionsAllGroups()))
	if info.Mapping != nil {
		groupVersioner = info.Mapping.GroupVersionKind.GroupVersion()
	}

	versionedObject := info.Object
	obj, err := converter.ConvertToVersion(info.Object, groupVersioner)
	if err == nil {
		versionedObject = obj
	}

	// Unstructured objects, such as CRDs, may not have an not registered error
	// returned from ConvertToVersion. Anything that's unstructured should
	// use the jsonpatch.CreateMergePatch. Strategic Merge Patch is not supported
	// on objects like CRDs.
	_, isUnstructured := versionedObject.(runtime.Unstructured)

	switch {
	case runtime.IsNotRegisteredError(err), isUnstructured:
		// fall back to generic JSON merge patch
		patch, err := jsonpatch.CreateMergePatch(oldData, newData)
		return patch, types.MergePatchType, err
	case err != nil:
		return nil, types.StrategicMergePatchType, fmt.Errorf("failed to get versionedObject: %s", err)
	default:
		patch, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, versionedObject)
		return patch, types.StrategicMergePatchType, err
	}
}

// TODO: The update can be improved, this code below is from `kubectl` to update
// a resource (with some modifications), use it to improve the current patching.

// func (c *Client) patch(info *resource.Info) error {
// 	// Get the modified configuration of the object. Embed the result
// 	// as an annotation in the modified configuration, so that it will appear
// 	// in the patch sent to the server.
// 	modified, err := getModifiedConfiguration(info.Object, true, unstructured.UnstructuredJSONScheme)
// 	if err != nil {
// 		return fmt.Errorf("retrieving modified configuration from: %s. %v", info.String(), err)
// 	}

// 	helper := resource.NewHelper(info.Client, info.Mapping)
// 	client, err := c.Config.DynamicClient()
// 	if err != nil {
// 		return err
// 	}

// 	patcher := &Patcher{
// 		Mapping:       info.Mapping,
// 		Helper:        helper,
// 		Overwrite:     true,
// 		Force:         false,
// 		Cascade:       true,
// 		Timeout:       time.Duration(0),
// 		GracePeriod:   -1,
// 		DynamicClient: client,
// 	}

// 	patchBytes, patchedObject, err := patcher.Patch(info.Object, modified, info.Source, info.Namespace, info.Name)
// 	if err != nil {
// 		return fmt.Errorf("applying patch: %s to: %s. %v", patchBytes, info.Name, err)
// 	}

// 	info.Refresh(patchedObject, true)
// 	return nil
// }

// // Patcher ...
// type Patcher struct {
// 	Mapping       *meta.RESTMapping
// 	Helper        *resource.Helper
// 	Overwrite     bool
// 	Force         bool
// 	Cascade       bool
// 	Timeout       time.Duration
// 	GracePeriod   int
// 	DynamicClient dynamic.Interface
// }

// const (
// 	// maxPatchRetry is the maximum number of conflicts retry for during a patch operation before returning failure
// 	maxPatchRetry = 5
// 	// backOffPeriod is the period to back off when apply patch resutls in error.
// 	backOffPeriod = 1 * time.Second
// 	// how many times we can retry before back off
// 	triesBeforeBackOff = 1
// )

// Patch ...
// func (p *Patcher) Patch(current runtime.Object, modified []byte, source, namespace, name string) ([]byte, runtime.Object, error) {
// 	var getErr error
// 	patchBytes, patchObject, err := p.patchSimple(current, modified, source, namespace, name)
// 	for i := 1; i <= maxPatchRetry && errors.IsConflict(err); i++ {
// 		if i > triesBeforeBackOff {
// 			time.Sleep(backOffPeriod)
// 		}
// 		current, getErr = p.Helper.Get(namespace, name, false)
// 		if getErr != nil {
// 			return nil, nil, getErr
// 		}
// 		patchBytes, patchObject, err = p.patchSimple(current, modified, source, namespace, name)
// 	}
// 	if err != nil && (errors.IsConflict(err) || errors.IsInvalid(err)) && p.Force {
// 		patchBytes, patchObject, err = p.deleteAndCreate(current, modified, namespace, name)
// 	}
// 	return patchBytes, patchObject, err
// }

// func (p *Patcher) patchSimple(obj runtime.Object, modified []byte, source, namespace, name string) ([]byte, runtime.Object, error) {
// 	// Serialize the current configuration of the object from the server.
// 	current, err := runtime.Encode(unstructured.UnstructuredJSONScheme, obj)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("serializing current configuration from: %v. %v", obj, err)
// 	}

// 	// Retrieve the original configuration of the object from the annotation.
// 	annots, err := metadataAccessor.Annotations(obj)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("retrieving original configuration from: %v. %v", obj, err)
// 	}

// 	var original []byte
// 	if annots != nil {
// 		if originalAnnots, ok := annots[v1.LastAppliedConfigAnnotation]; ok {
// 			original = []byte(originalAnnots)
// 		}
// 	}

// 	var patchType types.PatchType
// 	var patch []byte
// 	var lookupPatchMeta strategicpatch.LookupPatchMeta
// 	// var schema oapi.Schema
// 	createPatchErrFormat := "creating patch with: original: %s modified: %s current: %s. %v"

// 	// Create the versioned struct from the type defined in the restmapping
// 	// (which is the API version we'll be submitting the patch to)
// 	versionedObject, err := scheme.Scheme.New(p.Mapping.GroupVersionKind)
// 	switch {
// 	case runtime.IsNotRegisteredError(err):
// 		// fall back to generic JSON merge patch
// 		patchType = types.MergePatchType
// 		preconditions := []mergepatch.PreconditionFunc{mergepatch.RequireKeyUnchanged("apiVersion"),
// 			mergepatch.RequireKeyUnchanged("kind"), mergepatch.RequireMetadataKeyUnchanged("name")}
// 		patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current, preconditions...)
// 		if err != nil {
// 			if mergepatch.IsPreconditionFailed(err) {
// 				return nil, nil, fmt.Errorf("%s", "At least one of apiVersion, kind and name was changed")
// 			}
// 			return nil, nil, fmt.Errorf(createPatchErrFormat, original, modified, current, err)
// 		}
// 	case err != nil:
// 		return nil, nil, fmt.Errorf("getting instance of versioned object for %s. %v", p.Mapping.GroupVersionKind, err)
// 	case err == nil:
// 		// Compute a three way strategic merge patch to send to server.
// 		patchType = types.StrategicMergePatchType

// 		// Try to use openapi first if the openapi spec is available and can successfully calculate the patch.
// 		// Otherwise, fall back to baked-in types.
// 		// if p.OpenapiSchema != nil {
// 		// 	if schema = p.OpenapiSchema.LookupResource(p.Mapping.GroupVersionKind); schema != nil {
// 		// 		lookupPatchMeta = strategicpatch.PatchMetaFromOpenAPI{Schema: schema}
// 		// 		if openapiPatch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, p.Overwrite); err != nil {
// 		// 			fmt.Fprintf(errOut, "warning: error calculating patch from openapi spec: %v\n", err)
// 		// 		} else {
// 		// 			patchType = types.StrategicMergePatchType
// 		// 			patch = openapiPatch
// 		// 		}
// 		// 	}
// 		// }

// 		if patch == nil {
// 			lookupPatchMeta, err = strategicpatch.NewPatchMetaFromStruct(versionedObject)
// 			if err != nil {
// 				return nil, nil, fmt.Errorf(createPatchErrFormat, original, modified, current, err)
// 			}
// 			patch, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, p.Overwrite)
// 			if err != nil {
// 				return nil, nil, fmt.Errorf(createPatchErrFormat, original, modified, current, err)
// 			}
// 		}
// 	}

// 	if string(patch) == "{}" {
// 		return patch, obj, nil
// 	}

// 	options := metav1.UpdateOptions{}
// 	patchedObj, err := p.Helper.Patch(namespace, name, patchType, patch, &options)
// 	return patch, patchedObj, err
// }

// func (p *Patcher) deleteAndCreate(original runtime.Object, modified []byte, namespace, name string) ([]byte, runtime.Object, error) {
// 	if err := p.delete(namespace, name); err != nil {
// 		return modified, nil, err
// 	}
// 	// TODO: use wait
// 	if err := wait.PollImmediate(1*time.Second, p.Timeout, func() (bool, error) {
// 		if _, err := p.Helper.Get(namespace, name, false); !errors.IsNotFound(err) {
// 			return false, err
// 		}
// 		return true, nil
// 	}); err != nil {
// 		return modified, nil, err
// 	}
// 	versionedObject, _, err := unstructured.UnstructuredJSONScheme.Decode(modified, nil, nil)
// 	if err != nil {
// 		return modified, nil, err
// 	}
// 	options := metav1.CreateOptions{}
// 	createdObject, err := p.Helper.Create(namespace, true, versionedObject, &options)
// 	if err != nil {
// 		// restore the original object if we fail to create the new one
// 		// but still propagate and advertise error to user
// 		recreated, recreateErr := p.Helper.Create(namespace, true, original, &options)
// 		if recreateErr != nil {
// 			err = fmt.Errorf("an error occurred force-replacing the existing object with the newly provided one:\n\n%v.\n\nAdditionally, an error occurred attempting to restore the original object:\n\n%v", err, recreateErr)
// 		} else {
// 			createdObject = recreated
// 		}
// 	}
// 	return modified, createdObject, err
// }

// func (p *Patcher) delete(namespace, name string) error {
// 	mapping := p.Mapping
// 	cascade := p.Cascade
// 	gracePeriod := p.GracePeriod

// 	options := &metav1.DeleteOptions{}
// 	if gracePeriod >= 0 {
// 		options = metav1.NewDeleteOptions(int64(gracePeriod))
// 	}
// 	policy := metav1.DeletePropagationForeground
// 	if !cascade {
// 		policy = metav1.DeletePropagationOrphan
// 	}
// 	options.PropagationPolicy = &policy
// 	return p.DynamicClient.Resource(mapping.Resource).Namespace(namespace).Delete(name, options)
// }

// var metadataAccessor = meta.NewAccessor()

// func getModifiedConfiguration(obj runtime.Object, annotate bool, codec runtime.Encoder) ([]byte, error) {
// 	// First serialize the object without the annotation to prevent recursion,
// 	// then add that serialization to it as the annotation and serialize it again.
// 	var modified []byte

// 	// Otherwise, use the server side version of the object.
// 	// Get the current annotations from the object.
// 	annots, err := metadataAccessor.Annotations(obj)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if annots == nil {
// 		annots = map[string]string{}
// 	}

// 	original := annots[v1.LastAppliedConfigAnnotation]
// 	delete(annots, v1.LastAppliedConfigAnnotation)
// 	if err := metadataAccessor.SetAnnotations(obj, annots); err != nil {
// 		return nil, err
// 	}

// 	modified, err = runtime.Encode(codec, obj)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if annotate {
// 		annots[v1.LastAppliedConfigAnnotation] = string(modified)
// 		if err := metadataAccessor.SetAnnotations(obj, annots); err != nil {
// 			return nil, err
// 		}

// 		modified, err = runtime.Encode(codec, obj)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	// Restore the object to its original condition.
// 	annots[v1.LastAppliedConfigAnnotation] = original
// 	if err := metadataAccessor.SetAnnotations(obj, annots); err != nil {
// 		return nil, err
// 	}

// 	return modified, nil
// }
