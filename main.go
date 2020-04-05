package main

import (
	"context"
	"io/ioutil"
	"log"
	"time"

	// "os"
	// "flag"

	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/auth"
	"github.com/lexffe/backend.lexffe.io/handlers"
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
	Admin struct {
		Pass string
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

	// OTP initialisation

	if err := auth.OTPInitialization(ctx, conf.Admin.Pass, db); err != nil {
		log.Fatal(err)
	}

	// API Cache

	keycache := cache.New(1*time.Hour, 2*time.Hour)

	// Router

	if conf.Web.Prod {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	// Registering database middleware

	r.Use(func(ctx *gin.Context) {
		ctx.Set("db", db) // set kv
		ctx.Next()
	})

	// registering cache

	r.Use(func(ctx *gin.Context) {
		ctx.Set("keycache", keycache)
		ctx.Next()
	})

	// registering auth variables

	r.Use(func(ctx *gin.Context) {
		ctx.Set("otp_crypt", conf.Admin.Pass)
		ctx.Set("Authenticated", false)
		ctx.Next()
	})

	r.POST("/auth", auth.AuthenticateHandler)
	r.Use(auth.BearerMiddleware)

	handlers.RegisterPostRoutes(r.Group("/posts"))
	handlers.RegisterProjectRoutes(r.Group("/projects"))
	handlers.RegisterCustomPageRoutes(r.Group("/custom-page"))
	handlers.RegisterHLRoutes(r.Group("/highlights"))
	handlers.RegisterCVRoutes(r.Group("/cv"))

	// bon voyage
	log.Fatal(r.Run(conf.Web.Port))
}
