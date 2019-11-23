package ssh

import (
	"archive/zip"
	"io"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
)

// CreateZip creates a zip file with the filename 'target' and content given in 'data'
func (c *Config) CreateZip(target, data string) error {
	err := c.setClient()
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	// zipReader, err := zip.NewReader(strings.NewReader(data), int64(len(data)))
	zipReader := strings.NewReader(data)
	// if err != nil {
	// 	return err
	// }

	targetFile, err := sftpClient.Create(target)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	// _, err = destFile.Write([]byte(data))
	_, err = io.Copy(targetFile, zipReader)

	return err
}

// UnZip unzips the content of 'data' into the remote directory given in 'target'
func (c *Config) UnZip(target, data string) error {
	err := c.setClient()
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	zipReader, err := zip.NewReader(strings.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	err = sftpClient.MkdirAll(target)
	if err != nil {
		return err
	}

	for _, file := range zipReader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			err = sftpClient.MkdirAll(path)
			if err != nil {
				return err
			}
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			if fileReader != nil {
				fileReader.Close()
			}
			return err
		}

		targetFile, err := sftpClient.Create(path)
		if err != nil {
			fileReader.Close()
			if targetFile != nil {
				targetFile.Close()
			}
			return err
		}

		_, err = io.Copy(targetFile, fileReader)
		fileReader.Close()
		targetFile.Close()
		if err != nil {
			return err
		}
	}

	return err
}
