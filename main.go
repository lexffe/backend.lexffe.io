package main

import (
	"context"
	"log"
	"time"
	"os"

	"github.com/pelletier/go-toml"
	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type config struct {
	mongo struct {
		addr string
		database string
		user string
		pass string
	}
	web struct {
		port string
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

	if 	_, err = file.Read(confContent); err != nil {
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

	// Database connection initialisation

	mongoOpts := options.Client()
	mongoOpts.ApplyURI(conf.mongo.addr)
	mongoOpts.SetAuth(options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username: conf.mongo.user,
		Password: conf.mongo.pass,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
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

	// Router

	r := gin.Default()

	// Registering database middleware

	r.Use(func(c *gin.Context) {
		c.Set("db", db) // set kv
		c.Next()
	})

	handlers.RegisterPostRoutes(r.Group("/blog"))

	log.Fatal(r.Run(conf.web.port))
}
