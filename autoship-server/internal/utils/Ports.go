package utils

import (
	"net"
)

// GetFreePort asks the OS for a free open port that is ready to use.
func GetFreePort() (int, error) {
	listener, err := net.Listen("tcp", ":0") // :0 asks the OS to assign an available port
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
