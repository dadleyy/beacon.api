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

import "github.com/joho/godotenv"
import "github.com/gorilla/websocket"
import "github.com/garyburd/redigo/redis"

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
	options := struct {
		port     string
		hostname string
		envFile  string
		redisURL string
	}{}

	logger := log.New(os.Stdout, "beacon ", log.Ldate|log.Ltime|log.Lshortfile)
	flag.StringVar(&options.port, "port", "12345", "the port to attach the http listener to")
	flag.StringVar(&options.hostname, "hostname", "0.0.0.0", "the hostname to bind the http.Server to")
	flag.StringVar(&options.envFile, "envfile", ".env", "the environment variable file to load")
	flag.StringVar(&options.redisURL, "redisuri", "redis://0.0.0.0:6379", "redis server uri")
	flag.Parse()

	if valid := len(options.port) >= 1; !valid {
		logger.Printf("invalid port: %s", options.port)
		flag.PrintDefaults()
		return
	}

	if e := godotenv.Load(options.envFile); len(options.envFile) > 1 && e != nil {
		logger.Printf("failed loading env file: %s", e.Error())
		return
	}

	redisConnection, e := redis.DialURL(options.redisURL)

	if e != nil {
		logger.Printf("unable to establish connection to redis server: %s", e.Error())
		return
	}

	defer redisConnection.Close()

	websocketUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     security.AnyOrigin,
	}

	backgroundChannels := defs.BackgroundChannels{
		defs.DeviceControlChannelName:  make(chan io.Reader, 10),
		defs.DeviceFeedbackChannelName: make(chan io.Reader, 10),
	}

	registrationStream := make(chan *device.Connection, 10)

	registry := device.RedisRegistry{
		Conn:   redisConnection,
		Logger: log.New(os.Stdout, "device registry ", log.Ldate|log.Ltime|log.Lshortfile),
	}

	deviceChannels := bg.DeviceChannels{
		Feedback:      backgroundChannels[defs.DeviceFeedbackChannelName],
		Commands:      backgroundChannels[defs.DeviceControlChannelName],
		Registrations: registrationStream,
	}

	control := bg.NewDeviceControlProcessor(&deviceChannels, &registry)

	feedback := bg.DeviceFeedbackProcessor{
		Logger:    log.New(os.Stdout, "device control ", log.Ldate|log.Ltime|log.Lshortfile),
		LogStream: backgroundChannels[defs.DeviceFeedbackChannelName],
	}

	processors := []bg.Processor{control, &feedback}

	deviceRoutes := &routes.Devices{&registry}
	registrationRoutes := &routes.Registration{registrationStream}

	routes := net.RouteList{
		net.RouteConfig{"GET", regexp.MustCompile("^/system")}:                routes.System,
		net.RouteConfig{"GET", regexp.MustCompile("^/register")}:              registrationRoutes.Register,
		net.RouteConfig{"POST", regexp.MustCompile("^/device-message")}:       routes.CreateDeviceMessage,
		net.RouteConfig{"GET", regexp.MustCompile(defs.DeviceShorthandRoute)}: deviceRoutes.UpdateShorthand,
	}

	runtime := net.ServerRuntime{
		Logger:             log.New(os.Stdout, "server runtime", log.Ldate|log.Ltime|log.Lshortfile),
		Upgrader:           websocketUpgrader,
		RouteList:          routes,
		BackgroundChannels: backgroundChannels,
		RedisConnection:    redisConnection,
	}

	logger.Printf("starting server on port: %s\n", options.port)

	wg, signalChan, killers := sync.WaitGroup{}, make(chan os.Signal, 1), make([]bg.KillSwitch, 0)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	for _, processor := range processors {
		wg.Add(1)
		stop := make(bg.KillSwitch)
		killers = append(killers, stop)
		go processor.Start(&wg, stop)
	}

	serverAddress := fmt.Sprintf("%s:%s", options.hostname, options.port)
	server := http.Server{Addr: serverAddress, Handler: &runtime}

	go systemWatch(signalChan, killers, &server)

	if e := server.ListenAndServe(); e != nil {
		logger.Printf("server shutdown: %s", e.Error())
	}

	wg.Wait()
	os.Exit(1)
}
