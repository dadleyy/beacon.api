package main

import "os"
import "io"
import "fmt"
import "log"
import "flag"
import "sync"
import "context"
import "syscall"
import "net/url"
import "net/http"
import "os/signal"

import "crypto/rand"
import "encoding/hex"

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

// TokenGenerator is the used by the redis registry to generate random strings for device tokens.
type TokenGenerator struct {
}

// GenerateToken returns a random hex string.
func (t TokenGenerator) GenerateToken() (string, error) {
	buffer := make([]byte, defs.SecurityUserDeviceTokenSize)

	if _, e := rand.Read(buffer); e != nil {
		return "", e
	}

	return hex.EncodeToString(buffer), nil
}

type wsUpgrader struct {
	websocket.Upgrader
}

func (u *wsUpgrader) UpgradeWebsocket(w http.ResponseWriter, r *http.Request, h http.Header) (defs.Streamer, error) {
	return u.Upgrader.Upgrade(w, r, h)
}

func main() {
	options := struct {
		port       string
		hostname   string
		envFile    string
		redisURI   string
		privateKey string
	}{}

	logger := logging.New(defs.MainLogPrefix, logging.Green)
	flag.StringVar(&options.port, "port", defs.DefaultPort, "the port to attach the http listener to")
	flag.StringVar(&options.hostname, "hostname", defs.DefaultHostname, "the hostname to bind the http.Server to")
	flag.StringVar(&options.envFile, "envfile", ".env", "the environment variable file to load")
	flag.StringVar(&options.redisURI, "redisuri", defs.DefaultRedisURI, "redis server uri")
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

	if os.Getenv("REDIS_URI") != "" {
		options.redisURI = os.Getenv("REDIS_URI")
	}

	if os.Getenv("PORT") != "" {
		options.port = os.Getenv("PORT")
	}

	if os.Getenv("HOSTNAME") != "" {
		options.hostname = os.Getenv("HOSTNAME")
	}

	logger.Debugf("permissions: (admin: %b) (controller %b) (viewer: %b)",
		defs.SecurityDeviceTokenPermissionAdmin,
		defs.SecurityDeviceTokenPermissionController,
		defs.SecurityDeviceTokenPermissionViewer,
	)

	redisURL, e := url.Parse(options.redisURI)

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

	websocket := wsUpgrader{
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     security.AnyOrigin,
		},
	}

	// Create our two device channels - one for holding a connection to the device & one for processing messages from it.
	publisher := bg.ChannelStore{
		defs.DeviceControlChannelName:  make(chan io.Reader, 10),
		defs.DeviceFeedbackChannelName: make(chan io.Reader, 10),
	}

	registrationStream := make(device.RegistrationStream, 10)

	redisPool := redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(options.redisURI)

			if err != nil {
				return nil, err
			}

			password := redisURL.Query().Get("password")

			if password == "" {
				return c, nil
			}

			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}

			return c, nil
		},
	}

	defer redisPool.Close()

	// Create our device store - responsible for providing a persistence layer for connected device information.
	registry := device.RedisRegistry{
		Pool:           &redisPool,
		Logger:         logging.New(defs.RegistryLogPrefix, logging.Green),
		TokenGenerator: TokenGenerator{},
	}

	// Bundle our two message channels w/ the registration stream.
	deviceChannels := bg.DeviceChannels{
		Feedback:      publisher[defs.DeviceFeedbackChannelName],
		Commands:      publisher[defs.DeviceControlChannelName],
		Registrations: registrationStream,
	}

	// Create the main device controller that handles registrations & sending messages to the connected devices.
	control := bg.NewDeviceControlProcessor(&deviceChannels, &registry, serverKey)

	// Create the secondary processor that will receive messages from devices.
	feedback := bg.NewDeviceFeedbackProcessor(publisher[defs.DeviceFeedbackChannelName])

	processors := []bg.Processor{control, feedback}

	deviceRoutes := routes.NewDevicesAPI(&registry, &registry)
	registrationRoutes := routes.NewRegistrationAPI(registrationStream, &registry)
	messageRoutes := routes.NewDeviceMessagesAPI(&registry, &registry)
	feedbackRoutes := routes.NewFeedbackAPI(&registry, &registry)
	tokenRoutes := routes.NewTokensAPI(&registry, &registry)

	routes := net.RouteConfigMapMatcher{
		// [/system]
		net.RouteConfig{
			Method:  "GET",
			Pattern: defs.SystemRoute,
		}: routes.SystemInfo,

		// [/registration]
		net.RouteConfig{
			Method:  "GET",
			Pattern: defs.DeviceRegistrationRoute,
		}: registrationRoutes.Register,
		net.RouteConfig{
			Method:  "POST",
			Pattern: defs.DeviceRegistrationRoute,
		}: registrationRoutes.Preregister,

		// [/device-feedback]
		net.RouteConfig{
			Method:  "POST",
			Pattern: defs.DeviceFeedbackRoute,
		}: feedbackRoutes.CreateFeedback,
		net.RouteConfig{
			Method:  "GET",
			Pattern: defs.DeviceFeedbackRoute,
		}: feedbackRoutes.ListFeedback,

		// [/tokens]
		net.RouteConfig{
			Method:  "POST",
			Pattern: defs.DeviceTokensRoute,
		}: tokenRoutes.CreateToken,
		net.RouteConfig{
			Method:  "GET",
			Pattern: defs.DeviceTokensRoute,
		}: tokenRoutes.ListTokens,

		// [/device-messages]
		net.RouteConfig{
			Method:  "POST",
			Pattern: defs.DeviceMessagesRoute,
		}: messageRoutes.CreateMessage,

		// [/devices/:id/:color]
		net.RouteConfig{
			Method:  "GET",
			Pattern: defs.DeviceShorthandRoute,
		}: deviceRoutes.UpdateShorthand,

		// [/devices]
		net.RouteConfig{
			Method:  "GET",
			Pattern: defs.DeviceListRoute,
		}: deviceRoutes.ListDevices,
	}

	runtime := net.ServerRuntime{
		Logger:             logging.New(defs.ServerRuntimeLogPrefix, logging.Magenta),
		WebsocketUpgrader:  &websocket,
		Multiplexer:        &routes,
		ChannelPublisher:   &publisher,
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
