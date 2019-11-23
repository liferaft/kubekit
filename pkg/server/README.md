# Report

## Services running

| Service                                                  | Healthz | Healthz (5824)<br />`--healthz-port 5824` | Secure | Insecure<br />`--insecure` |
| -------------------------------------------------------- | ------- | ----------------------------------------- | ------ | -------------------------- |
| HTTP (5823) (*)<br />`--no-grpc`                         | na      | na                                        | na(3)  | na(1)                      |
| GRPC (5823)<br />`--no-http`                             | ok(2)   | ok(2)                                     | no(4)  | ok(2)                      |
| HTTP (5823) & GRPC (15823) (**)<br />`--port-grpc 15823` | ok(5)   | ok(5)                                     | no(6)  | ok(5)                      |
| HTTP/GRPC (5823)                                         | ok(8)   | ok(8)                                     | ok(8)  | ok(7)                      |

### TODO/FIX

* (4) : Send the CA certificate and use it at the client
* (6) : Create a certificate for gRPC using its port. GRPC does not work bc it's using port 15823 and cert is for port 5823. But, do we really need HTTP & GRPC running on different ports?

### Comments

(*)  : This option do not expose GRPC, but GRPC is required internally so the HTTP Gateway can use it. If this is what the user wants, then use a HTTP framework.
(**) : Why would be needed HTTP & GRPC needed in 2 different ports?

### Available Options

* `start server`: start server without options expose HTTP and gRPC API securely (TLS) on same port. HealthCheck also runs on same port.
* `--insecure`: expose HTTP or GRPC API insecurely, no TLS configured.
* `--no-http`: HTTP/REST API is not required, only GRPC. This option disable HealCheck on HTTP port, it is only listening on gRPC.
* `--healthz-port`: run health check in a different HTTP port. With `--no-http` this option does not apply, health check runs only on gRPC. Default to same port as HTTP/REST

## Create TLS Keys

```bash
openssl req -newkey rsa:2048 -nodes -keyout kubekit.key -x509 -days 365 -out kubekit.crt -subj "/C=US/ST=California/L=San Diego/O=LifeRaft/OU=KubeKit/CN=localhost:5823"
openssl req -newkey rsa:2048 -nodes -keyout kubekit-grpc.key -x509 -days 365 -out kubekit-grpc.crt -subj "/C=US/ST=California/L=San Diego/O=LifeRaft/OU=KubeKit/CN=localhost:15823"

rm -f kubekit{,-grpc}.{key,crt}
```

## Commands

1. HTTP (5823) Insecure:

  ```bash
  kubekit start server --debug --no-grpc --insecure [--healthz-port 5824]
  kubekitctl --no-grpc --insecure [--healthz-port 5824]
  curl -X GET http://localhost:5823/api/v1/version
  ```

1. GRPC (5823) Insecure:

  ```bash
  kubekit start server --debug --no-http --insecure [--healthz-port 5824]
  kubekitctl --no-http --insecure [--healthz-port 5824]
  ```
  
1. HTTP (5823) Secure:

  ```bash
  kubekit start server --debug --no-grpc [--healthz-port 5824]
  kubekitctl --no-grpc [--healthz-port 5824]
  curl -X GET -k https://localhost:5823/api/v1/version
  ```

1. GRPC (5823) Secure:

  ```bash
  kubekit start server --debug --no-http [--healthz-port 5824]
  kubekitctl --no-http [--healthz-port 5824]
  ```
  
1. HTTP (5823) & GRPC (15823) Insecure:

  ```bash
  kubekit start server --debug --grpc-port 15823 --insecure [--healthz-port 5824]
  kubekitctl --grpc-port 15823 --insecure [--healthz-port 5824]
  curl -X GET http://localhost:5823/api/v1/version
  ```
  
1. HTTP (5823) & GRPC (15823) Secure:

  ```bash
  kubekit start server --debug --port-grpc 15823 [--healthz-port 5824]
  kubekitctl [--healthz-port 5824]
  curl -X GET -k https://localhost:5823/api/v1/version
  ```

1. HTTP/GRPC (5823) Insecure:

  ```bash
  kubekit start server --debug --insecure [--healthz-port 5824]
  kubekitctl --insecure [--healthz-port 5824]
  curl -X GET http://localhost:5823/api/v1/version
  ```
  
1. HTTP/GRPC (5823) Secure:

  ```bash
  kubekit start server --debug [--healthz-port 5824]
  kubekitctl [--healthz-port 5824]
  curl -X GET -k https://localhost:5823/api/v1/version
  ```

## Debugging

Use the following env variables to identify gRPC errors at the client and server terminal.

```bash
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info
```
