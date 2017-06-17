package main

import "os"
import "fmt"
import "log"
import "flag"
import "regexp"
import "net/http"
import "github.com/gorilla/websocket"
import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/routes"
import "github.com/dadleyy/beacon.api/beacon/security"

func main() {
	flags := struct {
		port string
	}{}

	websocketUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     security.AnyOrigin,
	}

	routes := net.RouteList{
		net.RouteConfig{"GET", regexp.MustCompile("/system")}:   routes.System,
		net.RouteConfig{"GET", regexp.MustCompile("/register")}: routes.Register,
	}

	runtime := net.ServerRuntime{
		Upgrader:  websocketUpgrader,
		RouteList: routes,
		Logger:    log.New(os.Stdout, "beacon", log.Ldate|log.Ltime|log.Lshortfile),
	}

	flag.StringVar(&flags.port, "port", "12345", "the port to attach the http listener to")
	flag.Parse()

	if valid := len(flags.port) >= 1; !valid {
		fmt.Printf("invalid port: %s", flags.port)
		flag.PrintDefaults()
		return
	}

	fmt.Printf("starting server on port: %s\n", flags.port)
	if e := http.ListenAndServe("0.0.0.0:12345", &runtime); e != nil {
		fmt.Printf("unable to start server: %s", e.Error())
	}
}
