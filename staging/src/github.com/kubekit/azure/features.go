package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2015-12-01/features"
	"github.com/Azure/go-autorest/autorest/azure"
)

const (
	FeatureRegisteredState  = "Registered"
	FeatureRegisteringState = "Registering"
)

// FeaturesClient returns a network.InterfacesClient
func FeaturesClient(environment azure.Environment, session *Session) (*features.Client, error) {
	client := features.NewClientWithBaseURI(environment.ResourceManagerEndpoint, session.SubscriptionID)
	client.Authorizer = session.Authorizer
	return &client, nil
}

// FeaturesClientByEnvStr returns a network.InterfacesClient by looking up the environment by name
func FeaturesClientByEnvStr(environmentName string, session *Session) (*features.Client, error) {
	env, err := EnvironmentFromName(environmentName)
	if err != nil {
		return nil, err
	}
	return FeaturesClient(env, session)
}

// RegisterFeature registers the preview feature
// Note: there is no deregister provided in the sdk
func RegisterFeature(featuresClient *features.Client, resourceProviderNamespace, featureName string) (features.Result, error) {
	return featuresClient.Register(context.Background(), resourceProviderNamespace, featureName)
}

// GetFeature gets the preview feature
func GetFeature(featuresClient *features.Client, resourceProviderNamespace, featureName string) (features.Result, error) {
	return featuresClient.Get(context.Background(), resourceProviderNamespace, featureName)
}

// IsFeatureRegistered indicates if a feature is registered
func IsFeatureRegistered(featuresClient *features.Client, resoureProviderNamespace, featureName string) (bool, error) {
	result, err := GetFeature(featuresClient, resoureProviderNamespace, featureName)
	if err != nil {
		return false, err
	}
	if *result.Properties.State == FeatureRegisteredState {
		return true, nil
	}
	return false, err
}
