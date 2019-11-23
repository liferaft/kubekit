package ssh

import (
	"bytes"
	"encoding/json"
	"sync"

	toml "github.com/pelletier/go-toml"
	yaml "gopkg.in/yaml.v2"
)

// Command encapsulate a remote command
type Command struct {
	Command    string
	Stdin      bytes.Buffer
	Stdout     bytes.Buffer
	Stderr     bytes.Buffer
	ExitStatus int
}

// HostCommandResult encapsulate the results of a command from one host
type HostCommandResult struct {
	Stdout     string `json:"stdout" yaml:"stdout" toml:"stdout" mapstructure:"stdout"`
	Stderr     string `json:"stderr" yaml:"stderr" toml:"stderr" mapstructure:"stderr"`
	ExitStatus int    `json:"exitstatus" yaml:"exitstatus" toml:"exitstatus" mapstructure:"exitstatus"`
}

// HostCommandResultMap is a type safe map wrapped with mutex locks around get/set calls
// where the keys are the hosts and values are HostCommandResults
type HostCommandResultMap struct {
	sync.RWMutex `json:"-" yaml:"-" toml:"-" mapstructure:"-"`
	Results      map[string]*HostCommandResult `json:"hosts" yaml:"hosts" toml:"hosts" mapstructure:"hosts"` // this was set to an exportable field to allow for easier marshaling
}

// CommandResult encapsulate the results of a command execution per host
type CommandResult struct {
	Hosts    HostCommandResultMap
	Success  uint32 `json:"success" yaml:"success" toml:"success" mapstructure:"success"`
	Failures uint32 `json:"failures" yaml:"failures" toml:"failures" mapstructure:"failures"`
}

// JSON returns the command result in JSON format. With pp will be an indent JSON
func (r *CommandResult) JSON(pp bool) ([]byte, error) {
	if pp {
		return json.MarshalIndent(r, "", "  ")
	}
	return json.Marshal(r)
}

// YAML returns the command result in YAML format
func (r *CommandResult) YAML() ([]byte, error) {
	return yaml.Marshal(r)
}

// TOML returns the command result in TOML format
func (r *CommandResult) TOML() ([]byte, error) {
	return toml.Marshal(r)
}

// NewHostCommandResultMap initializes/makes the map
func NewHostCommandResultMap() HostCommandResultMap {
	return HostCommandResultMap{
		Results: make(map[string]*HostCommandResult),
	}
}

// Load returns the HostCommandResult associated with the host in the given HostCommandResultMap
func (rm *HostCommandResultMap) Load(key string) (*HostCommandResult, bool) {
	rm.RLock()
	result, ok := rm.Results[key]
	rm.RUnlock()
	return result, ok
}

// Store updates the value at the given key in the given HostCommandResultMap
func (rm *HostCommandResultMap) Store(key string, value *HostCommandResult) {
	rm.Lock()
	rm.Results[key] = value
	rm.Unlock()
}

// GetSnapshot returns a snapshot of the given HostCommandResultMap
func (rm *HostCommandResultMap) GetSnapshot() map[string]*HostCommandResult {
	rm.RLock()
	defer rm.RUnlock()
	return rm.Results
}
