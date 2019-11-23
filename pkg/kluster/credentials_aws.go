package kluster

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awscredentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/go-ini/ini"
	homedir "github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
)

// AwsCredentials represents the credentials just for AWS
type AwsCredentials struct {
	Platform          string `json:"platform" yaml:"platform" toml:"platform" mapstructure:"platform" env:"-"`
	AccessKey         string `json:"access_key" yaml:"access_key" toml:"access_key" mapstructure:"access_key" env:"ACCESS_KEY_ID"`
	SecretKey         string `json:"secret_key" yaml:"secret_key" toml:"secret_key" mapstructure:"secret_key" env:"SECRET_ACCESS_KEY"`
	SessionToken      string `json:"session_token" yaml:"session_token" toml:"session_token" mapstructure:"session_token" env:"SESSION_TOKEN"`
	Region            string `json:"region" yaml:"region" toml:"region" mapstructure:"region" env:"DEFAULT_REGION"`
	Profile           string `json:"-" yaml:"-" toml:"-" mapstructure:"-" env:"PROFILE"`
	sharedCredentials *awscredentials.Value
	sharedRegion      string
	cluster           string
	path              string
}

// NewAWSCredentials creates an struct ready for AWS credentials
func NewAWSCredentials(clustername, path string) *AwsCredentials {
	return &AwsCredentials{
		Platform: "AWS",
		cluster:  clustername,
		path:     path,
	}
}

func (c *AwsCredentials) clusterName() string {
	return c.cluster
}

func (c *AwsCredentials) clusterPath() string {
	return c.path
}

// SetPath sets the path of the credentials file be stored
func (c *AwsCredentials) SetPath(path string) {
	c.path = path
}

func (c *AwsCredentials) platform() string {
	return c.Platform
}

func (c *AwsCredentials) parameters() []string {
	return []string{c.AccessKey, c.SecretKey, c.SessionToken, c.Region}
}

func (c *AwsCredentials) asMap() map[string]string {
	return map[string]string{
		"access_key":    c.AccessKey,
		"secret_key":    c.SecretKey,
		"session_token": c.SessionToken,
		"region":        c.Region,
	}
}

// SetParameters sets the credentials parameters
func (c *AwsCredentials) SetParameters(params ...string) error {
	if len(params) != 4 {
		return fmt.Errorf("incorrect number of parameters, expecting 4 and got %d", len(params))
	}
	c.AccessKey = params[0]
	c.SecretKey = params[1]
	c.SessionToken = params[2]
	c.Region = params[3]

	return nil
}

// AssignFromMap sets the credentials parameters from a map. The key is the name
// of the parameter as defined in the json metadata of the AWSCredentials structure
func (c *AwsCredentials) AssignFromMap(params map[string]string) error {
	return assignFromMap(params, c)
}

// Empty returns true if there isn't any credentials set
func (c *AwsCredentials) Empty() bool {
	return len(c.AccessKey) == 0 && len(c.SecretKey) == 0 && len(c.Region) == 0
}

// Complete returns true if all credentials are set. Use it as !Complete() to
// know if there is a missing parameter. !Complete() is not Empty()
func (c *AwsCredentials) Complete() bool {
	return len(c.AccessKey) != 0 && len(c.SecretKey) != 0 && len(c.Region) != 0
}

// Getenv gets the AWS credentials from environment variables
func (c *AwsCredentials) Getenv(force bool) error {
	return parseFn(c,
		func(key string) string {
			return os.Getenv(strings.ToUpper(c.platform()) + "_" + key)
		},
		func(k, v string) {},
		force,
	)
}

func (c *AwsCredentials) setenv() error {
	return parseFn(c,
		func(key string) string {
			return ""
		},
		func(key, value string) {
			os.Setenv(strings.ToUpper(c.platform())+"_"+key, value)
		},
		true,
	)
}

// List prints to stdout the AWS credentials in table format
func (c *AwsCredentials) List() {
	header := "Name\tPlatform\tRegion\tAccess Key\tSecret Key\tSession Token\tProfile" //\tLocation"
	accessKey := c.AccessKey
	if len(accessKey) == 0 {
		accessKey = emptyValue
	} else {
		accessKey = filter(accessKey, truncLen, false)
	}
	secretKey := c.SecretKey
	if len(secretKey) == 0 {
		secretKey = emptyValue
	} else {
		secretKey = filter(secretKey, truncLen, false)
	}
	sessionToken := c.SessionToken
	if len(sessionToken) == 0 {
		sessionToken = emptyValue
	}
	region := c.Region
	if len(region) == 0 {
		region = emptyValue
	}
	profile := c.Profile
	if len(profile) == 0 {
		profile = emptyValue
	}
	row := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", c.cluster, c.Platform, region, accessKey, secretKey, sessionToken, profile) //, c.path)
	printCredentials(header, row)
}

// Ask to the user from stdin the AWS credentials suggesting values from the environment
func (c *AwsCredentials) Ask() error {
	return parseFn(c,
		func(key string) string {
			param := strings.Replace(strings.ToLower(key), "_", " ", -1)
			title := fmt.Sprintf("%s %s", c.Platform, param)
			env := os.Getenv(strings.ToUpper(c.platform()) + "_" + key)
			if len(env) == 0 && c.sharedCredentials != nil {
				switch key {
				case "ACCESS_KEY_ID":
					env = c.sharedCredentials.AccessKeyID
				case "SECRET_ACCESS_KEY":
					env = c.sharedCredentials.SecretAccessKey
				case "DEFAULT_REGION":
					env = c.sharedRegion
				}
			}
			v, _ := AskDefault(title, env, true)
			return v
		},
		func(k, v string) {},
		true,
	)
}

// Read reads the AWS credentials from the cluster credentials file
func (c *AwsCredentials) Read() error {
	credentialsBytes, err := read(c.path)
	if err != nil || credentialsBytes == nil {
		return err
	}
	return yaml.Unmarshal(credentialsBytes, &c)
}

// Write writes the AWS credentials to the cluster credentials file
func (c *AwsCredentials) Write() error {
	credentialsBytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.path, credentialsBytes, 0600)
}

// LoadSharedCredentialsFromProfile loads the AWS credentials from the AWS shared credentials file for the given profile
func (c *AwsCredentials) LoadSharedCredentialsFromProfile(profile string, force bool) error {
	// Read the AWS credentials from the credentials shared file
	// (~/.aws/credentials). If the user provide --profile or AWS_PROFILE then
	// that overwrite the AWS credentials, otherwise read the default profile
	// and keep it in shared to provide then as defaults when ask but do not
	// overwrite the AWS credentials

	// If --profile or AWS_PROFILE not set, read the default profile
	sharedCred, err := awscredentials.NewSharedCredentials("", profile).Get()
	if err != nil {
		return fmt.Errorf("Cannot read AWS shared credentials. %s", err)
	}

	// if instructed to use force was given or the field is empty, assing such credentials to the
	if force || len(c.AccessKey) == 0 {
		c.AccessKey = sharedCred.AccessKeyID
	}
	if force || len(c.SecretKey) == 0 {
		c.SecretKey = sharedCred.SecretAccessKey
	}
	if force || len(c.SessionToken) == 0 {
		c.SessionToken = sharedCred.SessionToken
	}
	// assign such credentials into shared
	c.sharedCredentials = &sharedCred

	return nil
}

// LoadSharedRegionFromProfile loads the AWS region from the AWS shared configuration file for the given profile
func (c *AwsCredentials) LoadSharedRegionFromProfile(profile string, force bool) error {
	// AWS do not provide a simple API to read the shared configuration, unless
	// creating a session which is too much just for this.

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	sharedConfigFilename := filepath.Join(home, ".aws", "config")

	b, err := ioutil.ReadFile(sharedConfigFilename)
	if err != nil {
		return fmt.Errorf("Cannot read AWS shared config file %s. %s", sharedConfigFilename, err)
	}

	f, err := ini.Load(b)
	if err != nil {
		return fmt.Errorf("Cannot load AWS shared config file %s. %s", sharedConfigFilename, err)
	}
	section, err := f.GetSection(profile)
	if err != nil {
		return fmt.Errorf("Cannot get %q profile from AWS shared config file %s. %s", profile, sharedConfigFilename, err)
	}

	key, err := section.GetKey("region")
	if err != nil {
		return fmt.Errorf("Cannot get region in profile %s from AWS shared config file %s. %s", profile, sharedConfigFilename, err)
	}

	if force || len(c.Region) == 0 {
		// if instructed to use force was given or the field is empty, assign the read Region
		c.Region = key.Value()
	}

	// assign such region into shared
	c.sharedRegion = key.Value()

	return nil
}

// Refresh ensures that the AWS credentials are valid
// if the credentials are invalid it will check the environment variables
// and AWS profile fro valid credentials in that order
// write flag determines whether the new creds are written
// reload flag determines whether the new creds affect the current session
func (c *AwsCredentials) Refresh(write, reload bool) error {
	err := c.Validate()
	if err == nil {
		return nil
	}

	// let's try to reset them if they're invalid
	// create the new creds for testing different options
	creds := NewAWSCredentials(c.cluster, c.path)

	// try first with environment variables
	err = creds.Getenv(true)
	if err != nil {
		return err
	}

	// function to update the creds if they're valid
	updateCreds := func() error {
		if write {
			if err = creds.Write(); err != nil {
				return err
			}
		}

		if reload {
			if err = c.AssignFromMap(creds.asMap()); err != nil {
				return err
			}
		}

		return nil
	}

	// check if they're valid
	if err = creds.Validate(); err == nil {
		return updateCreds()
	}

	// next we'll try with the AWS profile
	if len(c.Profile) == 0 {
		c.Profile = os.Getenv("AWS_PROFILE")
	}

	profile := c.Profile
	if len(c.Profile) == 0 {
		profile = "default"
	}

	errCred := creds.LoadSharedRegionFromProfile(profile, true)
	errReg := creds.LoadSharedCredentialsFromProfile(profile, true)

	if errCred == nil && errReg == nil {
		c.Profile = profile
	}

	if err = c.Validate(); err == nil {
		return updateCreds()
	}

	return errors.New("couldn't load valid AWS credentials (tried cluster credentials, env, and AWS profile)")
}

// Validate validates the AWS session by executing a simple call to the AWS API
func (c *AwsCredentials) Validate() error {
	sess, err := GetSession(c.asMap())
	if err != nil {
		return err
	}

	return validateSession(sess)
}

func validateSession(sess *session.Session) error {
	svc := sts.New(sess)
	return retry(1, func() (err error) {
		_, err = svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		return
	})
}

// GetSession creates an AWS session from the provided creds
func GetSession(creds map[string]string) (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:      aws.String(creds["region"]),
		Credentials: awscredentials.NewStaticCredentials(creds["access_key"], creds["secret_key"], creds["session_token"]),
	})
}

func retry(timeoutSeconds int, f func() error) (err error) {
	err = f()
	if err == nil {
		return
	}

	timer := time.NewTicker(time.Second)
	defer timer.Stop()

	timeout := time.After(time.Second * time.Duration(timeoutSeconds))

	for {
		select {
		case <-timer.C:
			err = f()
			if err == nil {
				return
			}
		case <-timeout:
			return
		}
	}
}
