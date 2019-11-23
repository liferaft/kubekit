package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-03-01/resources"
	"github.com/Azure/go-autorest/autorest/azure"
)

const (
	ProviderRegisteredState = "Registered"
)

// ProvidersClient returns a network.InterfacesClient
func ProvidersClient(environment azure.Environment, session *Session) (*resources.ProvidersClient, error) {
	client := resources.NewProvidersClientWithBaseURI(environment.ResourceManagerEndpoint, session.SubscriptionID)
	client.Authorizer = session.Authorizer
	return &client, nil
}

// ProvidersClientByEnvStr returns a network.InterfacesClient by looking up the environment by name
func ProvidersClientByEnvStr(environmentName string, session *Session) (*resources.ProvidersClient, error) {
	env, err := EnvironmentFromName(environmentName)
	if err != nil {
		return nil, err
	}
	return ProvidersClient(env, session)
}

// RegisterProvider registers the provider
func RegisterProvider(providersClient *resources.ProvidersClient, resourceProviderNamespace string) (resources.Provider, error) {
	return providersClient.Register(context.Background(), resourceProviderNamespace)
}

// UnregisterProvider unregisters the provider
func UnregisterProvider(providersClient *resources.ProvidersClient, resourceProviderNamespace string) (resources.Provider, error) {
	return providersClient.Unregister(context.Background(), resourceProviderNamespace)
}

// GetProvider gets the provider
func GetProvider(providersClient *resources.ProvidersClient, resourceProviderNamespace, expand string) (resources.Provider, error) {
	return providersClient.Get(context.Background(), resourceProviderNamespace, expand)
}

// IsProviderRegistered indicates if a provider is registered
func IsProviderRegistered(providersClient *resources.ProvidersClient, resoureProviderNamespace string) (bool, error) {
	result, err := GetProvider(providersClient, resoureProviderNamespace, "")
	if err != nil {
		return false, err
	}
	if *result.RegistrationState == ProviderRegisteredState {
		return true, nil
	}
	return false, err
}
