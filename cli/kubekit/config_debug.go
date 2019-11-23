package kubekit

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type myFlag struct {
	value    string
	commands map[string][]string
}
type myFlagSet map[string]*myFlag

func (c *Config) debug() string {
	var b bytes.Buffer

	fmt.Fprintf(&b, "\nLogger:\n")
	fmt.Fprintf(&b, "Prefix:\t\t\t%s\n", c.UI.Log.GetPrefix())
	fmt.Fprintf(&b, "Level:\t\t\t%s\n", c.UI.Log.Level)

	fmt.Fprintf(&b, "\nViper:\n")
	for k, v := range c.viper.AllSettings() {
		tabs := "\t\t"
		if len(k) < 7 {
			tabs += "\t"
		}
		fmt.Fprintf(&b, "%s:%s%v\n", k, tabs, v)
	}

	flagsMap := make(myFlagSet)
	var flagType string
	var cmdName string

	var printFlags func(*pflag.Flag)
	printFlags = func(f *pflag.Flag) {
		if _, ok := flagsMap[f.Name]; !ok {
			flagsMap[f.Name] = &myFlag{
				value: f.Value.String(),
				commands: map[string][]string{
					cmdName: []string{flagType},
				},
			}
		} else {
			flagsMap[f.Name].commands[cmdName] = append(flagsMap[f.Name].commands[cmdName], flagType)
			// Just in case different commands assign different values to the same flag
			// if flagsMap[f.Name].value != f.Value.String() {
			// 	flagsMap[f.Name].commands = fmt.Sprintf("%s, %s(%s)[%s]", flagsMap[f.Name].commands, cmdName, flagType, f.Value)
			// } else {
			// 	flagsMap[f.Name].commands = fmt.Sprintf("%s, %s(%s)", flagsMap[f.Name].commands, cmdName, flagType)
			// }
		}
	}
	var printCommands func(*cobra.Command)
	printCommands = func(x *cobra.Command) {
		cmdName = x.Name()
		if x.HasFlags() {
			flagType = "F"
			x.Flags().VisitAll(printFlags)
			flagType = "I"
			x.InheritedFlags().VisitAll(printFlags)
			flagType = "L"
			x.LocalFlags().VisitAll(printFlags)
			flagType = "LnP"
			x.LocalNonPersistentFlags().VisitAll(printFlags)
			flagType = "nI"
			x.NonInheritedFlags().VisitAll(printFlags)
			flagType = "P"
			x.PersistentFlags().VisitAll(printFlags)
		}
		if x.HasSubCommands() {
			for _, y := range x.Commands() {
				printCommands(y)
			}
		}
	}

	printCommands(c.command)

	fmt.Fprintf(&b, "\nFlags:\n")
	for f, v := range flagsMap {
		tabs := "\t\t"
		if len(f) < 7 {
			tabs += "\t"
		}
		var cmds []string
		for cmd, fTypes := range v.commands {
			cmds = append(cmds, fmt.Sprintf("%s (%s)", cmd, strings.Join(fTypes, ",")))
		}
		fmt.Fprintf(&b, "%s:%s%s\t[%s]\n", f, tabs, v.value, strings.Join(cmds, ", "))
	}

	return b.String()

	// fmt.Fprintf(b, "\nViper Debug:")
	// c.viper.Debug()
}
