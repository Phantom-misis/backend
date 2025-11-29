package server

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gocelery/gocelery"
	redigo "github.com/gomodule/redigo/redis"
)

var cli *gocelery.CeleryClient
var redisPool *redigo.Pool

func Start() {
	initCelery()
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Analyses
	r.POST("/analyses", createAnalysis)
	r.GET("/analyses", listAnalyses)
	r.GET("/analyses/:id", getAnalysis)
	r.DELETE("/analyses/:id", deleteAnalysis)

	// Reviews
	r.GET("/analyses/:id/reviews", listReviews)
	r.GET("/reviews/:id", getReview)
	r.PATCH("/reviews/:id", updateReview)

	// Clusters
	r.GET("/analyses/:id/clusters", listClusters)
	r.GET("/clusters/:id", getCluster)

	r.Run(":8080")
}

func initCelery() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisAddr := redisHost + ":" + redisPort

	if redisPassword != "" {
		log.Printf("Connecting to Redis at %s (with password)", redisAddr)
	} else {
		log.Printf("Connecting to Redis at %s (no password)", redisAddr)
	}

	redisPool = &redigo.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigo.Conn, error) {
			conn, err := redigo.Dial("tcp", redisAddr)
			if err != nil {
				return nil, err
			}

			if redisPassword != "" {
				_, err = conn.Do("AUTH", redisPassword)
				if err != nil {
					conn.Close()
					return nil, err
				}
			}

			return conn, nil
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	var err error
	cli, err = gocelery.NewCeleryClient(
		gocelery.NewRedisBroker(redisPool),
		gocelery.NewRedisBackend(redisPool),
		0,
	)
	if err != nil {
		log.Fatalf("failed to init celery client: %v", err)
	}
	log.Printf("Celery client initialized successfully")
}
