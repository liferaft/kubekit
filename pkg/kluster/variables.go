package kluster

import (
	"fmt"

	"github.com/johandry/merger"
)

// ConfigVariables returns a map of string with the configuration in form of
// kubekit input variables
func (k *Kluster) ConfigVariables() (vars map[string]string, err error) {
	// config, err := k.Config.Map()
	// if err != nil {
	// 	return nil, err
	// }

	vars = map[string]string{}
	same := []string{}

	if k.Config != nil {
		// k.ui.Log.Debugf("getting configuration variables for cluster %q", k.Name)
		if vars, err = merger.TransformToMap(k.Config, "mapstructure", "yaml", "json"); err != nil {
			return nil, err
		}
	}

	platform := k.Platform()
	var pConfig interface{}
	if p, ok := k.provisioner[platform]; ok && p != nil {
		pConfig = p.Config()
	}

	if pConfig != nil {
		// k.ui.Log.Debugf("getting provisioning variables for cluster %q", k.Name)
		pMap := map[string]string{}
		if pMap, err = merger.TransformToMap(pConfig, "mapstructure", "yaml", "json"); err != nil {
			return nil, err
		}

		for k, v := range pMap {
			if _, ok := vars[k]; ok {
				same = append(same, k)
			}
			vars[k] = v
		}
	}

	if len(same) != 0 {
		err = fmt.Errorf("same parameters in provisioner and configurator: %v", same)
	}

	return vars, err
}
