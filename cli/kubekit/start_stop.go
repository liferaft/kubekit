package kubekit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/liferaft/kubekit/pkg/server"
	"github.com/liferaft/kubekit/pkg/service"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Hidden: true,
	Use:    "start [cluster] NAME[,NAME ...]",
	Short:  "Starts a cluster or nodes",
	Long: `Starts a cluster, a single or multiple nodes of a cluster filtered by node name,
IP, DNS or by the pool name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'start' still not implemented")
		return nil
	},
}

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Hidden: true,
	Use:    "stop [cluster] NAME[,NAME ...]",
	Short:  "Stop a cluster or nodes",
	Long: `Stops a cluster, a single or multiple nodes of a cluster filtered by node name,
IP, DNS or by the pool name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'stop' still not implemented")
		return nil
	},
}

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Hidden: true,
	Use:    "restart [cluster] NAME[,NAME ...]",
	Short:  "Restart a cluster or nodes",
	Long: `Restarts a cluster, a single or multiple nodes of a cluster filtered by node
name, IP, DNS or by the pool name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'restart' still not implemented")
		return nil
	},
}

// startClusterCmd represents the start custer command
var startClusterCmd = &cobra.Command{
	Hidden: true,
	Use:    "cluster NAME[,NAME ...]",
	Short:  "Starts a cluster or nodes",
	Long: `Starts a cluster, a single or multiple nodes of a cluster filtered by node name,
IP, DNS or by the pool name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'start cluster' still not implemented")
		return nil
	},
}

// stopClusterCmd represents the stop cluster command
var stopClusterCmd = &cobra.Command{
	Hidden: true,
	Use:    "cluster NAME[,NAME ...]",
	Short:  "Stop a cluster or nodes",
	Long: `Stops a cluster, a single or multiple nodes of a cluster filtered by node name,
IP, DNS or by the pool name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'stop cluster' still not implemented")
		return nil
	},
}

// restartCmd represents the restart cluster command
var restartClusterCmd = &cobra.Command{
	Hidden: true,
	Use:    "cluster NAME[,NAME ...]",
	Short:  "Restart a cluster or nodes",
	Long: `Restarts a cluster, a single or multiple nodes of a cluster filtered by node
name, IP, DNS or by the pool name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'restart cluster' still not implemented")
		return nil
	},
}

// startServerCmd represents the start server command
var startServerCmd = &cobra.Command{
	Hidden: true,
	Use:    "server",
	Short:  "Starts the KubeKit server",
	Long:   `Starts the KubeKit server to expose a REST API and a gRPC API.`,
	RunE:   startServerRun,
}

// stopServerCmd represents the stop server command
var stopServerCmd = &cobra.Command{
	Hidden: true,
	Use:    "server",
	Short:  "Stop the KubeKit server",
	Long:   `Stops the KubeKit server to stop exposing the REST API and the gRPC API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'stop server' still not implemented")
		return nil
	},
}

// restartServerCmd represents the restart server command
var restartServerCmd = &cobra.Command{
	Hidden: true,
	Use:    "server",
	Short:  "Restart the KubeKit server",
	Long:   `Stop and starts the KubeKit server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[ERROR] command 'restart server' still not implemented")
		return nil
	},
}

func addStartStopCmd() {
	// start [cluster] NAME[,NAME ...]
	RootCmd.AddCommand(startCmd)
	startCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes to start")
	startCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools where are the nodes to start")
	startCmd.AddCommand(startClusterCmd)
	startClusterCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes to start")
	startClusterCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools where are the nodes to start")
	// stop [cluster] NAME[,NAME ...]
	RootCmd.AddCommand(stopCmd)
	stopCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes to stop")
	stopCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools where are the nodes to stop")
	stopCmd.AddCommand(stopClusterCmd)
	stopClusterCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes to stop")
	stopClusterCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools where are the nodes to stop")
	// restart [cluster] NAME[,NAME ...]
	RootCmd.AddCommand(restartCmd)
	restartCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes to restart")
	restartCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools where are the nodes to restart")
	restartCmd.AddCommand(restartClusterCmd)
	restartClusterCmd.Flags().StringSliceP("nodes", "n", nil, "list of nodes to restart")
	restartClusterCmd.Flags().StringSliceP("pools", "p", nil, "list of node pools where are the nodes to restart")

	// start server --host 0.0.0.0 --port 5823 --healthz-port 5824
	// 							--cert-dir $KUBEKIT_HOME/server/pki
	// 							--tls-cert-file /path/to/my/ca/certs/kubekit.cert
	// 							--tls-private-key-file /path/to/my/ca/certs/kubekit.key
	// 							--ca-file /path/to/my/ca/certs/client-ca.key
	// 							--insecure --allow-cors
	startCmd.AddCommand(startServerCmd)
	startServerCmd.Flags().String("host", defServerHost, "The hostname or IP address for KubeKit Server to serve on.")
	startServerCmd.Flags().Int("port", defServerPort, "The port for the KubeKit Server to serve on.")
	startServerCmd.Flags().Int("grpc-port", 0, "The port for the KubeKit gRPC Server to serve on. This variable will be deprecated when gRPC and HTTP/REST serve on same port.")
	startServerCmd.Flags().Bool("no-http", false, "Do not start the HTTP/REST service. The server expose the API only on gRPC")
	startServerCmd.Flags().Int("healthz-port", defHealthzPort, "The port of the localhost healthz endpoint.")
	startServerCmd.Flags().String("cert-dir", "", "The directory where the TLS certs are located. If --tls-cert-file and --tls-private-key-file are provided, this flag will be ignored.")
	startServerCmd.Flags().String("tls-cert-file", "", "File containing x509 Certificate used for serving HTTPS. If --tls-cert-file and --tls-private-key-file are not provided, a self-signed certificate and key are generated for the public address and saved to the directory passed to --cert-dir.")
	startServerCmd.Flags().String("tls-private-key-file", "", "File containing x509 private key matching --tls-cert-file.")
	startServerCmd.Flags().String("ca-file", "", "If set, any request presenting a client certificate signed by one of the authorities in the ca-file is authenticated with an identity corresponding to the CommonName of the client certificate.")
	startServerCmd.Flags().Bool("insecure", false, "Do not use TLS to provide security to the server to expose the API gRPC and HTTP/REST")
	startServerCmd.Flags().Bool("allow-cors", false, "Make the server to allow Cross-Origin Resource Sharing (CORS). Usefull for developement & testing")
	startServerCmd.Flags().Bool("dry-run", false, "Makes API calls inert and returns generic output for testing purposes")

	// stop server
	stopCmd.AddCommand(stopServerCmd)

	// restart server
	restartCmd.AddCommand(restartServerCmd)
}

func startServerRun(cmd *cobra.Command, args []string) error {
	host := cmd.Flags().Lookup("host").Value.String()
	// TODO: validation using TCP IP functions

	port := cmd.Flags().Lookup("port").Value.String()
	// Do not need to parse or validate the value, Cobra does it for us.
	// TODO: When gRPC and HTTP/REST serve in the same port, the portHTT may disapear
	portGRPC := cmd.Flags().Lookup("grpc-port").Value.String()
	// Do not need to parse or validate the value, Cobra does it for us.
	if portGRPC == "0" {
		portGRPC = port
	}
	// TODO: validation using TCP IP functions

	healthzPort := cmd.Flags().Lookup("healthz-port").Value.String()
	// Do not need to parse or validate the value, Cobra does it for us.
	// TODO: validation using TCP IP functions

	var err error
	certDir := cmd.Flags().Lookup("cert-dir").Value.String()
	if certDir, err = validCertDir(certDir); err != nil {
		return err
	}

	tlsCertFile := cmd.Flags().Lookup("tls-cert-file").Value.String()
	if tlsCertFile, err = validFile(certDir, tlsCertFile); err != nil {
		return err
	}

	tlsPrivateKeyFile := cmd.Flags().Lookup("tls-private-key-file").Value.String()
	if tlsPrivateKeyFile, err = validFile(certDir, tlsPrivateKeyFile); err != nil {
		return err
	}

	caFile := cmd.Flags().Lookup("ca-file").Value.String()
	if caFile, err = validFile(certDir, caFile); err != nil {
		return err
	}

	insecure := cmd.Flags().Lookup("insecure").Value.String() == "true"

	noHTTP := cmd.Flags().Lookup("no-http").Value.String() == "true"

	allowCORS := cmd.Flags().Lookup("allow-cors").Value.String() == "true"

	dry := cmd.Flags().Lookup("dry-run").Value.String() == "true"

	// DEBUG:
	// var insecureFlag string
	// if insecure {
	// 	insecureFlag = " --insecure"
	// }
	// var noHTTPFlag string
	// if noHTTP {
	// 	noHTTPFlag = " --no-http"
	// }
	// var corsFlag string
	// if allowCORS {
	// 	corsFlag = " --allow-cors"
	// }
	// cmd.Printf("start server --host %s --port %s --grpc-port %s --healthz-port %s --cert-dir %q --tls-cert-file %q --tls-private-key-file %q --ca-file %q %s %s %s\n", host, port, portGRPC, healthzPort, certDir, tlsCertFile, tlsPrivateKeyFile, caFile, insecureFlag, noHTTPFlag, corsFlag)

	v1Services, err := service.ServicesForVersion("v1", dry, config.ClustersDir(), config.UI)
	if err != nil {
		return err
	}
	ctx := context.Background()

	s := server.New(ctx, "kubekit", host, config.UI).
		AddServices(v1Services).
		WithSwagger().
		WithHealthCheck(healthzPort).
		SetCORS(allowCORS).
		SetTLS(!insecure, certDir, tlsCertFile, tlsPrivateKeyFile, caFile)

	// if allowCORS {
	// 	s.AllowCORS()
	// }

	// if !insecure {
	// 	s.WithTLS(certDir, tlsCertFile, tlsPrivateKeyFile, caFile)
	// }

	// if no-http only start gRPC
	// if grpc-port is set, start HTTP on port `port` and gRPC on port `portGRPC`
	// if grpc-port is not set, start HTTP and gRPC on same port `port` using Mux
	if noHTTP {
		s.StartGRPC(port)
	} else {
		s.Start(port, portGRPC)
	}

	return s.Wait().Err()
}

// addHomeDir prefix the given path the user home directory if the path starts with `~`
func addHomeDir(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := homedir.Dir()
		if err != nil {
			panic(fmt.Errorf("cannot find home directory. %s", err))
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// filepathAbsTo prefix the base directory to the given path if it is not an
// absolute directory
func filepathAbsTo(base, path string) string {
	if len(path) != 0 && !filepath.IsAbs(path) {
		return filepath.Join(base, path)
	}
	return path
}

func existPath(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

// validCertDir returns a valid certificates directory
func validCertDir(path string) (string, error) {
	// If empty, use the default directory for certificates
	// If it's the default directory for certs, append the KK home dir and make
	// sure it exists
	if len(path) == 0 || path == defServerPKIDir {
		path = config.PKIDir()
		if !existPath(path) {
			os.MkdirAll(path, 0700)
		}
	}
	// Replace `~` with the home directory
	path = addHomeDir(path)

	// If it's absolute (and not the default dir) append the KK home dir
	if filepath.IsAbs(path) {
		path = filepathAbsTo(config.Dir(), path)
	}

	if !existPath(path) {
		return "", fmt.Errorf("cannot find the given certificates directory %q", path)
	}
	return path, nil
}

func validFile(basedir, filename string) (string, error) {
	if len(filename) == 0 {
		return "", nil
	}
	// If it's relative make it absolute to basedir
	filename = filepathAbsTo(basedir, filename)
	// Replace `~` with the home directory
	filename = addHomeDir(filename)

	if !existPath(filename) {
		return "", fmt.Errorf("file not found %q", filename)
	}
	return filename, nil
}
