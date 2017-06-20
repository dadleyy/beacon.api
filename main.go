package main

import "os"
import "io"
import "fmt"
import "log"
import "flag"
import "sync"
import "regexp"
import "context"
import "syscall"
import "net/http"
import "os/signal"
import "github.com/gorilla/websocket"

import "github.com/dadleyy/beacon.api/beacon/bg"
import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/routes"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/security"

func systemWatch(system chan os.Signal, killers []bg.KillSwitch, server *http.Server) {
	<-system
	log.Printf("receiving system exit signal, killing background processors")

	for _, switcher := range killers {
		switcher <- struct{}{}
	}

	server.Shutdown(context.Background())
}

func main() {
	flags := struct {
		port     string
		hostname string
	}{}

	logger := log.New(os.Stdout, "beacon ", log.Ldate|log.Ltime|log.Lshortfile)

	flag.StringVar(&flags.port, "port", "12345", "the port to attach the http listener to")
	flag.StringVar(&flags.hostname, "hostname", "0.0.0.0", "the hostname to bind the http.Server to")
	flag.Parse()

	if valid := len(flags.port) >= 1; !valid {
		fmt.Printf("invalid port: %s", flags.port)
		flag.PrintDefaults()
		return
	}

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
		net.RouteConfig{"GET", regexp.MustCompile("^/system")}: routes.System,
		net.RouteConfig{
			Method:  "GET",
			Pattern: regexp.MustCompile("^/register"),
		}: (&routes.Registration{registrationStream}).Register,
		net.RouteConfig{
			Method:  "POST",
			Pattern: regexp.MustCompile("^/device-message"),
		}: routes.CreateDeviceMessage,
		net.RouteConfig{
			Method:  "GET",
			Pattern: regexp.MustCompile("^/devices/(?P<uuid>[\\d\\w\\-]+)/(?P<color>red|blue|green)$"),
		}: routes.UpdateDeviceShorthand,
	}

	runtime := net.ServerRuntime{
		Logger:             log.New(os.Stdout, "server runtime", log.Ldate|log.Ltime|log.Lshortfile),
		Upgrader:           websocketUpgrader,
		RouteList:          routes,
		BackgroundChannels: backgroundChannels,
	}

	logger.Printf("starting server on port: %s\n", flags.port)

	wg, signalChan, killers := sync.WaitGroup{}, make(chan os.Signal, 1), make([]bg.KillSwitch, 0)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	for _, processor := range processors {
		wg.Add(1)
		stop := make(bg.KillSwitch)
		killers = append(killers, stop)
		go processor.Start(&wg, stop)
	}

	serverAddress := fmt.Sprintf("%s:%s", flags.hostname, flags.port)
	server := http.Server{Addr: serverAddress, Handler: &runtime}

	go systemWatch(signalChan, killers, &server)

	if e := server.ListenAndServe(); e != nil {
		logger.Printf("server shutdown: %s", e.Error())
	}

	wg.Wait()
	os.Exit(1)
}
