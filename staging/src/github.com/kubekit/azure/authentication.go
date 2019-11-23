package azure

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"unicode/utf16"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure/cli"
	"github.com/caarlos0/env"
	"github.com/dimchansky/utfbom"
)

// Session holds the subscription and authorizer settings so that we do not have to keep creating a new authorizer
type Session struct {
	Authorizer     autorest.Authorizer
	SubscriptionID string
}

// AuthInfo is essentially a combined file and settings struct in github.com/Azure/go-autorest/autorest/auth,
// but made to be exportable and allows for everything to be set through environment variables
type AuthInfo struct {
	ClientID                string `json:"clientId,omitempty" env:"AZURE_CLIENT_ID"`
	ClientSecret            string `json:"clientSecret,omitempty" env:"AZURE_CLIENT_SECRET"`
	SubscriptionID          string `json:"subscriptionId,omitempty" env:"AZURE_CERTIFICATE_PATH"`
	TenantID                string `json:"tenantId,omitempty" env:"AZURE_CERTIFICATE_PASSWORD"`
	ActiveDirectoryEndpoint string `json:"activeDirectoryEndpointUrl,omitempty" env:"AZURE_AD_ENDPOINT"`
	ResourceManagerEndpoint string `json:"resourceManagerEndpointUrl,omitempty" env:"AZURE_RM_ENDPOINT"`
	GraphResourceID         string `json:"activeDirectoryGraphResourceId,omitempty" env:"AZURE_GRAPH_RESOURCE_ID"`
	SQLManagementEndpoint   string `json:"sqlManagementEndpointUrl,omitempty" env:"AZURE_SQLM_ENDPOINT"`
	GalleryEndpoint         string `json:"galleryEndpointUrl,omitempty" env:"AZURE_GALLERY_ENDPOINT"`
	ManagementEndpoint      string `json:"managementEndpointUrl,omitempty" env:"AZURE_MANAGEMENT_ENDPOINT"`
	Environment             string `json:"environment,omitempty" env:"AZURE_ENVIRONMENT"`
	CertificatePath         string `json:"certificatePath,omitempty" env:"AZURE_CERTIFICATE_PATH"`
	CertificatePassword     string `json:"certificatePassword,omitempty" env:"AZURE_CERTIFICATE_PASSWORD"`
	Username                string `json:"username,omitempty" env:"AZURE_USERNAME"`
	Password                string `json:"password,omitempty" env:"AZURE_PASSWORD"`
	Resource                string `json:"resource,omitempty" env:"AZURE_AD_RESOURCE"`
}

// NewSession creates a new sessions based on settings in the AuthInfo
// this is to avoid having to create a new authorizer for multiple clients
func NewSession(a *AuthInfo, authByCLI bool) (*Session, error) {
	authorizer, err := a.NewResourceManagerAuthorizer(authByCLI)
	if err != nil {
		return nil, err
	}
	// create a new session
	session := &Session{
		SubscriptionID: a.SubscriptionID,
		Authorizer:     authorizer,
	}
	return session, nil
}

// NewResourceManagerAuthorizer retrieves an authorizer for the ResourceManagerEndpoint in the environment settings
func (a *AuthInfo) NewResourceManagerAuthorizer(authByCLI bool) (autorest.Authorizer, error) {
	var err error

	// get environment endpoints
	env, err := EnvironmentFromName(a.Environment)
	if err != nil {
		return nil, errors.New("No environment provided.")
	}

	// get oauth config
	resource := env.ResourceManagerEndpoint
	config, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, a.TenantID)
	if err != nil {
		return nil, err
	}

	// get adal token
	var adalToken adal.OAuthTokenProvider
	if !authByCLI {
		adalToken, err = adal.NewServicePrincipalToken(*config, a.ClientID, a.ClientSecret, resource)
		if err != nil {
			return nil, err
		}
	} else {
		cliToken, err := cli.GetTokenFromCLI(resource)
		if err != nil {
			return nil, err
		}
		token, err := cliToken.ToADALToken()
		if err != nil {
			return nil, err
		}
		adalToken = &token
	}

	return autorest.NewBearerAuthorizer(adalToken), nil
}

// this decode function is from:
// github.com/Azure/go-autorest/blob/master/autorest/azure/auth/auth.go
func decode(b []byte) ([]byte, error) {
	reader, enc := utfbom.Skip(bytes.NewReader(b))

	switch enc {
	case utfbom.UTF16LittleEndian:
		u16 := make([]uint16, (len(b)/2)-1)
		err := binary.Read(reader, binary.LittleEndian, &u16)
		if err != nil {
			return nil, err
		}
		return []byte(string(utf16.Decode(u16))), nil
	case utfbom.UTF16BigEndian:
		u16 := make([]uint16, (len(b)/2)-1)
		err := binary.Read(reader, binary.BigEndian, &u16)
		if err != nil {
			return nil, err
		}
		return []byte(string(utf16.Decode(u16))), nil
	}
	return ioutil.ReadAll(reader)
}

// GetAuthInfoFromFile extracts auth info from the file located at the AZURE_AUTH_LOCATION path
// it is similar to getAuthFile() in: github.com/Azure/go-autorest/blob/master/autorest/azure/auth/auth.go
func GetAuthInfoFromFile() (*AuthInfo, error) {
	fileLocation := os.Getenv("AZURE_AUTH_LOCATION")
	if fileLocation == "" {
		return nil, errors.New("environment variable AZURE_AUTH_LOCATION is not set")
	}

	contents, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return nil, err
	}

	// Auth file might be encoded
	decoded, err := decode(contents)
	if err != nil {
		return nil, err
	}

	authInfo := AuthInfo{}
	err = json.Unmarshal(decoded, &authInfo)
	if err != nil {
		return nil, err
	}

	return &authInfo, nil
}

// GetAuthInfoFromEnvVars allows a user to pass in auth info through environment variables instead of reading from file
func GetAuthInfoFromEnvVars() (*AuthInfo, error) {
	authInfo := AuthInfo{}
	err := env.Parse(&authInfo)
	if err != nil {
		return nil, err
	}
	return &authInfo, nil
}

func mergeAuthInfo(a, b *AuthInfo) *AuthInfo {
	// merge non-empty values
	// if a particular a, b field are both non-empty, then favor b

	if a == nil && b != nil {
		return b
	}
	if b == nil && a != nil {
		return a
	}
	if a == nil && b == nil {
		return b
	}

	elemA := reflect.ValueOf(a).Elem()
	elemB := reflect.ValueOf(b).Elem()
	typeOfB := elemB.Type()

	for i := 0; i < elemB.NumField(); i++ {
		valueFieldOfB := elemB.Field(i)
		fieldName := typeOfB.Field(i).Name

		if valueFieldOfB.String() == "" {
			valueFieldOfB.SetString(elemA.FieldByName(fieldName).String())
		}
	}

	return b
}

// GetAuthInfo gets auth info from both file and environment variables
// it merges the info and prioritizes non-empty values from the environment variables
func GetAuthInfo() (*AuthInfo, error) {
	authFromFile, errFromFile := GetAuthInfoFromFile()
	authFromEnvVars, errFromEnvVars := GetAuthInfoFromEnvVars()

	if errFromFile != nil && errFromEnvVars != nil {
		return nil, fmt.Errorf("Error retrieving auth info from file: %s\nError retrieving from auth info from environment variables: %s\n", errFromFile, errFromEnvVars)
	}

	return mergeAuthInfo(authFromFile, authFromEnvVars), nil
}
