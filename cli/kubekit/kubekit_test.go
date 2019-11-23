package kubekit_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kubekit/kubekit/cli/kubekit"
)

var waitingTime = 2 * time.Second

var kubekitCmd = kubekit.RootCmd

// type cmdExecFunc func()

func getOutput(f func()) (string, error) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	outCh := make(chan string)
	timeout := make(chan bool, 1)
	go func() {
		timer := time.AfterFunc(waitingTime, func() {
			r.Close()
			timeout <- true
		})
		defer timer.Stop()

		var buf bytes.Buffer
		io.Copy(&buf, r)
		outCh <- buf.String()
	}()

	w.Close()
	os.Stdout = oldStdout

	select {
	case out := <-outCh:
		return strings.TrimSpace(string(out)), nil
	case <-timeout:
		return "", fmt.Errorf("timeout, the command exceeded %f seconds ", waitingTime.Seconds())
	}
}

func TestLoadConfig(t *testing.T) {
	os.Setenv("KUBEKIT_DEBUG", "true")
	v, err := kubekit.LoadConfig("/this/directory/should/not/exists/config.yaml")
	if err != nil {
		t.Errorf("expected no error loading a config file that does not exists, but got %s", err)
	}
	actualDebug := v.GetBool("debug")
	expectedDebug := true
	if actualDebug != expectedDebug {
		t.Errorf("expected 'debug' to be 'true', but got '%t'", actualDebug)
	}
}
