package server

import (
	"fmt"
	"net"
	"strconv"
)

// GetFreePort returns an available port in the system
func GetFreePort() (string, error) {
	// addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	// if err != nil {
	// 	return 0, err
	// }

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port), nil
}

// IsPortFree checks if the given port is available to use
func IsPortFree(port string) bool {
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if nil == err {
		l.Close()
	}
	return err == nil
}

// getPorts returns the port to be used by multiplexer, gRPC and HTTP/REST given
// a set of desired ports.
// The rules are:
// (0) At least a port `port` should be given and a good and available port.
// (1) One port (`port`): Means the all services running on the same port, so there
//     will be Mux, gRPC and HTTP on `port`.
// (2) Two Ports (`port` & `ports[0]`) equal: Simmilar to rule #1.
// (3) Two Ports (`port` & `ports[0]`) different: Means the 2 services running on
//     different ports. The first port `port` is for HTTP and second port `ports[0]`
//     is for gRPC. There is no Multimplex service neither port.
// (4) Any other port (`ports[2]`...) should be ignored
func getPorts(port string, ports ...string) (portMux, portGRPC, portHTTP string, err error) {
	ok := func(prt string) bool {
		return prt != "" && prt != "0" && IsPortFree(prt)
	}

	// Rule #0: if port is not good, return an error
	if !ok(port) {
		return "", "", "", fmt.Errorf("port %q is invalid or in use", port)
	}

	if len(ports) == 0 {
		return port, port, port, nil
	}

	if port == ports[0] {
		return port, port, port, nil
	}

	return "", port, ports[0], nil

	// The rules are:
	// (0) At least a port `port` should be given and a good and available port.
	// (1) One port (`port`): Means the 2 services running on the same port, so there
	//     will be MP on `port`, gRPC and HTTP on 1`port` and 2`port` or an available one.
	// (2) Two Ports (`port` & `ports[0]`) equal: Simmilar to rule #1.
	// (3) Two Ports (`port` & `ports[0]`) different: Means the 2 services running on
	//     different ports. The first port `port` is for HTTP and second port `ports[0]`
	//     is for gRPC. There is no Multimplex service neither port.
	// (4) Three Ports (`port`, `ports[0]` & `ports[1]`), all of them equal: Simmilar to rule #1
	// (5) Three Ports (`port`, `ports[0]` & `ports[1]`), two of them equal: Simmilar to rule #3
	// (6) Three Ports (`port`, `ports[0]` & `ports[1]`), all of them different: Use them in the following order for these services:
	//       Multiplex, gRPC, HTTP. If any of these ports is invalid or in use, return an error
	// (7) Any other port (`ports[2]`...) should be ignored

	// // if `prt` is not ok use the alternative `altPrt`, if the alternative is not ok get a free one
	// getPort := func(prt, altPrt string) (string, error) {
	// 	if ok(prt) {
	// 		return prt, nil
	// 	}
	// 	if ok(altPrt) {
	// 		return altPrt, nil
	// 	}
	// 	return GetFreePort()
	// }

	// rule01 := func(port string) (portMux, portGRPC, portHTTP string, err error) {
	// 	portMux = port
	// 	if portGRPC, err = getPort("1"+port, ""); err != nil {
	// 		return "", "", "", err
	// 	}
	// 	if portHTTP, err = getPort("2"+port, ""); err != nil {
	// 		return "", "", "", err
	// 	}
	// 	return portMux, portGRPC, portHTTP, err
	// }

	// rule03 := func(port1, port2 string) (portMux, portGRPC, portHTTP string, err error) {
	// 	portMux = ""
	// 	portHTTP = port1
	// 	if !ok(port2) {
	// 		return "", "", "", fmt.Errorf("port %q is invalid or in use", port2)
	// 	}
	// 	portGRPC = port2
	// 	return portMux, portGRPC, portHTTP, nil
	// }

	// rule06 := func(port1, port2, port3 string) (portMux, portGRPC, portHTTP string, err error) {
	// 	portMux = port1
	// 	if !ok(port2) {
	// 		return "", "", "", fmt.Errorf("port %q is invalid or in use", port2)
	// 	}
	// 	portGRPC = port2
	// 	if !ok(port3) {
	// 		return "", "", "", fmt.Errorf("port %q is invalid or in use", port3)
	// 	}
	// 	portHTTP = port3
	// 	return portMux, portGRPC, portHTTP, nil
	// }

	// // from rule #0, port is a good port. Each rule validate the 2nd and 3rd port
	// switch len(ports) {
	// case 0:
	// 	return rule01(port)
	// case 1:
	// 	if port == ports[0] {
	// 		return rule01(port) // rule #2
	// 	}
	// 	return rule03(port, ports[0])
	// default:
	// 	if (port == ports[0]) && (port == ports[1]) {
	// 		return rule01(port) // rule #4
	// 	}
	// 	if (port != ports[0]) && (port != ports[1]) && (ports[0] != ports[1]) {
	// 		return rule06(port, ports[0], ports[1]) // rule #6
	// 	}
	// 	// rule #5
	// 	if port == ports[0] {
	// 		return rule03(port, ports[0])
	// 	}
	// 	if port == ports[1] {
	// 		return rule03(port, ports[1])
	// 	}
	// 	// (ports[0] == ports[1]) {
	// 	return rule03(ports[0], ports[1])
	// }

	// // var err1, err2 error
	// // switch len(ports) {
	// // case 0: // if only received a port, set gRPC to "1"+portMux and HTTP to "2"+portMux or get free one if not available
	// // 	portGRPC, err1 = getPort("1"+portMux, "")
	// // 	portHTTP, err2 = getPort("2"+portMux, "")
	// // case 1: // if received port for MP and gRPC: ...
	// // 	// ... use the gRPC given one, or "1"+portMux if not good, or a free one. ...
	// // 	if ports[0] == portMux {
	// // 		portGRPC, err1 = getPort("1"+portMux, "")
	// // 	} else {
	// // 		portGRPC, err1 = getPort(ports[0], "1"+portMux)
	// // 	}
	// // 	portHTTP, err2 = getPort("2"+portMux, "") // ... and set HTTP to "2"+portMux or get free one if not available.
	// // default: // if received ports for MP, gRPC and HTTP (maybe more but those are ignored): ...
	// // 	// ... use the gRPC given one, or "1"+portMux if not good, or a free one. ...
	// // 	if ports[0] == portMux {
	// // 		portGRPC, err1 = getPort("1"+portMux, "")
	// // 	} else {
	// // 		portGRPC, err1 = getPort(ports[0], "1"+portMux)
	// // 	}
	// // 	// ... use the HTTP given one, or "2"+portMux if not good, or a free one.
	// // 	if ports[1] == portMux || ports[1] == portGRPC {
	// // 		portHTTP, err2 = getPort("2"+portMux, "")
	// // 	} else {
	// // 		portHTTP, err2 = getPort(ports[1], "2"+portMux)
	// // 	}
	// // }

	// // var errStr string
	// // if err1 != nil {
	// // 	errStr = fmt.Sprintf("failed to get a good port for gRPC (%s)", err1)
	// // 	if err2 != nil {
	// // 		errStr += fmt.Sprintf(" and HTTP/REST (%s)", err2)
	// // 	}
	// // } else if err2 != nil {
	// // 	errStr = fmt.Sprintf("failed to get a good port for HTTP/REST (%s)", err2)
	// // }

	// // if errStr != "" {
	// // 	err = fmt.Errorf(errStr)
	// // }

	// // return portMux, portGRPC, portHTTP, err
}
