package ssh

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/pkg/sftp"
)

// CreateFile creates a file with the filename 'target' and the given
// content using the uid, gid and permissions provided.
func (c *Config) CreateFile(target, data string, perm os.FileMode) error {
	dataReader := bufio.NewReader(strings.NewReader(data))
	return c.CreateFileWithReader(target, dataReader, perm)
}

// CreateFileWithReader creates a file with the filename 'target' and the given
// content using the uid, gid and permissions provided.
func (c *Config) CreateFileWithReader(target string, reader io.Reader, perm os.FileMode) error {
	err := c.setClient()
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	targetFile, err := sftpClient.Create(target)
	if err != nil {
		return err
	}

	defer targetFile.Close()

	_, err = io.Copy(targetFile, reader)
	if err != nil {
		return err
	}
	return targetFile.Chmod(perm)
}

// SetChown to set the owner of a remote object
func (c *Config) SetChown(name string, uid string) error {
	err := c.setClient()
	if err != nil {
		return err
	}
	command := "sudo chown " + uid + " " + name
	// stdOut, stdErr, exitStatus,
	_, _, _, err = c.Exec(command)
	return err
}

// SetChmod to set the permissions of a remote object
func (c *Config) SetChmod(name, mode string) error {
	err := c.setClient()
	if err != nil {
		return err
	}
	command := "sudo chmod " + mode + " " + name

	// stdOut, stdErr, exitStatus,
	_, _, _, err = c.Exec(command)
	return err
}

// CreateGroup create group for users
func (c *Config) CreateGroup(name string) error {
	err := c.setClient()
	if err != nil {
		return err
	}
	command := "sudo groupadd -f " + name

	// stdOut, stdErr, exitStatus,
	_, _, _, err = c.Exec(command)
	return err
}

// UserMod adds user to group
func (c *Config) UserMod(name string, group string) error {
	err := c.setClient()
	if err != nil {
		return err
	}
	command := "sudo usermod -a -G " + group + " " + name

	// stdOut, stdErr, exitStatus,
	_, _, _, err = c.Exec(command)

	return err
}

// GetFile streams a file from the remote host and writes it out to "to"
func (c *Config) GetFile(to, from string, perm os.FileMode) (int64, error) {
	err := c.setClient()
	if err != nil {
		return 0, err
	}

	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return 0, err
	}
	defer sftpClient.Close()

	dataFrom, err := sftpClient.Open(from)
	if err != nil {
		return 0, err
	}
	defer dataFrom.Close()

	dataTo, err := os.OpenFile(to, os.O_RDWR|os.O_CREATE, perm)
	if err != nil {
		return 0, err
	}
	defer dataTo.Close()

	return io.Copy(dataTo, dataFrom)
}

// ExistsFile return true if a file exists
func (c *Config) ExistsFile(file string) (bool, error) {
	err := c.setClient()
	if err != nil {
		return false, err
	}

	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return false, err
	}
	defer sftpClient.Close()

	info, err := sftpClient.Stat(file)
	if err != nil {
		return false, err
	}

	return info.Mode().IsRegular(), nil
}

// MkDir creates a remote directory
func (c *Config) MkDir(dir string) error {
	err := c.setClient()
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	err = sftpClient.MkdirAll(dir)
	if err != nil {
		return err
	}

	if _, errStat := sftpClient.Stat(dir); errStat == nil {
		return errStat
		// there is a file with the same name as 'dir'
		// or the directory is there and the cause of why it fail (if this scenario
		// is possible) is unknonw
	}

	return nil
}

// SudoMkDir creates a remote directory
func (c *Config) SudoMkDir(dir string) error {
	err := c.setClient()
	if err != nil {
		return err
	}

	// if we are here, then maybe it's not possible to create the file due to permissions
	command := "sudo mkdir -p " + dir

	// stdOut, stdErr, exitStatus,
	_, _, _, err = c.Exec(command)
	return err
}
