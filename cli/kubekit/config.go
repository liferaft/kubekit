package kubekit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/johandry/log"
	"github.com/kraken/ui"
	"github.com/liferaft/kubekit/cli"
	homedir "github.com/mitchellh/go-homedir"
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	defCfgFilename    = "config"
	defCfgFileFormat  = "json"
	defKfgFileFormat  = "yaml"
	defScroll         = false
	defVerbose        = true
	defQuiet          = false
	defDebug          = false
	defLogLevel       = "info"
	defLogForceColors = true
	defLogFile        = "" // default log output to os.Stderr
	defServerPKIDir   = "server/pki"
	defClustersDir    = "clusters"
	defTemplatesDir   = "templates"
	defKubeKitHomeDir = ".kubekit.d"
)

const (
	defServerHost  = "0.0.0.0"
	defServerPort  = 5823
	defHealthzPort = 5823
)

const (
	logRootPrefix = "kubekit"
	envPrefix     = "kubekit"
)

// Config is the KubeKit configuration that are obtained from all the sub-commands
type Config struct {
	cfgFilename    string
	UI             *ui.UI `json:"-" yaml:"-" toml:"-" mapstructure:"-"`
	Scroll         bool   `json:"scroll" yaml:"scroll" toml:"scroll" mapstructure:"scroll"`
	Verbose        bool   `json:"verbose" yaml:"verbose" toml:"verbose" mapstructure:"verbose"`
	Quiet          bool   `json:"quiet" yaml:"quiet" toml:"quiet" mapstructure:"quiet"`
	Debug          bool   `json:"debug" yaml:"debug" toml:"debug" mapstructure:"debug"`
	LogLevel       string `json:"log_level" yaml:"log_level" toml:"log_level" mapstructure:"log_level"`
	LogForceColors bool   `json:"log_color" yaml:"log_color" toml:"log_color" mapstructure:"log_color"`
	LogFile        string `json:"log" yaml:"log" toml:"log" mapstructure:"log"`
	ClustersPath   string `json:"clusters_path" yaml:"clusters_path" toml:"clusters_path" mapstructure:"clusters_path"`
	TemplatesPath  string `json:"templates_path" yaml:"templates_path" toml:"templates_path" mapstructure:"templates_path"`
	PKIPath        string `json:"pki_path" yaml:"pki_path" toml:"pki_path" mapstructure:"pki_path"`

	// Keep viper and command just in case a parameter is missing or to compare them
	// Remove them when no needed anymore.
	viper   *viper.Viper
	command *cobra.Command
}

var config *Config

// configCmd represents the config command
var showConfigCmd = &cobra.Command{
	Hidden: true,
	Use:    "show-config",
	Short:  "Show KubeKit configuration",
	Long: `List all the configuration settings from the flags, environment variables and 
configuration file. It may also save all the settings to a new or existing 
configuration file`,
	RunE: showConfigRun,
}

func addShowConfigCmd() {
	// show-config --output (wide|json|yaml|toml) --pp
	RootCmd.AddCommand(showConfigCmd)
	showConfigCmd.Flags().StringP("output", "o", defCfgFileFormat, "Format to show the configuration. Available formats: json, yaml, toml and text (default "+defCfgFileFormat+")")
	showConfigCmd.Flags().BoolP("pp", "p", false, "Pretty print. Show the configuration in a human readable format. Applies only for json and text formats. (defaults to false)")
	showConfigCmd.Flags().String("to", "", "file to save the configuration file")
}

func showConfigRun(cmd *cobra.Command, args []string) error {
	pp := cmd.Flags().Lookup("pp").Value.String() == "true"
	format := cmd.Flags().Lookup("output").Value.String()

	configStr, err := config.Stringf(format, pp)
	if err != nil {
		return err
	}

	filename := cmd.Flags().Lookup("to").Value.String()
	if len(filename) != 0 {
		return ioutil.WriteFile(filename, []byte(configStr), 0644)
	}
	fmt.Println(configStr)
	return nil
}

// NewConfig create a new Config with default values
func NewConfig(command *cobra.Command) *Config {
	config := Config{
		cfgFilename:    defCfgFilename,
		UI:             nil,
		Verbose:        defVerbose,
		Quiet:          defQuiet,
		Debug:          defDebug,
		LogLevel:       defLogLevel,
		LogForceColors: defLogForceColors,
		LogFile:        defLogFile,
		viper:          nil,
		command:        command,
	}

	return &config
}

// UpdateViper refresh all the config settings from viper keys
func (c *Config) UpdateViper(v *viper.Viper) {
	v.Unmarshal(c)

	c.cfgFilename = v.ConfigFileUsed()
	c.LogFile = v.GetString(log.FilenameKey)

	c.viper = v
}

// Dir returns the absolute directory path of the configuration file
func (c *Config) Dir() string {
	return dir(c.cfgFilename)
}

// dir returns the absolute directory path of given path joint with the current dir
func dir(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Dir(path)
	}
	pwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("cannot find local working directory. %s", err))
	}
	p := filepath.Join(pwd, filepath.Dir(path))
	pathAbs, err := filepath.Abs(p)
	if err != nil {
		panic(fmt.Errorf("cannot create absolute path for directory %s. %s", p, err))
	}
	return pathAbs
}

// ClustersDir returns the absolute directory path where the clusters config are
func (c *Config) ClustersDir() string {
	return absDir(c.Dir(), c.ClustersPath)
}

// TemplatesDir returns the absolute directory path where the templates are
func (c *Config) TemplatesDir() string {
	return absDir(c.Dir(), c.TemplatesPath)
}

// PKIDir returns the KubeKit PKI directory. Here is where the server store the
// certificates
func (c *Config) PKIDir() string {
	return absDir(c.Dir(), c.PKIPath)
}

func absDir(baseDir, path string) string {
	if filepath.IsAbs(path) {
		return path
		// return filepath.Dir(path)
	}

	p := filepath.Join(baseDir, path)
	pathAbs, err := filepath.Abs(p)
	if err != nil {
		panic(fmt.Errorf("cannot create absolute path for directory %s. %s", p, err))
	}
	return pathAbs
}

// Stringf returns the configuration in the requested format and alows to choose
// pretty print if it's JSON. If no format is provided will return it in the
// format defined in cfgShowFormat or JSON
func (c *Config) Stringf(format string, pp bool) (string, error) {
	if len(format) == 0 {
		format = defCfgFileFormat
	}
	switch format {
	case "json":
		return c.JSON(pp)
	case "yaml", "yml":
		return c.YAML()
	case "toml", "tml":
		return c.TOML()
	case "text", "txt":
		return c.Text(pp)
	default:
		return "", cli.UserErrorf("unknown format %s", format)
	}
}

// Stringf returns the configuration in the format defined in cfgShowFormat and
// pretty print if it's JSON
func (c *Config) String() string {
	str, err := c.Stringf(defCfgFileFormat, true)
	if err != nil {
		return err.Error()
	}
	return str
}

// JSON return the configuration in JSON format
func (c *Config) JSON(pp bool) (string, error) {
	var b []byte
	var err error

	if pp {
		b, err = json.MarshalIndent(c, "", "  ")
	} else {
		b, err = json.Marshal(c)
	}

	return string(b), err
}

// YAML return the configuration in YAML format
func (c *Config) YAML() (string, error) {
	b, err := yaml.Marshal(c)
	return string(b), err
}

// TOML return the configuration in TOML format
func (c *Config) TOML() (string, error) {
	b, err := toml.Marshal(c)
	return string(b), err
}

// Text return the configuration in human readable format with more information
// and even more if debug flag is set
func (c *Config) Text(pp bool) (string, error) {
	var b bytes.Buffer

	fmt.Fprintf(&b, "Config Filename:\t%s\n", c.cfgFilename)
	fmt.Fprintf(&b, "Config Path:\t%s\n", c.Dir())
	fmt.Fprintf(&b, "Scroll:\t\t%t\n", c.Scroll)
	fmt.Fprintf(&b, "Verbose:\t\t%t\n", c.Verbose)
	fmt.Fprintf(&b, "Quiet:\t\t%t\n", c.Quiet)
	fmt.Fprintf(&b, "Debug:\t\t\t%t\n", c.Debug)
	fmt.Fprintf(&b, "Log Level:\t\t%s\n", c.LogLevel)
	fmt.Fprintf(&b, "Log Force Colors:\t%t\n", c.LogForceColors)
	fmt.Fprintf(&b, "Log File:\t\t%s\n", c.LogFile)
	fmt.Fprintf(&b, "Log Prefix:\t\t%s\n", c.UI.Log.GetPrefix())

	if c.command.Flags().Lookup("debug").Value.String() == "true" {
		b.WriteString(c.debug())
	}

	return b.String(), nil
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set. It's executed before
// a command Run() function is executed.
func initConfig() {
	if err := initConfigE(); err != nil {
		errMsg := fmt.Sprintf("Failed to load the configuration from %s. %s", cfgFile, err)
		if config.UI == nil || config.UI.Log == nil {
			fmt.Fprintf(os.Stderr, "\x1B[91;1m[ERROR]\x1B[0m %s", errMsg)
			os.Exit(1)
		}
		config.UI.Log.Fatal(errMsg)
	}
}

func initConfigE() error {
	config = NewConfig(RootCmd)

	v, err := LoadConfig(cfgFile)
	if err != nil {
		return err
	}
	setDefaultsAndBindPFlags(v)

	ui := initUI(v)

	config.UpdateViper(v)

	config.UI = ui

	return nil
}

// LoadConfig loads the configuration file into a Viper object
func LoadConfig(cfgFile string) (*viper.Viper, error) {
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvPrefix(envPrefix)

	if newCfgFile, ok := configFile(cfgFile); ok {
		v.SetConfigFile(newCfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		envHome := os.Getenv(strings.ToUpper(envPrefix) + "_HOME")

		if len(envHome) != 0 {
			v.AddConfigPath(envHome)
		}
		v.AddConfigPath(filepath.Join(home, defKubeKitHomeDir))
		v.AddConfigPath(".")

		v.SetConfigName(defCfgFilename)
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigParseError); ok {
			return nil, err
		}
		// Do not return if error is not 'ConfigParseError', that means the file was
		// not found and it is OK if there is no config file.
		// return nil, fmt.Errorf("cannot locate config file (%s). %s", filename, err)
	}

	return v, nil
}

func configFile(cfgFile string) (string, bool) {
	if len(cfgFile) != 0 {
		return cfgFile, true
	}

	envCfgFile := os.Getenv(strings.ToUpper(envPrefix) + "_CONFIG")
	if len(envCfgFile) != 0 {
		return envCfgFile, true
	}

	envHome := os.Getenv(strings.ToUpper(envPrefix) + "_HOME")
	if len(envHome) != 0 {
		return filepath.Join(envHome, defCfgFilename+"."+defCfgFileFormat), true
	}

	return "", false
}

func setDefaultsAndBindPFlags(v *viper.Viper) {
	// Root cmd persistent flags:
	setDefaultAndBindPFlag(v, RootCmd.PersistentFlags().Lookup("scroll"), defScroll)
	setDefaultAndBindPFlag(v, RootCmd.PersistentFlags().Lookup("verbose"), defVerbose)
	setDefaultAndBindPFlag(v, RootCmd.PersistentFlags().Lookup("quiet"), defQuiet)
	setDefaultAndBindPFlag(v, RootCmd.PersistentFlags().Lookup("debug"), defDebug)
	setDefaultAndBindPFlag(v, RootCmd.PersistentFlags().Lookup("log"), defLogFile)

	// Logging defaults, doesn't have flags, so not binded:
	v.SetDefault(log.LevelKey, defLogLevel)
	v.SetDefault(log.ForceColorsKey, defLogForceColors)
	v.SetDefault(log.FilenameKey, defLogFile)

	var kubekitHomeDir string
	envHome := os.Getenv(strings.ToUpper(envPrefix) + "_HOME")
	if len(envHome) != 0 {
		kubekitHomeDir = envHome
	} else {
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}
		kubekitHomeDir = filepath.Join(home, defKubeKitHomeDir)
	}

	v.SetDefault("clusters_path", filepath.Join(kubekitHomeDir, defClustersDir))
	v.SetDefault("templates_path", filepath.Join(kubekitHomeDir, defTemplatesDir))
	v.SetDefault("pki_path", filepath.Join(kubekitHomeDir, defServerPKIDir))
}

func setDefaultAndBindPFlag(v *viper.Viper, f *pflag.Flag, value interface{}) {
	v.SetDefault(f.Name, value)
	if f != nil {
		v.BindPFlag(f.Name, f)
	}
}

func initUI(v *viper.Viper) *ui.UI {
	v.Set(log.PrefixField, logRootPrefix)

	// 'debug', 'quiet' or 'verbose' have priority over 'log_level'
	if v.GetBool("quiet") {
		v.Set(log.LevelKey, "error")
	} else if v.GetBool("debug") {
		v.Set(log.LevelKey, "debug")
	} else if v.GetBool("verbose") {
		v.Set(log.LevelKey, "info")
	}

	// If 'log' is a file logs will go to that file.
	// Otherwise will go to the default log file or Stdout
	if v.IsSet("log") && v.GetString("log") != "" {
		logFile := v.GetString("log")
		logFilePath := absDir(dir(v.ConfigFileUsed()), logFile)
		v.Set(log.FilenameKey, logFilePath)
	} else {
		v.Set(log.OutputKey, os.Stderr)
	}

	l := log.New(v)

	return ui.New(v.GetBool("scroll"), l)
}
