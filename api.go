package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func login(c *gin.Context) {
	var form LoginForm
	if c.ShouldBind(&form) == nil {
		c.JSON(200, gin.H{"message": fmt.Sprintf("Hello %s", form.User)})
	}
}

func createAnalysis(c *gin.Context) {
	file, _ := c.FormFile("file")
	id := nextAnalysisID
	nextAnalysisID++

	analysis := &Analysis{
		ID:        id,
		Status:    "pending",
		Filename:  file.Filename,
		CreatedAt: time.Now(),
		Error:     nil,
		Stats:     nil,
	}
	analyses[id] = analysis

	c.JSON(http.StatusOK, analysis)
}

func listAnalyses(c *gin.Context) {
	var data []*Analysis
	for _, a := range analyses {
		data = append(data, a)
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func getAnalysis(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if a, ok := analyses[id]; ok {
		c.JSON(http.StatusOK, a)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

func deleteAnalysis(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	delete(analyses, id)

	for k, v := range reviews {
		if v.AnalysisID == id {
			delete(reviews, k)
		}
	}

	for k, v := range clusters {
		if v.AnalysisID == id {
			delete(clusters, k)
		}
	}

	c.Status(http.StatusNoContent)
}

func listReviews(c *gin.Context) {
	analysisID, _ := strconv.Atoi(c.Param("id"))
	var data []*Review
	for _, r := range reviews {
		if r.AnalysisID == analysisID {
			data = append(data, r)
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func getReview(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if r, ok := reviews[id]; ok {
		c.JSON(http.StatusOK, r)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

func updateReview(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if r, ok := reviews[id]; ok {
		var payload struct {
			Sentiment *string `json:"sentiment"`
		}
		c.BindJSON(&payload)

		if payload.Sentiment != nil {
			r.Sentiment = *payload.Sentiment
			r.Confidence = 1
		}

		c.JSON(http.StatusOK, r)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

func listClusters(c *gin.Context) {
	analysisID, _ := strconv.Atoi(c.Param("id"))
	var data []*Cluster
	for _, cl := range clusters {
		if cl.AnalysisID == analysisID {
			data = append(data, cl)
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func getCluster(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if cl, ok := clusters[id]; ok {
		c.JSON(http.StatusOK, cl)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}
