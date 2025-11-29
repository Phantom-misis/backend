package server

import (
	"log"
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
	redisPool = &redigo.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigo.Conn, error) {

			return redigo.Dial("tcp", "localhost:13394")
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
}
