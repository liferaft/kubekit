package kubekitctl

import (
	"fmt"
	"os"

	"github.com/johandry/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kubekitctl",
	Short: "KubeKit client",
	Long:  `It's a client to interact with the KubeKit server using the gRPC or REST API.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	RootCmd.SilenceUsage = true
	RootCmd.SilenceErrors = true

	addCommands()

	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n\x1B[91;1m[ERROR]\x1B[0m %s\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	v := viper.New()

	if cfgFile != "" {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kubekitctl" (without extension).
		v.AddConfigPath(home)
		v.AddConfigPath(".")
		v.SetConfigName(".kubekitctl")
	}

	v.AutomaticEnv() // read in environment variables that match
	v.SetEnvPrefix("KUBEKITCTL")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	setDefaultAndBindPFlag(v, RootCmd.PersistentFlags().Lookup("debug"), defDebug)
	setDefaultAndBindPFlag(v, RootCmd.PersistentFlags().Lookup("host"), defHost)
	setDefaultAndBindPFlag(v, RootCmd.PersistentFlags().Lookup("no-grpc"), defNoGRPC)
	setDefaultAndBindPFlag(v, RootCmd.PersistentFlags().Lookup("no-http"), defNoHTTP)
	setDefaultAndBindPFlag(v, RootCmd.PersistentFlags().Lookup("insecure"), defInsecure)
	setDefaultAndBindPFlag(v, RootCmd.PersistentFlags().Lookup("port"), defPort)

	for _, name := range []string{
		"grpc-port",
		"healthz-port",
		"cert-dir",
		"tls-cert-file",
		"tls-private-key-file",
		"ca-file",
	} {
		setDefaultAndBindPFlag(v, RootCmd.PersistentFlags().Lookup(name), "")
	}

	v.Set(log.PrefixField, "kubekitctl")
	if v.GetBool("debug") {
		v.Set(log.LevelKey, "debug")
	} else if v.GetBool("quiet") {
		v.Set(log.LevelKey, "error")
	} else {
		v.Set(log.LevelKey, "info")
	}

	v.Set(log.OutputKey, os.Stderr)

	logger := log.New(v)

	config = &Config{
		Logger:      logger,
		Debug:       v.GetBool("debug"),
		Insecure:    v.GetBool("insecure"),
		Host:        v.GetString("host"),
		Port:        v.GetString("port"),
		PortGRPC:    v.GetString("grpc-port"),
		PortHealthz: v.GetString("healthz-port"),
		NoHTTP:      v.GetBool("no-http"),
		NoGRPC:      v.GetBool("no-grpc"),
		CertDir:     v.GetString("cert-dir"),
		CertFile:    v.GetString("tls-cert-file"),
		KeyFile:     v.GetString("tls-private-key-file"),
		CAFile:      v.GetString("ca-file"),
	}
	// v.Unmarshal(config)

	// DEBUG:
	// fmt.Printf("[DEBUG] debug: %t\n", config.Debug)
	// fmt.Printf("[DEBUG] no-grpc: %t\n", config.NoGRPC)
	// fmt.Printf("[DEBUG] no-http: %t\n", config.NoHTTP)
	// fmt.Printf("[DEBUG] insecure: %t\n", config.Insecure)
	// fmt.Printf("[DEBUG] port: %s\n", config.Port)
	// fmt.Printf("[DEBUG] grpc-port: %s\n", config.PortGRPC)
	// fmt.Printf("[DEBUG] healthz-port: %s\n", config.PortHealthz)
	// fmt.Printf("[DEBUG] cert-dir: %s\n", config.CertDir)
	// fmt.Printf("[DEBUG] tls-cert-file: %s\n", config.CertFile)
	// fmt.Printf("[DEBUG] tls-private-key-file: %s\n", config.KeyFile)
	// fmt.Printf("[DEBUG] ca-file: %s\n", config.CAFile)

	if err := config.init(); err != nil {
		logger.Error(err.Error())
		os.Exit(2)
	}
}

func setDefaultAndBindPFlag(v *viper.Viper, f *pflag.Flag, value interface{}) {
	v.SetDefault(f.Name, value)
	if f != nil {
		v.BindPFlag(f.Name, f)
	}
}

func addCommands() {
	// <any command> --config [FILE] --debug --no-grpc --no-http --insecure --port [PORT] --grpc-port [PORT] --healthz-port [PORT] --cert-dir [PATH] --tls-cert-file [FILE] --tls-private-key-file [FILE] --ca-file [FILE]
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kubekitctl.yaml)")
	RootCmd.PersistentFlags().Bool("debug", defDebug, "debug output useful for develpment")
	RootCmd.PersistentFlags().String("host", defHost, "hostname or IP address used by the server")
	RootCmd.PersistentFlags().Bool("no-grpc", defNoGRPC, "do not check gRPC, only REST/HTTP")
	RootCmd.PersistentFlags().Bool("no-http", defNoHTTP, "do not check REST/HTTP, only gRPC")
	RootCmd.PersistentFlags().Bool("insecure", defInsecure, "connect to the server not using the TLS")
	RootCmd.PersistentFlags().String("port", defPort, "port used by the server")
	RootCmd.PersistentFlags().String("grpc-port", "", "GRPC port used by the server")
	RootCmd.PersistentFlags().String("healthz-port", "", "GRPC port used by the server")
	RootCmd.PersistentFlags().String("cert-dir", "", "port used by the server")
	RootCmd.PersistentFlags().String("tls-cert-file", "", "port used by the server")
	RootCmd.PersistentFlags().String("tls-private-key-file", "", "port used by the server")
	RootCmd.PersistentFlags().String("ca-file", "", "port used by the server")

	healthzAddCommands()
	versionAddCommands()
	tokenAddCommands()
	initAddCommands()
	applyAddCommands()
	deleteAddCommands()
	describeAddCommands()
	getAddCommands()
	updateAddCommands()
}
