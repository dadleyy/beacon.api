package main

import "os"
import "io"
import "fmt"
import "log"
import "flag"
import "sync"
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
import "github.com/dadleyy/beacon.api/beacon/logging"
import "github.com/dadleyy/beacon.api/beacon/security"
import "github.com/dadleyy/beacon.api/beacon/version"

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
		port       string
		hostname   string
		envFile    string
		redisURL   string
		privateKey string
	}{}

	logger := logging.New(defs.MainLogPrefix, logging.Green)
	flag.StringVar(&options.port, "port", "12345", "the port to attach the http listener to")
	flag.StringVar(&options.hostname, "hostname", "0.0.0.0", "the hostname to bind the http.Server to")
	flag.StringVar(&options.envFile, "envfile", ".env", "the environment variable file to load")
	flag.StringVar(&options.redisURL, "redisuri", "redis://0.0.0.0:6379", "redis server uri")
	flag.StringVar(&options.privateKey, "private-key", ".keys/private.pem", "pem encoded rsa private key")
	flag.Parse()

	if valid := len(options.port) >= 1; !valid {
		logger.Errorf("invalid port: %s", options.port)
		flag.PrintDefaults()
		return
	}

	if e := godotenv.Load(options.envFile); len(options.envFile) > 1 && e != nil {
		logger.Errorf("failed loading env file: %s", e.Error())
		return
	}

	logger.Debugf("permissions: (admin: %b) (controller %b) (viewer: %b)",
		defs.SecurityDeviceTokenPermissionAdmin,
		defs.SecurityDeviceTokenPermissionController,
		defs.SecurityDeviceTokenPermissionViewer,
	)

	redisConnection, e := redis.DialURL(options.redisURL)

	if e != nil {
		logger.Errorf("unable to establish connection to redis server: %s", e.Error())
		return
	}

	serverKey, e := security.ReadServerKeyFromFile(options.privateKey)

	if e != nil {
		logger.Errorf("unable to load server key from file[%s]: %s", options.privateKey, e.Error())
		return
	}

	if s, e := serverKey.SharedSecret(); e == nil {
		logger.Debugf("server key loaded, shared secret: \n%x\n\n", s)
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

	registrationStream := make(device.RegistrationStream, 10)

	registry := device.RedisRegistry{
		Conn:   redisConnection,
		Logger: logging.New(defs.RegistryLogPrefix, logging.Green),
	}

	deviceChannels := bg.DeviceChannels{
		Feedback:      backgroundChannels[defs.DeviceFeedbackChannelName],
		Commands:      backgroundChannels[defs.DeviceControlChannelName],
		Registrations: registrationStream,
	}

	control := bg.NewDeviceControlProcessor(&deviceChannels, &registry, serverKey)
	feedback := bg.NewDeviceFeedbackProcessor(backgroundChannels[defs.DeviceFeedbackChannelName])

	processors := []bg.Processor{control, feedback}

	deviceRoutes := routes.NewDevicesAPI(&registry, &registry)
	registrationRoutes := routes.NewRegistrationAPI(registrationStream, &registry)
	messageRoutes := routes.NewDeviceMessagesAPI(&registry, &registry)
	feedbackRoutes := routes.NewFeedbackAPI(&registry)
	tokenRoutes := routes.NewTokensAPI(&registry, &registry)

	routes := net.RouteList{
		net.RouteConfig{"GET", defs.SystemRoute}: routes.System,

		net.RouteConfig{"GET", defs.DeviceRegistrationRoute}:  registrationRoutes.Register,
		net.RouteConfig{"POST", defs.DeviceRegistrationRoute}: registrationRoutes.Preregister,

		net.RouteConfig{"POST", defs.DeviceFeedbackRoute}: feedbackRoutes.CreateFeedback,
		net.RouteConfig{"GET", defs.DeviceFeedbackRoute}:  feedbackRoutes.ListFeedback,

		net.RouteConfig{"POST", defs.DeviceTokensRoute}: tokenRoutes.CreateToken,

		net.RouteConfig{"POST", defs.DeviceMessagesRoute}: messageRoutes.CreateMessage,

		net.RouteConfig{"GET", defs.DeviceShorthandRoute}: deviceRoutes.UpdateShorthand,
		net.RouteConfig{"GET", defs.DeviceListRoute}:      deviceRoutes.ListDevices,
	}

	runtime := net.ServerRuntime{
		Logger:             logging.New(defs.ServerRuntimeLogPrefix, logging.Magenta),
		Upgrader:           websocketUpgrader,
		RouteList:          routes,
		BackgroundChannels: backgroundChannels,
		RedisConnection:    redisConnection,
		ApplicationVersion: version.Semver,
	}

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

	logger.Infof("server (version %s) starting, binding on: %s\n", version.Semver, serverAddress)

	if e := server.ListenAndServe(); e != nil {
		logger.Debugf("server shutdown: %s", e.Error())
	}

	wg.Wait()
	os.Exit(1)
}
