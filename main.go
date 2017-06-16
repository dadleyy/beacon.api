package main

import "fmt"
import "flag"
import "net/http"
import "github.com/gorilla/websocket"
import "github.com/dadleyy/pylite.api/api/server"
import "github.com/dadleyy/pylite.api/api/security"

func main() {
	runtime, port := server.Runtime{
		Upgrader: websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024, CheckOrigin: security.AnyOrigin},
	}, ""

	flag.StringVar(&port, "port", "12345", "the port to attach the http listener to")
	flag.Parse()

	if valid := len(port) >= 1; !valid {
		fmt.Printf("invalid port: %s", port)
		return
	}

	fmt.Printf("starting server on port: %s\n", port)
	if e := http.ListenAndServe("0.0.0.0:12345", &runtime); e != nil {
		fmt.Printf("unable to start server: %s", e.Error())
	}
}
