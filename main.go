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
	"github.com/lexffe/backend.lexffe.io/handlers"
	"github.com/lexffe/backend.lexffe.io/models"
	"github.com/patrickmn/go-cache"
	"github.com/pelletier/go-toml"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type appConfig struct {
	Meta struct {
		AppName string `toml:"appname"`
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
	client, err := mongo.Connect(ctx, mongoOpts)
	defer cancel()

	if err != nil {
		log.Fatal(err)
	}

	// check alive
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	db := client.Database(conf.Mongo.Database)

	// API Key cache

	keycache := cache.New(1*time.Hour, 2*time.Hour)

	// Router

	if conf.Web.Prod {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	// CORS

	r.Use(cors.Default())

	// registering authentication routes

	authHandler := auth.AuthenticateHandler{
		DB:         db,
		Cache:      keycache,
		Collection: "auth",
	}

	// initialize otp
	if err := authHandler.OTPInitialization(ctx); err != nil {
		log.Fatal(err)
	}

	r.POST("/auth", authHandler.Handler)
	r.Use(authHandler.BearerMiddleware)

	Posts := handlers.PageHandler{
		Router:     r.Group("/posts"),
		DB:         db,
		PageType:   models.TypePostPage,
		Collection: string(models.TypePostPage),
	}

	Posts.RegisterRoutes()

	// Note: the view should render custom pages in a nav.
	Pages := handlers.PageHandler{
		Router:     r.Group("/pages"),
		DB:         db,
		PageType:   models.TypeGenericPage,
		Collection: string(models.TypeGenericPage),
	}

	Pages.RegisterRoutes()

	// TODO: set capped collection
	CV := handlers.PageHandler{
		Router:     r.Group("/cv"),
		DB:         db,
		PageType:   models.TypeCVPage,
		Collection: string(models.TypeCVPage),
	}

	CV.RegisterRoutes()

	Projects := handlers.ReferenceHandler{
		Router:        r.Group("/projects"),
		DB:            db,
		ReferenceType: models.TypeProjectRef,
		Collection:    string(models.TypeProjectRef),
	}

	Projects.RegisterRoutes()

	Highlights := handlers.ReferenceHandler{
		Router:        r.Group("/highlights"),
		DB:            db,
		ReferenceType: models.TypeHighlightRef,
		Collection:    string(models.TypeHighlightRef),
	}

	Highlights.RegisterRoutes()

	r.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Alive")
	})

	// http server

	srv := &http.Server{
		Addr:         conf.Web.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// unix?

	if conf.Web.TCP == true {
		// new goroutine to serve the http server

		tcpListener, err := net.Listen("tcp", conf.Web.Port)

		if err != nil {
			log.Println("Cannot listen on tcp socket")
			log.Fatal(err)
		}

		defer tcpListener.Close()

		go func(l *net.Listener) {
			log.Fatal(srv.Serve(*l))
		}(&tcpListener)
	}

	unixListener, err := net.ListenUnix("unix", &net.UnixAddr{
		Name: conf.Web.UnixPath,
		Net: "unix",
	})

	// special routine for cleaning up unix socket

	unixListener.SetUnlinkOnClose(true)

	if err != nil {
		log.Println("Cannot listen on unix socket")
		log.Fatal(err)
	}

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

	// bon voyage
	log.Fatal(srv.Serve(unixListener))
}
