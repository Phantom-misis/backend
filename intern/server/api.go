package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"backend/intern/types"

	"github.com/gin-gonic/gin"
)

var analyses = map[int]*types.Analysis{}
var reviews = map[int]*types.Review{}
var clusters = map[int]*types.Cluster{}

var nextAnalysisID = 1
var nextReviewID = 1
var nextClusterID = 1

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func login(c *gin.Context) {
	var form types.LoginForm
	if c.ShouldBind(&form) == nil {
		c.JSON(200, gin.H{"message": fmt.Sprintf("Hello %s", form.User)})
	}
}

func createAnalysis(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open file"})
		return
	}
	defer f.Close()

	csvData, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read file"})
		return
	}

	taskID := fmt.Sprintf("analysis-%d", nextAnalysisID)

	asyncResult, err := cli.Delay("worker.process_file", string(csvData), taskID)
	if err != nil {
		log.Printf("celery delay error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send task"})
		return
	}

	result, err := asyncResult.Get(30 * time.Second)
	if err != nil {
		log.Printf("celery get error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "task failed or timed out"})
		return
	}
	var stats types.Stats

	raw, err := json.Marshal(result)
	if err != nil {
		log.Printf("marshal worker result error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid worker result"})
		return
	}

	if err := json.Unmarshal(raw, &stats); err != nil {
		log.Printf("unmarshal worker result error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid worker result"})
		return
	}

	analysis := &types.Analysis{
		ID:        nextAnalysisID,
		Status:    "done",
		Filename:  fileHeader.Filename,
		CreatedAt: time.Now(),
		Error:     nil,
		Stats:     &stats,
	}
	analyses[nextAnalysisID] = analysis
	nextAnalysisID++

	c.JSON(http.StatusOK, analysis)
}

func listAnalyses(c *gin.Context) {
	var data []*types.Analysis
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
	strid := c.Param("id")
	id, _ := strconv.Atoi(strid)

	delete(analyses, id)

	for k, v := range reviews {
		if v.AnalysisID == strid {
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
	analysisID := c.Param("id")
	var data []*types.Review
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
	var data []*types.Cluster
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
