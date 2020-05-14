package main

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/auth"
	"github.com/lexffe/backend.lexffe.io/coll"
	"github.com/patrickmn/go-cache"
	"github.com/pelletier/go-toml"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type appConfig struct {
	Meta struct {
		AppName  string `toml:"appname"`
		CorsHost string `toml:"cors_host"`
	}
	Mongo struct {
		Addr     string
		Database string
		Auth     bool
		User     string
		Pass     string
	}
	Web struct {
		TCP      bool
		UnixPath string `toml:"unixpath"`
		Prod     bool
		Port     string
	}
}

/**
What is here?

Database connection initialisation
Database instance propagation
Router initialisation
Route registration
*/

//noinspection GoNilness
func main() {

	// Config: read file

	confContent, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Println("Cannot read configuration file.")
		log.Fatal(err)
	}

	// Config: parse config

	var conf appConfig
	if err = toml.Unmarshal(confContent, &conf); err != nil {
		log.Println("Cannot unmarshal configuration file.")
		log.Fatal(err)
	}

	// Database: connection initialisation

	mongoOpts := options.Client().ApplyURI(conf.Mongo.Addr).SetAppName(conf.Meta.AppName)

	if conf.Mongo.Auth == true {
		mongoOpts.SetAuth(options.Credential{
			AuthMechanism: "SCRAM-SHA-256",
			Username:      conf.Mongo.User,
			Password:      conf.Mongo.Pass,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, mongoOpts)
	defer client.Disconnect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	// Database: check alive
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	db := client.Database(conf.Mongo.Database)

	// Auth: API Key cache

	keycache := cache.New(1*time.Hour, 2*time.Hour)

	// Webserver: Router

	if conf.Web.Prod {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	// Webserver: CORS

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = append([]string{}, conf.Meta.CorsHost)

	r.Use(cors.New(corsConfig))

	// Webserver: registering authentication routes

	authHandler := auth.AuthenticateHandler{
		Issuer: conf.Meta.AppName,
		Cache:  keycache,
	}

	// Webserver: initialize otp
	if err := authHandler.OTPInitialization(); err != nil {
		log.Fatal(err)
	}

	r.POST("/auth", authHandler.Handler)
	r.Use(authHandler.BearerMiddleware)

	// Webserver: Bootstrap Existing collections in database

	bootstrapper := coll.CollectionDelegate{
		Engine: r,
		DB:     db,
	}

	bootstrapper.RegisterRoutes()

	if err := bootstrapper.Bootstrap(ctx); err != nil {
		log.Fatal(err)
	}

	r.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Alive")
	})

	// Webserver: http server

	srv := &http.Server{
		Addr:         conf.Web.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// HTTP Server: peaceful shutdown routine

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// TCP Server

	if conf.Web.TCP == true {
		// new goroutine to serve the http server

		tcpListener, err := net.Listen("tcp", conf.Web.Port)

		if err != nil {
			log.Println("Cannot listen on tcp socket")
			log.Fatal(err)
		}

		defer tcpListener.Close()

		// TCP Server: Serve on different goroutine for non-blocking

		go func(l *net.Listener) {
			log.Fatal(srv.Serve(*l))
		}(&tcpListener)
	}

	// UNIX Socket

	unixListener, err := net.ListenUnix("unix", &net.UnixAddr{
		Name: conf.Web.UnixPath,
		Net:  "unix",
	})

	// special routine for cleaning up unix socket

	if err != nil {
		log.Println("Cannot listen on unix socket")
		log.Fatal(err)
	}

	unixListener.SetUnlinkOnClose(true)

	// UNIX Socket: Goroutine for checking system signals

	go func() {
		for sig := range c {
			log.Printf("signal detected: %v, cleaning up unix.", sig)
			if err := unixListener.Close(); err != nil {
				log.Println("unix dirty close")
				os.Exit(1)
			}
			os.Exit(0)
		}
	}()

	// UNIX Socket: bon voyage
	log.Fatal(srv.Serve(unixListener))
}
