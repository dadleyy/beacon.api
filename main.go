package main

import "os"
import "io"
import "fmt"
import "log"
import "flag"
import "sync"
import "regexp"
import "net/http"
import "github.com/gorilla/websocket"

import "github.com/dadleyy/beacon.api/beacon/bg"
import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/routes"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/security"

func main() {
	flags := struct {
		port string
	}{}

	logger := log.New(os.Stdout, "beacon ", log.Ldate|log.Ltime|log.Lshortfile)

	websocketUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     security.AnyOrigin,
	}

	backgroundChannels := defs.BackgroundChannels{
		defs.DeviceControlChannelName: make(chan io.Reader, 10),
	}

	registrationStream := make(chan *device.Connection, 10)

	control := bg.DeviceControlProcessor{
		Logger:        log.New(os.Stdout, "device control ", log.Ldate|log.Ltime|log.Lshortfile),
		ControlStream: backgroundChannels[defs.DeviceControlChannelName],
		Registrations: registrationStream,
	}

	processors := []bg.Processor{
		&control,
	}

	routes := net.RouteList{
		net.RouteConfig{"GET", regexp.MustCompile("/system")}:          routes.System,
		net.RouteConfig{"GET", regexp.MustCompile("/register")}:        (&routes.Registration{registrationStream}).Register,
		net.RouteConfig{"POST", regexp.MustCompile("/device-message")}: routes.CreateDeviceMessage,
	}

	runtime := net.ServerRuntime{
		Logger:             log.New(os.Stdout, "server runtime", log.Ldate|log.Ltime|log.Lshortfile),
		Upgrader:           websocketUpgrader,
		RouteList:          routes,
		BackgroundChannels: backgroundChannels,
	}

	flag.StringVar(&flags.port, "port", "12345", "the port to attach the http listener to")
	flag.Parse()

	if valid := len(flags.port) >= 1; !valid {
		fmt.Printf("invalid port: %s", flags.port)
		flag.PrintDefaults()
		return
	}

	logger.Printf("starting server on port: %s\n", flags.port)

	wg := sync.WaitGroup{}

	for _, processor := range processors {
		wg.Add(1)
		go processor.Start(&wg)
	}

	if e := http.ListenAndServe("0.0.0.0:12345", &runtime); e != nil {
		logger.Fatalf("unable to start server: %s", e.Error())
	}

	wg.Wait()
}
