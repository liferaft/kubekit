package azure

import (
	"fmt"
	"strings"

	"github.com/Azure/go-autorest/autorest/azure"
)

// EnvironmentFromName gets the environment settings by the environment name
// it also supports additional names that the azure sdk does not
func EnvironmentFromName(name string) (azure.Environment, error) {
	var env azure.Environment
	var err error
	switch strings.ToUpper(name) {
	case "CHINA", "AZURECHINACLOUD":
		env = azure.ChinaCloud
	case "GERMAN", "AZUREGERMANCLOUD":
		env = azure.GermanCloud
	case "", "PUBLIC", "AZUREPUBLICCLOUD":
		env = azure.PublicCloud
	case "USGOVERNMENT", "AZUREUSGOVERNMENTCLOUD":
		env = azure.USGovernmentCloud
	default:
		err = fmt.Errorf("No cloud environment matching the name %q", name)
		//env = azure.PublicCloud
	}
	return env, err
}
