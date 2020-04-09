package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	// "os"
	// "flag"

	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/auth"
	"github.com/lexffe/backend.lexffe.io/handlers"
	"github.com/lexffe/backend.lexffe.io/models"
	"github.com/patrickmn/go-cache"
	"github.com/pelletier/go-toml"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const appName = "backend"

type config struct {
	Mongo struct {
		Addr     string
		Database string
		User     string
		Pass     string
	}
	Web struct {
		Prod bool
		Port string
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
		log.Fatal(err)
	}

	// Config: parse config

	var conf config
	if err = toml.Unmarshal(confContent, &conf); err != nil {
		log.Fatal(err)
	}
	// Database: connection initialisation

	mongoOpts := options.Client().ApplyURI(conf.Mongo.Addr).SetAuth(options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username:      conf.Mongo.User,
		Password:      conf.Mongo.Pass,
	}).SetAppName(appName)

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

	// registering authentication routes

	authHandler := auth.AuthenticateHandler{
		DB: db,
		Cache: keycache,
		Collection: "auth",
	}

	// initialize otp
	if err := authHandler.OTPInitialization(ctx); err != nil {
		log.Fatal(err)
	}

	r.POST("/auth", authHandler.Handler)
	r.Use(authHandler.BearerMiddleware)

	Posts := handlers.PageHandler{
		Router: r.Group("/posts"),
		DB: db,
		PageType: models.TypePostPage,
		Collection: string(models.TypePostPage),
	}

	Posts.RegisterRoutes()

	// Note: the view should render custom pages in a nav.
	Pages := handlers.PageHandler{
		Router: r.Group("/pages"),
		DB: db,
		PageType: models.TypeGenericPage,
		Collection: string(models.TypeGenericPage),
	}

	Pages.RegisterRoutes()

	// TODO: set capped collection
	CV := handlers.PageHandler{
		Router: r.Group("/cv"),
		DB: db,
		PageType: models.TypeCVPage,
		Collection: string(models.TypeCVPage),
	}

	CV.RegisterRoutes()

	Projects := handlers.ReferenceHandler{
		Router: r.Group("/projects"),
		DB: db,
		ReferenceType: models.TypeProjectRef,
		Collection: string(models.TypeProjectRef),
	}

	Projects.RegisterRoutes()

	Highlights := handlers.ReferenceHandler{
		Router: r.Group("/highlights"),
		DB: db,
		ReferenceType: models.TypeHighlightRef,
		Collection: string(models.TypeHighlightRef),
	}

	Highlights.RegisterRoutes()

	// http server

	srv := &http.Server{
		Addr: conf.Web.Port,
		Handler: r,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// bon voyage
	log.Fatal(srv.ListenAndServe())
}
