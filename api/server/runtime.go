package server

import "io"
import "fmt"
import "bytes"
import "net/http"
import "github.com/gorilla/websocket"

type Runtime struct {
	websocket.Upgrader
}

func (r *Runtime) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	_, e := r.Upgrade(responseWriter, request, nil)

	if e != nil {
		fmt.Printf("[request] could not upgrate request: %s\n", e.Error())
		return
	}

	fmt.Printf("[request] %s %s\n", request.URL.Path, request.URL.Host)
	buf := bytes.NewBuffer([]byte("hello world"))
	responseWriter.WriteHeader(200)
	io.Copy(responseWriter, buf)
}
