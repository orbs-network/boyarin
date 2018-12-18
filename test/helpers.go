package test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

func LocalIP() string {
	ifaces, _ := net.Interfaces()

	for _, i := range ifaces {
		if addrs, err := i.Addrs(); err == nil {
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				if ip != nil && ip.To4() != nil && ip.To4().String() != "127.0.0.1" {
					return ip.To4().String()
				}
			}
		}
	}

	return "127.0.0.1"
}

func NodeAddresses() []string {
	return []string{
		"a328846cd5b4979d68a8c58a9bdfeee657b34de7",
		"d27e2e7398e2582f63d0800330010b3e58952ff6",
		"6e2cb55e4cbe97bf5b1e731d51cc2c285d83cbf9",
	}
}

const eventuallyIterations = 25

func Eventually(timeout time.Duration, f func() bool) bool {
	for i := 0; i < eventuallyIterations; i++ {
		if f() {
			return true
		}
		time.Sleep(timeout / eventuallyIterations)
	}
	return false
}

type HttpServer interface {
	Start()
	Shutdown()
	Url() string
	Port() int
}

type httpServer struct {
	path     string
	listener net.Listener
	server   *http.Server
}

func (h *httpServer) Start() {
	go h.server.Serve(h.listener)
}

func (h *httpServer) Shutdown() {
	h.server.Shutdown(context.TODO())
}

func (h *httpServer) Url() string {
	return fmt.Sprintf("http://127.0.0.1:%d%s", h.listener.Addr().(*net.TCPAddr).Port, h.path)
}

func (h *httpServer) Port() int {
	return h.listener.Addr().(*net.TCPAddr).Port
}

func CreateHttpServer(path string, f func(writer http.ResponseWriter, request *http.Request)) HttpServer {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		panic(err)
	}

	router := http.NewServeMux()
	router.HandleFunc(path, f)

	server := &http.Server{
		Handler: router,
	}

	return &httpServer{
		path:     path,
		listener: listener,
		server:   server,
	}
}
