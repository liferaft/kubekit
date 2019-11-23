package configurator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kraken/ui"
	"github.com/kubekit/kubekit/pkg/configurator/ssh"
)

// Command is a command to execute in a cluster node or to/from it
type Command struct {
	Hosts Hosts
	ui    *ui.UI
}

// NewCommand returns a new command
func NewCommand(hosts Hosts, platformConfig interface{}, ui *ui.UI) (*Command, error) {
	config := make(map[string]interface{})
	configB, err := json.Marshal(platformConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the platform configuration. %s", err)
	}
	err = json.Unmarshal(configB, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the platform configuration. %s", err)
	}

	username, ok := config["username"]
	if !ok {
		return nil, fmt.Errorf("not found username in platform configuration")
	}
	privKey, err := GetPrivateKey(config)
	if err != nil {
		return nil, err
	}
	var password string
	if p, ok := config["password"]; ok {
		password = p.(string)
	}

	if err := hosts.Config(username.(string), privKey, password, false); err != nil {
		return nil, err
	}

	return &Command{
		Hosts: hosts,
		ui:    ui,
	}, nil
}

func (c *Command) executeFnInHosts(hosts Hosts, wg *sync.WaitGroup, fn func(Host)) {
	wg.Add(len(hosts))

	for _, host := range hosts {
		go fn(host)
	}

	wg.Wait()
}

func (c *Command) executeFnInAllHosts(wg *sync.WaitGroup, fn func(Host)) {
	c.executeFnInHosts(c.Hosts, wg, fn)
}

// Copy copies files from/to the cluster hosts
func (c *Command) Copy(from, to string, forceFiles, backupFiles, sudoFiles bool, owner, group, mode string) error {
	// Remote locations should start with colon `:`.
	// It's not allow nor implemented to copy files from remote to remote locations

	// If 'from' is a remote location (begins with `:`) and to doesn't, them copy from 'from' to 'to' locations
	if from[0:1] == ":" && to[0:1] != ":" {
		return c.CopyFrom(from, to, forceFiles, backupFiles, sudoFiles, owner, group, mode)
	}
	// If 'from' is a local location and to is a remote location (begins with `:`), them copy from 'to' to 'from' locations
	if from[0:1] != ":" && to[0:1] == ":" {
		return c.CopyTo(from, to, forceFiles, backupFiles, sudoFiles, owner, group, mode)
	}
	// If we get to this point, means both locations begins with colon `:` (both
	// are remote) or neither has a prefix colon (both are local), and this is a
	// feature not allowed neither implemented: to copy from remote to remote,
	// neither from local to local
	return fmt.Errorf("source and target location are both remote or local. Remote locations begings with ':'")
}

// CopyFrom copies files from the cluster hosts to localhost
func (c *Command) CopyFrom(from, to string, forceFiles, backupFiles, sudoFiles bool, owner, group, mode string) error {
	// 'from' has to be a remote location, (should begin with colon `:`)
	if from[0:1] != ":" {
		return fmt.Errorf("source location is not remote, remote locations begins with ':'. (%s)", from)
	}
	from = from[1:]

	filename := filepath.Base(from)

	bkpEpoch := time.Now().Unix()

	var perm os.FileMode = 0644
	if len(mode) != 0 {
		permUint64, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			return fmt.Errorf("failed to parse the given mode (%s)", mode)
		}
		perm = os.FileMode(permUint64)
	}

	toInfo, err := os.Stat(to)
	if os.IsNotExist(err) {
		return fmt.Errorf("target directory %q does not exists", to)
	}
	if !toInfo.IsDir() {
		return fmt.Errorf("target location %q is not a directory", to)
	}

	errMsg := []string{}
	var wg sync.WaitGroup
	c.executeFnInAllHosts(&wg, func(host Host) {
		defer wg.Done()
		defer host.ssh.Close()

		handleError := func(err error, msg string, a ...interface{}) {
			errMsg = append(errMsg, fmt.Sprintf("%s (%s)", host.PublicIP, err))
			fmt.Printf(msg+"\n", a...)
		}

		targetDir := filepath.Join(to, host.PublicIP)
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			handleError(err, "failed to create the directory for host %s on %s", host.PublicIP, to)
			return
		}
		targetFile := filepath.Join(targetDir, filename)

		_, err = os.Stat(targetFile)
		exists := err == nil
		if !forceFiles && exists && !backupFiles {
			handleError(fmt.Errorf("file exists"), "file %q from %s exists at localhost. Use force to overwrite or make a backup", targetFile, host.PublicIP)
			return
		}

		var bkpMsg string
		if backupFiles && exists {
			if err := os.Rename(targetFile, fmt.Sprintf("%s.%d.bkp", targetFile, bkpEpoch)); err != nil {
				handleError(err, "failed to backup file %q from host %s", targetFile, host.PublicIP)
				return
			}
			bkpMsg = fmt.Sprintf(", file backed up to %s.%d.bkp", targetFile, bkpEpoch)
		}

		length, err := host.ssh.GetFile(targetFile, from, perm)
		if err != nil {
			handleError(err, "failed to create the file %s from host %s. %s", targetFile, host.PublicIP, err)
			return
		}

		fmt.Printf("file %q, %d bytes, copied from host %s%s\n", targetFile, length, host.PublicIP, bkpMsg)
	})
	if len(errMsg) != 0 {
		return fmt.Errorf("failed to copy the file %q from the following hosts: %s", from, strings.Join(errMsg, ", "))
	}
	return nil

}

// CopyTo copies files from the cluster hosts to localhost
func (c *Command) CopyTo(from, to string, forceFiles, backupFiles, sudoFiles bool, owner, group, mode string) error {
	// 'to' has to be a remote location, (should begin with colon `:`)
	if to[0:1] != ":" {
		return fmt.Errorf("target location is not remote, remote locations begins with ':'. (%s)", to)
	}
	to = to[1:]

	filename := filepath.Base(from)
	targetFile := filepath.Join(to, filename)

	bkpEpoch := time.Now().Unix()

	var perm os.FileMode = 0644
	if len(mode) != 0 {
		permUint64, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			return fmt.Errorf("failed to parse the given mode (%s)", mode)
		}
		perm = os.FileMode(permUint64)
	}

	var sudoCmd string
	if sudoFiles {
		sudoCmd = "sudo "
	}

	errMsg := []string{}
	var wg sync.WaitGroup
	c.executeFnInAllHosts(&wg, func(host Host) {
		defer wg.Done()
		defer host.ssh.Close()

		var bkpMsg string

		c.ui.Notify(host.RoleName, "file", "<file>", "", ui.Upload)
		defer func() {
			fmt.Printf("file %q copied to host %s%s\n", targetFile, host.PublicIP, bkpMsg)
			c.ui.Notify(host.RoleName, "file", fmt.Sprintf("file %s uploaded to %s node %s%s", targetFile, host.RoleName, host.PublicIP, bkpMsg), "")
			c.ui.Notify(host.RoleName, "file", "</file>", "", ui.Upload)
		}()

		handleError := func(err error, msg string, a ...interface{}) {
			errMsg = append(errMsg, fmt.Sprintf("%s (%s)", host.PublicIP, err))
			fmt.Printf(msg+"\n", a...)
		}

		content, err := os.Open(from)
		if err != nil {
			handleError(err, "failed to open the source file %q", from)
			return
		}
		defer content.Close()

		dataReader := bufio.NewReader(content)

		exists, _ := host.ssh.ExistsFile(targetFile)

		if !forceFiles && exists && !backupFiles {
			handleError(fmt.Errorf("file exists"), "file %q exists at host %s. Use force to overwrite or make a backup", targetFile, host.PublicIP)
			return
		}

		if backupFiles && exists {
			_, errCmdMsg, exitStat, err := host.ssh.Exec(fmt.Sprintf("%smv %s{,.%d.bkp}", sudoCmd, targetFile, bkpEpoch))
			if err != nil {
				handleError(err, "failed to execute command to backup file %q at host %s", targetFile, host.PublicIP)
				return
			}
			if exitStat != 0 {
				handleError(fmt.Errorf(errCmdMsg), "failed to backup file %q at host %s", targetFile, host.PublicIP)
				return
			}
			bkpMsg = fmt.Sprintf(", file backed up to %s.%d.bkp", targetFile, bkpEpoch)
		}

		if err := host.ssh.CreateFileWithReader(targetFile, dataReader, perm); err != nil {
			handleError(err, "failed to copy the file %q to host %s%s. %s", targetFile, host.PublicIP, bkpMsg, err)
			return
		}

		if len(owner) != 0 {
			if len(group) != 0 {
				owner += ":" + group
			}
			if _, errCmdMsg, exitStat, err := host.ssh.Exec(fmt.Sprintf("%schown %s %s", sudoCmd, owner, targetFile)); err != nil || exitStat != 0 {
				handleError(err, "failed to change the owner (%s) to file %q to host %s%s. %s", owner, targetFile, host.PublicIP, bkpMsg, errCmdMsg)
				return
			}
		} else {
			if len(group) != 0 {
				_, errCmdMsg, exitStat, err := host.ssh.Exec(fmt.Sprintf("%schgrp %s %s", sudoCmd, group, targetFile))
				if err != nil {
					handleError(err, "failed to execute command to change group to file %q at host %s", targetFile, host.PublicIP)
					return
				}
				if exitStat != 0 {
					handleError(fmt.Errorf(errCmdMsg), "failed to change the group (%s) to file %q to host %s%s", group, targetFile, host.PublicIP, bkpMsg)
					return
				}
			}
		}
	})
	if len(errMsg) != 0 {
		return fmt.Errorf("failed to copy the file %q to the following hosts: %s", from, strings.Join(errMsg, ", "))
	}
	return nil
}

// Exec executes a script file or command line on every command host
func (c *Command) Exec(command, script string, sudoExec bool) (*ssh.CommandResult, error) {
	var result ssh.CommandResult
	result.Hosts = ssh.NewHostCommandResultMap()

	var sudoCmd string
	if sudoExec {
		sudoCmd = "sudo "
	}

	var content []byte
	var targetFile string
	if len(script) != 0 {
		filename := filepath.Base(script)
		targetFile = filepath.Join("/tmp", filename)

		var err error
		content, err = ioutil.ReadFile(script)
		if err != nil {
			return &result, fmt.Errorf("failed to read the script file %q", script)
		}
		command = filepath.Join("/tmp", filename)
	}

	errMsg := []string{}
	var wg sync.WaitGroup
	c.executeFnInAllHosts(&wg, func(host Host) {
		defer wg.Done()
		defer host.ssh.Close()

		handleError := func(err error, msg string, a ...interface{}) {
			errMsg = append(errMsg, fmt.Sprintf("%s (%s)", host.PublicIP, err))
			fmt.Printf(msg+"\n", a...)
		}

		if len(content) != 0 {
			if err := host.ssh.CreateFile(targetFile, string(content), 0700); err != nil {
				handleError(err, "failed to copy the script %q to host %s", targetFile, host.PublicIP)
				return
			}
		}

		execCmd := strings.TrimSpace(command)
		if sudoCmd == "sudo " && !strings.HasPrefix(execCmd, "sudo ") {
			execCmd = sudoCmd + execCmd
		}

		outCmdMsg, errCmdMsg, exitStat, err := host.ssh.Exec(execCmd)
		if err != nil {
			handleError(err, "failed to execute command %q at host %s", execCmd, host.PublicIP)
			return
		}

		result.Hosts.Store(host.PublicIP, &ssh.HostCommandResult{
			Stdout:     outCmdMsg,
			Stderr:     errCmdMsg,
			ExitStatus: exitStat,
		})

		// var outputMsg string
		if exitStat == 0 {
			atomic.AddUint32(&result.Success, 1)
			// outputMsg = "command %q successfuly executed at host %s"
		} else {
			atomic.AddUint32(&result.Failures, 1)
			// outputMsg = "command %q failed execution at host %s"
		}
		// fmt.Printf(outputMsg+"\n", command, host.PublicIP)
	})
	var err error
	if len(errMsg) != 0 {
		err = fmt.Errorf("failed to execute the command %q at the following hosts: %s", command, strings.Join(errMsg, ", "))
	}
	return &result, err
}

// TODO: Do benchmark test with previous code and the following:
// func (c *Command) generateHosts(done <-chan interface{}) <-chan *Host {
// 	hostStream := make(chan *Host)
// 	go func() {
// 		defer close(hostStream)
// 		for _, h := range c.Hosts {
// 			select {
// 			case <-done:
// 				return
// 			case hostStream <- &h:
// 			}
// 		}
// 	}()
// 	return hostStream
// }

// CopyTo copies files from the cluster hosts to localhost
// func (c *Command) CopyTo(from, to string, forceFiles, backupFiles bool, owner, group, mode string) error {
// 	if to[0:1] != ":" {
// 		return fmt.Errorf("target location is not remote, remote locations begins with ':'. (%s)", to)
// 	}
// 	to = to[1:]

// 	content, err := ioutil.ReadFile(from)
// 	if err != nil {
// 		return fmt.Errorf("failed to read the target file %q: %s", from, err)
// 	}

// 	type Result struct {
// 		Error error
// 		IP    string
// 	}

// 	copy := func(done <-chan interface{}, hostStream <-chan *Host) <-chan Result {
// 		results := make(chan Result)
// 		go func() {
// 			defer close(results)
// 			for {
// 				select {
// 				case <-done:
// 					return
// 				case h := <-hostStream:
// 					fmt.Println("Copying file to " + h.PrivateIP)
// 					err := h.ssh.CreateFile(to, string(content))
// 					result := Result{
// 						Error: err,
// 						IP:    h.PublicIP,
// 					}
// 					results <- result
// 				}
// 			}
// 		}()
// 		return results
// 	}

// 	done := make(chan interface{})
// 	results := copy(done, c.generateHosts(done))

// 	errMsg := []string{}
// 	for result := range results {
// 		if result.Error != nil {
// 			errMsg = append(errMsg, fmt.Sprintf("%s (%s)", result.IP, result.Error))
// 		}
// 	}
// 	if len(errMsg) != 0 {
// 		return fmt.Errorf("failed to copy the file %q to the following hosts: %s", from, strings.Join(errMsg, ", "))
// 	}

// 	return nil
// }
