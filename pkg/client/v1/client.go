package v1

import (
	"github.com/johandry/log"
	apiv1 "github.com/liferaft/kubekit/api/kubekit/v1"
	"github.com/liferaft/kubekit/pkg/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Config is the kubekit client configuration
type Config struct {
	*client.Config

	GrpcClient apiv1.KubekitClient
}

// New creates a new KubeKit Client
func New(apiVersion string, logger *log.Logger) *Config {
	c := client.New(apiVersion, logger)
	return &Config{
		Config: c,
	}
}

// WithGRPC makes the KubeKit client to have a GRPC client
func (c *Config) WithGRPC(host, port string) error {
	grpcConn, err := c.GetGRPCConn(host, port)
	if err != nil {
		return grpc.Errorf(codes.Internal, "cannot connect to %s:%s. %s", host, port, err)
	}
	// defer grpcConn.Close()
	grpcClient := apiv1.NewKubekitClient(grpcConn)

	c.GrpcClient = grpcClient
	c.GrpcConn = grpcConn
	c.GrpcPort = port

	if c.Host == "" {
		c.Host = host
	}

	return nil
}
