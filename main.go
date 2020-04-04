package main

import (
	"context"
	"log"
	"os"
	"time"

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
	mongo struct {
		addr     string
		database string
		user     string
		pass     string
	}
	web struct {
		prod bool
		port string
	}
	admin struct {
		pass string
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

	var confContent []byte

	file, err := os.OpenFile("config.toml", os.O_RDONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}

	if _, err = file.Read(confContent); err != nil {
		log.Fatal(err)
	}

	if err := file.Close(); err != nil {
		log.Fatal(err)
	}

	// Config: parse config

	conf := config{}
	if err = toml.Unmarshal(confContent, &conf); err != nil {
		log.Fatal(err)
	}

	// Database: connection initialisation

	mongoOpts := options.Client().ApplyURI(conf.mongo.addr).SetAuth(options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username:      conf.mongo.user,
		Password:      conf.mongo.pass,
	}).SetAppName(appName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, mongoOpts)
	defer cancel()

	if err != nil {
		log.Fatal(err)
	}

	// check alive
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	db := client.Database(conf.mongo.database)

	// OTP initialisation
	
	auth.OTPInitialization(ctx, conf.admin.pass, db)

	// API Cache

	keycache := cache.New(1*time.Hour, 2*time.Hour)

	// Router

	if conf.web.prod {
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
		ctx.Set("otp_crypt", conf.admin.pass)
		ctx.Set("Authenticated", false)
		ctx.Next()
	})

	r.POST("/auth", auth.AuthenticateHandler)
	
	handlers.RegisterPostRoutes(r.Group("/posts"))
	handlers.RegisterProjectRoutes(r.Group("/projects"))
	handlers.RegisterCustomPageRoutes(r.Group("/custom-page"))
	handlers.RegisterHLRoutes(r.Group("/highlights"))
	handlers.RegisterCVRoutes(r.Group("/cv"))

	// bon voyage
	log.Fatal(r.Run(conf.web.port))
}
