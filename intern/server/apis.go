package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Start() {
	r := gin.Default()
	r.Use(cors.Default())
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
