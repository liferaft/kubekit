package resources

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/cli-runtime/pkg/resource"
)

type applyError struct {
	resource string
	err      error
}
type applyErrors []applyError

func (e applyError) String() string {
	return fmt.Sprintf("%s: %v", e.resource, e.err)
}

func (ae applyErrors) Add(res string, err error) {
	e := applyError{
		resource: res,
		err:      err,
	}
	ae = append(ae, e)
}

func (ae applyErrors) Empty() bool {
	return len(ae) == 0
}

func (ae applyErrors) String() string {
	str := []string{}
	for _, e := range ae {
		str = append(str, e.String())
	}
	return strings.Join(str, "\n\t")
}

func (ae applyErrors) Error() string {
	return fmt.Sprintf("creating the following resources: %s\n", ae.String())
}

// ApplyAll mimic the `kubectl apply` command to create or update all the resources in the list
func (r *Resources) ApplyAll() error {
	errors := applyErrors{}

	// DEBUG:
	// r.ui.Log.Debugf("the following resources will be applied: %v", r.content)
	for _, res := range r.Names() {
		if _, ok := r.content[res]; !ok {
			r.ui.Log.Warnf("resource content for %q not found", res)
			continue
		}
		if err := r.Apply(res); err != nil {
			r.ui.Log.Errorf("failed applying resource %s. %v", res, err)
			errors.Add(res, err)
		}
	}

	if errors.Empty() {
		return nil
	}

	r.ui.Log.Errorf(errors.Error())

	return fmt.Errorf(errors.Error())
	// return errors
}

// Apply applies the given resource into the Kubernetes cluster
func (r *Resources) Apply(name string) (err error) {
	var result *resource.Result
	for l := 0; l < 6; time.Sleep(10 * time.Second) {
		l++
		switch {
		case isFile(name):
			if !isURL(name) {
				name = strings.TrimPrefix(name, "file://")
			}
			filenames := []string{name}
			result = r.kubeClient.ResultForFilenameParam(filenames, true)

		default:
			resContent, err := r.Render(name, "")
			if err != nil {
				return fmt.Errorf("failed rendering resource. %v", err)
			}
			result = r.kubeClient.ResultForContent(name, resContent, true)
		}

		err = r.kubeClient.ApplyResource(result)
		if err != nil {
			r.ui.Log.Debugf("received error during validation of apply, attempting retry: %s,", err)
			continue
		}

		exists, err := r.kubeClient.ExistsResource(result)
		if exists {
			// Resources applied and successfully found
			return nil
		}
		if err != nil {
			r.ui.Log.Debugf("received error during validation of apply, attempting retry. %s,", err)
		} else {
			r.ui.Log.Debugf("resource applied but not found, attempting re-apply. %s,", err)
		}
	}
	return err
}
