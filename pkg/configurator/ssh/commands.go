package ssh

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/shiena/ansicolor"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// Shell initiate an interactive remote shell
func (c *Config) Shell(inReader io.Reader, outWriter, errWriter io.Writer) error {
	err := c.setClient()
	if err != nil {
		return err
	}
	defer c.Close()

	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = ansicolor.NewAnsiColorWriter(outWriter)
	session.Stderr = ansicolor.NewAnsiColorWriter(errWriter)
	session.Stdin = inReader
	// in, _ := session.StdinPipe()

	modes := ssh.TerminalModes{
		// ssh.ECHO:  0, // Disable echoing
		// ssh.IGNCR: 1, // Ignore CR on input
		ssh.ECHO:          1,      // Print what I type
		ssh.ECHOCTL:       0,      // Don't print control chars
		ssh.TTY_OP_ISPEED: 115200, // baud in
		ssh.TTY_OP_OSPEED: 115200, // baud out
	}

	h, w := 80, 40
	var termFD int
	var ok bool
	if termFD, ok = isTerminal(inReader); ok {
		w, h, _ = terminal.GetSize(termFD)
	}

	termState, _ := terminal.MakeRaw(termFD)
	defer terminal.Restore(termFD, termState)

	// Request pseudo terminal
	// if err := session.RequestPty("xterm", h, w, modes); err != nil {
	// if err := session.RequestPty("xterm-256color", h, w, modes); err != nil {
	// if err := session.RequestPty("vt220", h, w, modes); err != nil {
	// if err := session.RequestPty("vt100", h, w, modes); err != nil {
	if err := session.RequestPty("xterm-256color", h, w, modes); err != nil {
		return fmt.Errorf("request for pseudo terminal failed: %s", err)
	}

	// Start remote shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %s", err)
	}

	return session.Wait()

	// // Handle control + C
	// ch := make(chan os.Signal, 1)
	// signal.Notify(ch, os.Interrupt)
	// go func() {
	// 	for {
	// 		<-ch
	// 		fmt.Println("^C")
	// 		fmt.Fprint(in, "\n")
	// 		//fmt.Fprint(in, '\t')
	// 	}
	// }()

	// // Accepting commands
	// for {
	// 	reader := bufio.NewReader(i)
	// 	str, _ := reader.ReadString('\n')
	// 	fmt.Fprint(in, str)
	// }
}

func isTerminal(r io.Reader) (int, bool) {
	switch v := r.(type) {
	case *os.File:
		return int(v.Fd()), terminal.IsTerminal(int(v.Fd()))
	default:
		return 0, false
	}
}

func keepAlive(cl *ssh.Session) error {
	const keepAliveInterval = time.Minute
	t := time.NewTicker(keepAliveInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			_, err := cl.SendRequest("keepalive@golang.org", false, nil)
			if err != nil {
				return err
			}
		}
	}
}

// // Start initiate the connection and runs command on remote host
// func (c *Config) Start(cmd *Command) error {
// 	err := c.setClient()
// 	if err != nil {
// 		return err
// 	}

// 	// Testing:
// 	session, err := c.client.NewSession()
// 	if err != nil {
// 		return err
// 	}
// 	defer session.Close()

// 	//We send the keepalive every minute in case the server has
// 	// a session timeout set.
// 	go keepAlive(session)

// 	session.Stdout = &cmd.Stdout
// //	session.Stdin = &cmd.Stdin
// 	session.Stderr = &cmd.Stderr
// 	if err := session.Run(cmd.Command); err != nil {
// 		return err
// 	}

// 	return nil
// }

// Start now calls StartAndWait for proper call handling
func (c *Config) Start(cmd *Command) error {
	return c.StartAndWait(cmd)
}

// StartAndWait initiates the connection, runs command on remote host and waits for output
func (c *Config) StartAndWait(cmd *Command) error {
	err := c.setClient()
	if err != nil {
		return err
	}

	// Testing:
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	//We send the keepalive every minute in case the server has
	// a session timeout set.
	go keepAlive(session)

	session.Stdout = &cmd.Stdout
	session.Stderr = &cmd.Stderr

	if err = session.Start(cmd.Command + "\n"); err != nil {
		return err
	}

	err = session.Wait()
	if err != nil {
		switch err.(type) {
		case *ssh.ExitError:
			cmd.ExitStatus = err.(*ssh.ExitError).ExitStatus()
		default:
			return err
		}
	}

	return nil
}

// Exec executes a command in the remote host
func (c *Config) Exec(command string) (stdOut string, stdErr string, exitStatus int, err error) {
	return c.ExecAndWait(command)
}

// ExecAndWait executes a command in the remote host and waits for it to exit
func (c *Config) ExecAndWait(command string) (stdOut string, stdErr string, exitStatus int, err error) {
	cmd := &Command{
		Command: command,
	}
	err = c.StartAndWait(cmd)
	defer c.Close()

	stdOut = strings.TrimRight(cmd.Stdout.String(), "\n")
	stdErr = strings.TrimRight(cmd.Stderr.String(), "\n")
	exitStatus = cmd.ExitStatus

	return
}
