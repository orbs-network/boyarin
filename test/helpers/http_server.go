package helpers

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

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

func CreateHttpServer(path string, port int, f func(writer http.ResponseWriter, request *http.Request)) HttpServer {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
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
