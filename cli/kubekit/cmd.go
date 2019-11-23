package kubekit

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/liferaft/kubekit/pkg/kluster"
)

var (
	cfgFile     string
	versionFlag bool
	verboseFlag bool
)

// Execute adds all child commands to the root command.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	RootCmd.SilenceUsage = true
	RootCmd.SilenceErrors = true

	AddCommands()

	if c, err := RootCmd.ExecuteC(); err != nil {
		if config != nil && config.UI != nil && config.UI.Log != nil {
			config.UI.Log.Error(err.Error())
		}
		fmt.Fprintf(os.Stderr, "\n\x1B[91;1m[ERROR]\x1B[0m %s\n", err)
		c.Printf("\n%s\n", c.UsageString())
		os.Exit(-1)
	}
}

// AddCommands adds child commands to the root command
func AddCommands() {
	// <any command> --config [FILE] --log [FILE] --verbose --quiet --debug --scroll
	initPersistentFlags()

	// init [cluster] NAME --platform NAME --path PATH --format FORMAT --template NAME --update
	// init template NAME  --platform NAME --path PATH --format FORMAT --update
	// init certificates CLUSTER-NAME --CERT-key-file FILE --CERT-cert-file FILE
	addInitCmd()

	// apply [cluster] NAME --provision --configure --certificates --generate-certs --export-tf --export-k8s --plan --CERT-key-file FILE --CERT-cert-file FILE
	addApplyCmd()

	// delete [cluster] NAME --force --all
	// delete clusters-config NAME --force
	// delete templates NAME[,NAME...] --force
	// delete files CLUSTER-NAME FILE[,FILE...] --force --nodes NODE[,NODE] --pools POOL[,POOL]
	addDeleteCmd()

	// edit [clusters-config] CLUSTER-NAME[,CLUSTER-NAME...] --editor FILE --read-only
	addEditCmd()

	// [get] clusters NAME[,NAME...] --output (wide|json|yaml|toml) --pp
	// [get] nodes CLUSTER-NAME NAME[,NAME...] --output (wide|json|yaml|toml) --pp --nodes NODE[,NODE] --pools POOL[,POOL]
	// [get] files CLUSTER-NAME FILENAME[,FILENAME...] --output (wide|json|yaml|toml) --pp --nodes NODE[,NODE] --pools POOL[,POOL] --path PATHT[,PATH]
	// [get] templates NAME[,NAME...] --output (wide|json|yaml|toml) --pp
	addGetCmd()

	// copy [cluster] NAME --to NEW-NAME --provision --configure --certificates --generate-certs --export --plan --CERT-key-file FILE --CERT-cert-file FILE
	// copy cluster-config CLUSTER-NAME --to NEW-NAME --platform NAME --path PATH --template NAME
	// copy template NAME
	// copy files
	// copy certificates
	addCopyCmd()

	// exec [cluster] NAME
	addExecCmd()

	// login [cluster] NAME --platform platform --list --access_key aws_access_key_id --secret_key aws_secret_access_key --region aws_default_region --server server_ip_or_dns --username username --password password
	// login node NAME --cluster NAME
	addLoginCmd()

	// describe [cluster] NAME[,NAME ...] --output (json|yaml|toml) --pp
	// describe templates NAME[,NAME ...] --output (json|yaml|toml) --pp
	// describe nodes CLUSTER-NAME --output (json|yaml|toml) --pp
	addDescribeCmd()

	// start [cluster] NAME[,NAME ...]
	// stop [cluster] NAME[,NAME ...]
	// restart [cluster] NAME[,NAME ...]
	addStartStopCmd()

	// scale [cluster] NAME POOL-NAME=[+|-]N
	addScaleCmd()

	// --version
	// version
	addVersionCmd()

	// show-config --output (wide|json|yaml|toml) --pp
	addShowConfigCmd()

	// token --cluster-id NAME
	addTokenCmd()
}

// initPersistentFlags set global flags, these flags will be available to the
// root command as well as every subcommand
func initPersistentFlags() {
	// <any command> --config X --log X --verbose --quiet --debug --scroll
	// TODO: Add '--no-color' flag

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "", "", "config file (default "+defCfgFilename+".<yaml|json|toml> at ~/"+defKubeKitHomeDir+"/ or ./)")
	RootCmd.PersistentFlags().SetAnnotation("config", cobra.BashCompFilenameExt, []string{"json", "yaml", "yml", "toml", "tml"})

	RootCmd.PersistentFlags().Bool("scroll", defScroll, "scroll output")
	RootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", defVerbose, "verbose output")
	RootCmd.PersistentFlags().BoolP("quiet", "q", defQuiet, "quiet or no output")
	RootCmd.PersistentFlags().Bool("debug", defDebug, "debug output useful for develpment")
	RootCmd.PersistentFlags().MarkHidden("debug")

	RootCmd.PersistentFlags().StringP("log", "l", defLogFile, "log file (default is Stderr)")
	RootCmd.PersistentFlags().SetAnnotation("log", cobra.BashCompFilenameExt, []string{})
}

func addCertFlags(command *cobra.Command) {
	for _, caCertInfo := range kluster.CACertNames {
		if len(caCertInfo.Desc) == 0 {
			continue
		}
		command.Flags().String(caCertInfo.CN+"-key-file", "", "CA RSA Key file "+caCertInfo.Desc+", recommended for production.")
		command.Flags().String(caCertInfo.CN+"-cert-file", "", "CA x509 Certificate file "+caCertInfo.Desc+", recommended for production.")
	}
}
