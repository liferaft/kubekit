package utils

import (
	"fmt"

	"github.com/kraken/ui"

	"github.com/kraken/terraformer"
)

const defaultLogPrefix = "Provisioner"

// NewTerraformer creates and returns a terraformer to be used by a provisioner
func NewTerraformer(code []byte, variables map[string]interface{}, state *terraformer.State, clusterName string, platform string, ui *ui.UI) (*terraformer.Terraformer, error) {
	var logPrefix string
	if len(platform) != 0 {
		logPrefix = fmt.Sprintf("@%s", platform)
	}
	if len(clusterName) != 0 {
		logPrefix = fmt.Sprintf("%s%s", clusterName, logPrefix)
	}
	if len(logPrefix) != 0 {
		logPrefix = fmt.Sprintf(" [ %s ]", logPrefix)
	}
	logPrefix = defaultLogPrefix + logPrefix
	ui.SetLogPrefix(logPrefix)

	hook := NewUIHook(platform, ui)

	t, err := terraformer.New(ui.Log, hook)
	if err != nil {
		return nil, err
	}

	t.State = state
	t.Vars = variables
	t.Code = code

	return t, nil
}

// NewLogger returns a Terraformer Logger.
// At this time are 3 kind of Loggers, choose the one you like most. Check the
// terraformer example repo.
// func NewLogger(w io.Writer, forceColors bool, prefix string, level string) terraformer.Logger {
// 	v := viper.New()
// 	v.Set(log.OutputKey, w)
// 	v.Set(log.ForceColorsKey, forceColors)
// 	v.Set(log.LevelKey, level)

// 	l := log.New(v)
// 	l.SetPrefix(prefix)

// 	return l
// }
