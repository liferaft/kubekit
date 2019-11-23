package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	toml "github.com/pelletier/go-toml"
	"google.golang.org/grpc/health/grpc_health_v1"
	yaml "gopkg.in/yaml.v2"
)

// HealthCheckResponse encapsulate the server response of the Health Check
type HealthCheckResponse struct {
	StatusCode     int `json:"code" yaml:"code" toml:"code"`
	httpStatusCode int
	Status         string `json:"status" yaml:"status" toml:"status"`
	Message        string `json:"message" yaml:"message" toml:"message"`
	Service        string `json:"service" yaml:"service" toml:"service"`
}

// HealthzResponse creates a HealthCheck Response. Uses a default message if the message is not provided
func HealthzResponse(code int, service string, message ...string) *HealthCheckResponse {
	var msg string
	var httpStatusCode int
	switch code {
	case int(grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN):
		msg = fmt.Sprintf("unknown service %q", service)
		code = int(grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN)
		httpStatusCode = http.StatusInternalServerError
	case int(grpc_health_v1.HealthCheckResponse_NOT_SERVING), http.StatusServiceUnavailable:
		msg = fmt.Sprintf("service %q is not serving", service)
		code = int(grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		httpStatusCode = http.StatusServiceUnavailable
	case int(grpc_health_v1.HealthCheckResponse_SERVING), http.StatusOK:
		msg = fmt.Sprintf("service %q is serving", service)
		code = int(grpc_health_v1.HealthCheckResponse_SERVING)
		httpStatusCode = http.StatusOK
	case int(grpc_health_v1.HealthCheckResponse_UNKNOWN):
		msg = fmt.Sprintf("unknown the status of service %q", service)
		code = int(grpc_health_v1.HealthCheckResponse_UNKNOWN)
		httpStatusCode = http.StatusInternalServerError
	default:
		msg = fmt.Sprintf("unknown health check status code %d", code)
		code = int(grpc_health_v1.HealthCheckResponse_UNKNOWN)
		httpStatusCode = http.StatusInternalServerError
	}
	if len(message) != 0 {
		msg = message[0]
	}
	status := grpc_health_v1.HealthCheckResponse_ServingStatus(code).String()
	return &HealthCheckResponse{
		StatusCode:     code,
		httpStatusCode: httpStatusCode,
		Status:         status,
		Message:        msg,
		Service:        service,
	}
}

// HealthzResponsef creates an error with a formatted error message
func HealthzResponsef(code int, service, message string, a ...interface{}) *HealthCheckResponse {
	msg := fmt.Sprintf(message, a...)
	return HealthzResponse(code, service, msg)
}

func (r *HealthCheckResponse) String() string {
	data, err := r.JSON(false)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

// Stringf returns the API Error in the given format: json, yaml or toml
func (r *HealthCheckResponse) Stringf(format string) string {
	var data []byte
	var err error

	switch format {
	case "yaml":
		data, err = r.YAML()
	case "json":
		data, err = r.JSON(false)
	case "toml":
		data, err = r.TOML()
	default:
		err = fmt.Errorf("can't stringify the API Error, unknown format %q", format)
	}

	if err != nil {
		return err.Error()
	}
	return string(data)
}

// YAML returns the API Error in YAML format
func (r *HealthCheckResponse) YAML() ([]byte, error) {
	return yaml.Marshal(r)
}

// JSON returns the API Error in JSON format
func (r *HealthCheckResponse) JSON(pp bool) ([]byte, error) {
	if pp {
		return json.MarshalIndent(r, "", "  ")
	}
	return json.Marshal(r)
}

// TOML returns the API Error in TOML format
func (r *HealthCheckResponse) TOML() ([]byte, error) {
	return toml.Marshal(r)
}

// Write writes the error to the given http response writer
func (r *HealthCheckResponse) Write(w http.ResponseWriter) {
	// If it's an error, return an error
	if r.httpStatusCode == http.StatusInternalServerError {
		Error(http.StatusInternalServerError, "healthz", r.Message).Write(w)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(r.httpStatusCode)
	fmt.Fprintln(w, r.String())
}
