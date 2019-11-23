package service

import (
	"fmt"

	"github.com/kraken/ui"
	apiv1 "github.com/kubekit/kubekit/api/kubekit/v1"
	"github.com/kubekit/kubekit/pkg/server"
	servicev1 "github.com/kubekit/kubekit/pkg/service/v1"
)

// ServicesForVersion returns the services for a given version
func ServicesForVersion(version string, dry bool, a ...interface{}) (server.Services, error) {
	switch version {
	case "v1":
		if len(a) < 1 {
			return nil, fmt.Errorf("missing parameters for version %s", version)
		}
		clustersPath := a[0].(string)
		parentUI := a[1].(*ui.UI)

		return server.Services{
			"v1.Kubekit": &server.Service{
				Version:                     version,
				Name:                        "Kubekit",
				ServiceRegister:             servicev1.NewKubeKitService(clustersPath, parentUI, dry),
				RegisterHandlerFromEndpoint: apiv1.RegisterKubekitHandlerFromEndpoint,
				SwaggerBytes:                apiv1.Swagger,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown version %s", version)
	}
}
