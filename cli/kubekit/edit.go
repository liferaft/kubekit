package kubekit

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/liferaft/kubekit/cli"
	"github.com/liferaft/kubekit/pkg/kluster"
	"github.com/spf13/cobra"
)

const defaultEditor = "/usr/bin/vi"

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:     "edit [clusters-config] CLUSTER-NAME[,CLUSTER-NAME ...]",
	Aliases: []string{"e"},
	Short:   "Edits the configuration file for each given cluster name",
	Long: `Open an editor with the configuration file for each given cluster name. The
editor used will be taken from the flag '--editor' or from the environment 
variable 'KUBEKIT_EDITOR' or '/usr/bin/vi'.`,
	RunE: editClustersConfigRun,
}

// editClustersConfigCmd represents the 'edit clusters-config' command
var editClustersConfigCmd = &cobra.Command{
	Hidden:  true,
	Use:     "clusters-config CLUSTER-NAME[,CLUSTER-NAME ...]",
	Aliases: []string{"cc"},
	Short:   "Edits the configuration file for each given cluster name",
	Long: `Open an editor with the configuration file for each given cluster name. The
editor used will be taken from the flag '--editor' or from the environment 
variable 'KUBEKIT_EDITOR' or '/usr/bin/vi'.`,
	RunE: editClustersConfigRun,
}

func addEditCmd() {
	// edit [clusters-config] CLUSTER-NAME[,CLUSTER-NAME...] --editor FILE --read-only
	RootCmd.AddCommand(editCmd)
	editCmd.Flags().StringP("editor", "e", "", "editor to use. If not provided will use it from $KUBEKIT_EDITOR or /usr/bin/vi")
	editCmd.Flags().BoolP("read-only", "r", false, "don't edit, just send to stdout the cluster configuration file")
	editCmd.AddCommand(editClustersConfigCmd)
	editClustersConfigCmd.Flags().StringP("editor", "e", "", "editor to use. If not provided will use it from $KUBEKIT_EDITOR or /usr/bin/vi")
	editClustersConfigCmd.Flags().BoolP("read-only", "r", false, "don't edit, just send to stdout the cluster configuration file")
}

func editClustersConfigRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cli.UserErrorf("requires a cluster name")
	}
	clustersName := args

	editor := cmd.Flags().Lookup("editor").Value.String()
	if len(editor) == 0 {
		editor = os.Getenv("KUBEKIT_EDITOR")
	}
	if len(editor) == 0 {
		editor = defaultEditor
	}

	ro := cmd.Flags().Lookup("read-only").Value.String() == "true"

	// DEBUG:
	// var roFlag string
	// if ro {
	// 	roFlag = " --read-only"
	// }
	// fmt.Printf("edit clusters %s --editor %v %s\n", strings.Join(clustersName, " "), editor, roFlag)

	return editClusters(editor, ro, clustersName...)
}

func editClusters(editor string, readOnly bool, clustersName ...string) error {
	if _, err := os.Stat(editor); os.IsNotExist(err) {
		return cli.UserErrorf("not found editor %q", editor)
	}

	klusterList, err := kluster.List(config.ClustersDir(), clustersName...)
	if err != nil {
		return err
	}
	if len(klusterList) == 0 || klusterList == nil {
		return fmt.Errorf("cluster configuration file for %v was not found", clustersName)
	}

	type editError struct {
		name string
		err  error
	}
	errs := []editError{}
	for _, k := range klusterList {
		if k == nil {
			// Shouldn't be nil, but just in case
			continue
		}
		if err := openFile(editor, k.Path(), readOnly); err != nil {
			e := editError{
				name: k.Name,
				err:  err,
			}
			errs = append(errs, e)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return fmt.Errorf("failed to edit the custer %s, %s", errs[0].name, errs[0].err)
	}

	errStr := "failed to edit custers:"
	for _, err := range errs {
		errStr = fmt.Sprintf("%s\nname: %s\terror: %s", errStr, err.name, err.err)
	}
	return fmt.Errorf(errStr)
}

func openFile(editor, file string, readOnly bool) error {
	if readOnly {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		fmt.Print(string(content))
		return nil
	}
	cmd := exec.Command(editor, file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
