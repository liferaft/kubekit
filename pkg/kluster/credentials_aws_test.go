package kluster

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johandry/log"
	"github.com/kraken/ui"
)

func invalidTestKluster(t *testing.T) *Kluster {
	cluster := &Kluster{
		path: filepath.Join("testdata", "invalid_eks", "cluster.yaml"),
		ui:   ui.New(false, log.StdLogger()),
	}

	err := cluster.Load()
	if err != nil {
		t.Fatalf("failed to load testing cluster: %s", err)
	}

	return cluster
}

func invalidTestKlusterCreds(t *testing.T) *AwsCredentials {
	cluster := invalidTestKluster(t)
	path := filepath.Join(filepath.Dir(cluster.Path()), CredentialsFileName)
	creds := NewCredentials(cluster.Name, cluster.Platform(), path)
	awsCreds, ok := creds.(*AwsCredentials)
	if !ok {
		t.Fatalf("failed to assert testing cluster credentials as AWS")
	}

	awsCreds.AccessKey = "test_access_key"
	awsCreds.SecretKey = "test_secret_key"
	awsCreds.SessionToken = "test_session_token"
	awsCreds.Region = "test_region"

	return awsCreds
}

func TestKluster_GetCredentialsAsMap(t *testing.T) {
	cases := []struct{ key, want string }{
		{"access_key", "test_access_key"},
		{"secret_key", "test_secret_key"},
		{"session_token", "test_session_token"},
		{"region", "test_region"},
	}

	awsCreds := invalidTestKlusterCreds(t)
	creds := awsCreds.asMap()

	for _, tc := range cases {
		actual := creds[tc.key]

		if actual != tc.want {
			t.Fatalf("clusted credential for %s: %s, expected: %s", tc.key, actual, tc.want)
		}
	}
}

func TestAwsCredentials_Getenv(t *testing.T) {
	creds := invalidTestKlusterCreds(t)

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if err := os.Setenv("AWS_ACCESS_KEY_ID", "new_access_key"); err != nil {
		t.Fatalf("failed to set environment variable")
	}

	if err := os.Setenv("AWS_SECRET_ACCESS_KEY", "new_secret_key"); err != nil {
		t.Fatalf("failed to set environment variable")
	}

	defer os.Setenv("AWS_ACCESS_KEY_ID", accessKey)
	defer os.Setenv("AWS_SECRET_ACCESS_KEY", secretKey)

	if err := creds.Getenv(true); err != nil {
		t.Fatalf("failed to get test cluster creds from env")
	}

	if creds.AccessKey == "test_access_key" {
		t.Errorf("failed to fetch credential access key from environment variables")
	}

	if creds.SecretKey == "test_secret_key" {
		t.Errorf("failed to fetch credential secret key from environment variables")
	}
}

func TestInvalidCreds(t *testing.T) {
	creds := invalidTestKlusterCreds(t)
	if err := creds.Validate(); err == nil {
		t.Errorf("invalid creds expected error but got nil")
	}
}

func TestValidCreds(t *testing.T) {
	creds := invalidTestKlusterCreds(t)
	if err := creds.Getenv(true); err != nil {
		t.Fatalf("failed to get test cluster credentials from env variables")
	}

	if strings.Contains(creds.SessionToken, "test") {
		creds.SessionToken = ""
	}

	if strings.Contains(creds.Region, "test") {
		creds.Region = "us-west-2"
	}

	if strings.Contains(creds.AccessKey, "test") {
		t.Skip("Invalid AWS Credentials provided. Can't test for Valid Credentials")
		t.SkipNow()
	}

	if err := creds.Validate(); err != nil {
		t.Errorf("valid creds expected but got: %s", err)
	}

}

func TestAwsCredentials_AssignFromMap(t *testing.T) {
	creds := invalidTestKlusterCreds(t)
	nuCreds := invalidTestKlusterCreds(t)

	if err := creds.Getenv(true); err != nil {
		t.Fatalf("failed to get test cluster creds env")
	}

	if err := nuCreds.AssignFromMap(creds.asMap()); err != nil {
		t.Fatalf("failed to assign creds from map")
	}

	if nuCreds.AccessKey != creds.AccessKey {
		t.Errorf("creds from map access key expected: %s, got: %s", creds.AccessKey, nuCreds.AccessKey)
	}

	if nuCreds.SecretKey != creds.SecretKey {
		t.Errorf("creds from map secret key expected: %s, got: %s", creds.SecretKey, nuCreds.SecretKey)
	}

	if nuCreds.SessionToken != creds.SessionToken {
		t.Errorf("creds from map session token expected: %s, got: %s", creds.SessionToken, nuCreds.SessionToken)
	}

	if nuCreds.Region != creds.Region {
		t.Errorf("creds from map region expected: %s, got: %s", creds.Region, nuCreds.Region)
	}
}
