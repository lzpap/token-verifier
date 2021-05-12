package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/capossele/asset-registry/pkg/registryservice"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

const (
	// defaultMongoDBOpTimeout defines the default MongoDB operation timeout.
	defaultMongoDBOpTimeout = 5 * time.Second
)

var (
	clientDB *mongo.Client
	// db       *mongo.Database
	dbOnce sync.Once
	// read locked by pingers and write locked by the routine trying to reconnect.
	mongoReconnectLock sync.RWMutex
	// server is the web API server.
	server     *echo.Echo
	serverOnce sync.Once

	log *zap.SugaredLogger

	bindAddress = "0.0.0.0:80"
)

var (
	mongodbUsername = flag.String("username", "root", "mongoDB username")
	mongoDBpassword = flag.String("password", "password", "mongoDB password")
	mongoDBHostAddr = flag.String("hostAddr", "mongodb:27017", "mongoDB host address")
)

func main() {
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	log = logger.Sugar()

	service := registryservice.NewService(mongoDB())
	httpHandler := registryservice.NewHTTPHandler(service, log)

	Server()

	// configure the server
	server.HideBanner = true
	server.HidePort = true

	server.GET("/", IndexRequest)
	server.POST("/registries/:network/assets", httpHandler.SaveAssets)
	server.GET("/registries/:network/assets", httpHandler.LoadAssets)
	server.GET("/registries/:network/assets/:ID", httpHandler.LoadAsset)

	log.Infof("Starting server ...")

	log.Fatal(server.Start(bindAddress))
}

// IndexRequest returns INDEX
func IndexRequest(c echo.Context) error {
	return c.String(http.StatusOK, "INDEX")
}

// Server gets the server instance.
func Server() *echo.Echo {
	serverOnce.Do(func() {
		server = echo.New()
		server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			Skipper:      middleware.DefaultSkipper,
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		}))

		server.HTTPErrorHandler = func(err error, c echo.Context) {
			log.Warnf("Request failed: %s", err)

			var statusCode int
			var message string

			switch errors.Unwrap(err) {
			case echo.ErrUnauthorized:
				statusCode = http.StatusUnauthorized
				message = "unauthorized"

			case echo.ErrForbidden:
				statusCode = http.StatusForbidden
				message = "access forbidden"

			case echo.ErrInternalServerError:
				statusCode = http.StatusInternalServerError
				message = "internal server error"

			case echo.ErrNotFound:
				statusCode = http.StatusNotFound
				message = "not found"

			case echo.ErrBadRequest:
				statusCode = http.StatusBadRequest
				message = "bad request"

			default:
				statusCode = http.StatusInternalServerError
				message = "internal server error"
			}

			message = fmt.Sprintf("%s, error: %+v", message, err)
			c.String(statusCode, message)
		}
	})
	return server
}

func mongoDB() *mongo.Database {
	dbOnce.Do(func() {
		client, err := newMongoDB()
		if err != nil {
			log.Fatal(err)
		}
		clientDB = client
	})
	return clientDB.Database("registry")
}

func newMongoDB() (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://" + *mongodbUsername + ":" + *mongoDBpassword + "@" + *mongoDBHostAddr))
	if err != nil {
		log.Fatalf("MongoDB NewClient failed: %s", err)
		return nil, err
	}

	if err := connectMongoDB(client); err != nil {
		return nil, err
	}

	if err := pingMongoDB(client); err != nil {
		return nil, err
	}

	return client, nil
}

func connectMongoDB(client *mongo.Client) error {
	ctx, cancel := operationTimeout(defaultMongoDBOpTimeout)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		log.Warnf("MongoDB connection failed: %s", err)
		return err
	}
	return nil
}

func pingMongoDB(client *mongo.Client) error {
	ctx, cancel := operationTimeout(defaultMongoDBOpTimeout)
	defer cancel()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Warnf("MongoDB ping failed: %s", err)
		return err
	}
	return nil
}

func operationTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
