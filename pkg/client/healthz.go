package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// Healthz returns the Health Status using HTTP/REST or gRPC
func (c *Config) Healthz(ctx context.Context, service string) (string, error) {
	c.Logger.Debugf("Checking Health Status for service %q", service)

	return c.RunGRPCnRESTFunc("healthz", false,
		func() (string, error) {
			return c.HealthzGRPC(ctx, service)
		},
		func() (string, error) {
			return c.HealthzHTTP(service)
		})
}

// HealthzGRPC returns the Health Status using gRPC
func (c *Config) HealthzGRPC(ctx context.Context, service string) (string, error) {
	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	c.Logger.Debugf(`grpcurl request: grpcurl -insecure -d '{"service": "%s"}' %s:%s grpc.health.v1.Health/Check"`, service, c.Host, c.GrpcPort)

	resp, err := grpc_health_v1.NewHealthClient(c.GrpcConn).Check(childCtx, &grpc_health_v1.HealthCheckRequest{Service: service})
	if err != nil {
		if stat, ok := status.FromError(err); ok && stat.Code() == codes.Unimplemented {
			return grpc_health_v1.HealthCheckResponse_UNKNOWN.String(), fmt.Errorf("this server does not implement the grpc health protocol (grpc.health.v1.Health)")
		}
		return grpc_health_v1.HealthCheckResponse_UNKNOWN.String(), fmt.Errorf("health rpc failed: %+v", err)
	}
	return resp.GetStatus().String(), nil
}

// HealthzHTTP returns the Health Status using HTTP/REST
func (c *Config) HealthzHTTP(service string) (string, error) {
	service = strings.Replace(service, ".", "/", -1)
	healthzURL := fmt.Sprintf("%s/healthz/%s", c.HTTPHealthzBaseURL, service)

	c.Logger.Debugf(`curl request: curl -s -k -X GET "%s"`, healthzURL)

	resp, err := c.HTTPHealthzClient.Get(healthzURL)
	if err != nil {
		return grpc_health_v1.HealthCheckResponse_UNKNOWN.String(), err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return grpc_health_v1.HealthCheckResponse_SERVING.String(), nil
	}
	if resp.StatusCode == http.StatusServiceUnavailable {
		return grpc_health_v1.HealthCheckResponse_NOT_SERVING.String(), nil
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return grpc_health_v1.HealthCheckResponse_UNKNOWN.String(), fmt.Errorf("unknown service status from HTTP status code %d (%q), Body: %q, URL: %s", resp.StatusCode, resp.Status, body, healthzURL)
}
