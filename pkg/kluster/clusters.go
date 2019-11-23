package kluster

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kraken/ui"
	uuid "github.com/nu7hatch/gouuid"
)

// List returns a list of Klusters existing in the given directory
func List(path string, clustersName ...string) ([]*Kluster, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("cluster path not found (%s)", path)
	}
	klusterList := make([]*Kluster, 0)

	listAll := len(clustersName) == 0

	uuidDirs, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, uuidDir := range uuidDirs {
		clusterDir, err := ioutil.ReadDir(filepath.Join(path, uuidDir.Name()))
		if err != nil {
			continue
		}
		// Find the files that begins with the default configuration filename, it
		// may have different extensions (yaml, yml, json, ...)
		var clusterFileName string
		for _, f := range clusterDir {
			if strings.HasPrefix(f.Name(), DefaultConfigFilename) {
				clusterFileName = filepath.Join(path, uuidDir.Name(), f.Name())
				break
			}
		}
		if len(clusterFileName) == 0 {
			continue
		}

		// If no error loading the cluster, add it to the list
		if klusterData, err := LoadSummary(clusterFileName); err == nil {
			if listAll || isIn(klusterData.Name, clustersName...) {
				klusterList = append(klusterList, klusterData)
			}
		} else {
			fmt.Printf("ERROR with the configuration: %s : %s\n", clusterFileName, err)
		}
	}
	return klusterList, nil
}

func isIn(text string, list ...string) bool {
	for _, elem := range list {
		if text == elem {
			return true
		}
	}
	return false
}

// ListNames returns the list of Klusters name existing in the given directory
func ListNames(path string) ([]string, error) {
	nameList := make([]string, 0)

	klusterList, err := List(path)
	if err != nil {
		return nameList, err
	}

	for _, k := range klusterList {
		nameList = append(nameList, k.Name)
	}

	return nameList, nil
}

// Unique checks if the given cluster name does not exists (it's unique) in the
// given cluster path.
func Unique(clusterName, path string) bool {
	klusterNames, _ := ListNames(path)
	for _, name := range klusterNames {
		if clusterName == name {
			return false
		}
	}
	return true
}

// Path returns the path for a given cluster name that should be locates in the
// given clusters path. Returns an empty string if not found
func Path(clusterName, path string) string {
	klusters, _ := List(path)
	for _, kluster := range klusters {
		if clusterName == kluster.Name {
			return kluster.path
		}
	}
	return ""
}

// NewPath generates a path for a new cluster
func NewPath(clustersPath string) (string, error) {
	var path string
	retry := 0

	// Retry 3 times if the path with the uuid exists (weird scenario)
	for {
		uuid, err := uuid.NewV4()
		if err != nil {
			return path, fmt.Errorf("failed to generate the UUID to create the Kluster config file. %s", err)
		}

		path = filepath.Join(clustersPath, uuid.String())
		if _, err := os.Stat(path); os.IsExist(err) {
			retry++
			if retry < 3 {
				continue
			}
			// If exists and tried 3 times, return error
			return path, fmt.Errorf("the directory %s already exists, failed to generate a unique UUID. %s", path, err)
		}
		// If does not exists, create and return it
		os.MkdirAll(path, 0755)
		return path, nil
	}
}

// LoadCluster return the cluster loacated in the clustersPath with name clusterName
func LoadCluster(clusterName, clustersPath string, ui *ui.UI) (*Kluster, error) {
	klusterFile := Path(clusterName, clustersPath)
	if len(klusterFile) == 0 {
		return nil, fmt.Errorf("failed to find the cluster named %q", clusterName)
	}
	cluster, err := Load(klusterFile, ui)
	if err != nil {
		return nil, fmt.Errorf("failed to load the kluster config file %s. %s", klusterFile, err)
	}

	return cluster, nil
}

// ValidClusterName return a valid a cluster name. If the cluster name changed
// the error contain the changes, if it's not possible to fix returns the error
// and no name
func ValidClusterName(clusterName string) (string, error) {
	if len(clusterName) == 0 {
		return "", fmt.Errorf("cluster name cannot be empty")
	}

	var err error

	newName := strings.ToLower(clusterName)
	if newName != clusterName {
		err = fmt.Errorf("cluster cannot have uppercase characters")
	}
	clusterName = newName

	if strings.Contains(clusterName, "_") {
		clusterName = strings.Replace(clusterName, "_", "-", -1)
		if err != nil {
			err = fmt.Errorf("cluster name cannot contain uppercase characters or underscore ('_')")
		} else {
			err = fmt.Errorf("cluster name cannot contain underscore ('_')")
		}
	}

	reAcceptedClusterName := regexp.MustCompile(`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*)$`)
	if !reAcceptedClusterName.MatchString(clusterName) {
		return "", fmt.Errorf("cluster name must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
	}

	return clusterName, err
}

// CreateCluster creates a cluster and cluster configuration file
func CreateCluster(clusterName, platform, path, format string, variables map[string]string, parentUI *ui.UI) (cluster *Kluster, err error) {
	// Check the cluster name is valid and unique
	if clusterName, err = ValidClusterName(clusterName); err != nil && clusterName == "" {
		return nil, err
	}

	if ok := Unique(clusterName, path); !ok {
		return nil, fmt.Errorf("cluster name %q already exists", clusterName)
	}

	// Get a path with the UUID to store the cluster
	if path, err = NewPath(path); err != nil {
		return nil, err
	}

	// Create the cluster and ...
	if cluster, err = New(clusterName, platform, path, format, parentUI, variables); err != nil {
		return nil, fmt.Errorf("failed to initialize the cluster %s. %s", clusterName, err)
	}

	// ... save it
	if err = cluster.Save(); err != nil {
		return nil, err
	}

	return cluster, nil
}
