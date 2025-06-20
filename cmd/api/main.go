package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort = "80"
	// remote procedure call
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	gRpcPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// connect to Mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}
	client = mongoClient

	// create a context that auto expires after 15s
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	// do both below bottom up after finishing above ops
	// frees timer
	defer cancel()

	// close connections (runs before cancel())
	defer func() {
		// gives mongoDB 15s to end connection, else throws error
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	// register RPC server
	err = rpc.Register(new(RPCServer))
	go app.rpcListen()

	log.Println("Starting service on port", webPort)
	srv := &http.Server{
		// alternatively, Addr: ":8080"
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		// stops execution
		log.Panic(err)
	}

	// app.serve()
}

// replacing serve()
// sets up RPC server
func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on port ", rpcPort)
	// start an RPC connection based on webPort
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(rpcConn)
	}
}

// func (app *Config) serve() {
// 	srv := &http.Server{
// 		// alternatively, Addr: ":8080"
// 		Addr:    fmt.Sprintf(":%s", webPort),
// 		Handler: app.routes(),
// 	}

// 	err := srv.ListenAndServe()
// 	if err != nil {
// 		// stops execution
// 		log.Panic(err)
// 	}
// }

func connectToMongo() (*mongo.Client, error) {
	// create connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error connecting:", err)
		return nil, err
	}

	log.Println("Connected to mongo!")

	return c, nil
}
