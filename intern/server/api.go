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
	"github.com/gocelery/gocelery"
)

var analyses = map[int]*types.Analysis{}
var reviews = map[int]*types.Review{}
var clusters = map[int]*types.Cluster{}
var asyncResults = map[int]*gocelery.AsyncResult{}

var nextAnalysisID = 1
var nextReviewID = 1
var nextClusterID = 1

type WorkerResult struct {
	Status   string          `json:"status"`
	Reviews  []WorkerReview  `json:"reviews"`
	Clusters []WorkerCluster `json:"clusters"`
}

type WorkerReview struct {
	SourceID   string       `json:"source_id"`
	Text       string       `json:"text"`
	Sentiment  string       `json:"sentiment"`
	Confidence float64      `json:"confidence"`
	ClusterID  int          `json:"cluster_id"`
	Coords     WorkerCoords `json:"coords"`
}

type WorkerCoords struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type WorkerCluster struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

type WorkerError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

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

	taskArgID := fmt.Sprintf("analysis-%d", nextAnalysisID)

	asyncResult, err := cli.Delay("worker.process_file", string(csvData), taskArgID)
	if err != nil {
		log.Printf("celery delay error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send task"})
		return
	}

	analysis := &types.Analysis{
		ID:        nextAnalysisID,
		Status:    "pending",
		Filename:  fileHeader.Filename,
		CreatedAt: time.Now(),
		Error:     nil,
		Stats:     nil,
		TaskID:    asyncResult.TaskID,
	}
	analyses[nextAnalysisID] = analysis
	asyncResults[nextAnalysisID] = asyncResult
	nextAnalysisID++

	c.JSON(http.StatusOK, analysis)
}

func listAnalyses(c *gin.Context) {
	for id, a := range analyses {
		if a.Status == "pending" {
			checkAnalysisResult(id, a)
		}
	}

	var data []*types.Analysis
	for _, a := range analyses {
		data = append(data, a)
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func checkAnalysisResult(id int, a *types.Analysis) {
	asyncResult, ok := asyncResults[id]
	if !ok {
		return
	}

	ready, err := asyncResult.Ready()
	if err != nil {
		log.Printf("error checking task ready status: %v", err)
		return
	}

	log.Printf("async result ready: %v", ready)

	if !ready {
		return
	}

	result, err := asyncResult.AsyncGet()
	if err != nil {
		log.Printf("error getting async result: %v", err)
		errMsg := fmt.Sprintf("failed to get result: %v", err)
		a.Status = "failed"
		a.Error = &errMsg
		delete(asyncResults, id)
		return
	}

	raw, err := json.Marshal(result)
	if err != nil {
		log.Printf("marshal worker result error: %v", err)
		errMsg := "invalid worker result format"
		a.Status = "failed"
		a.Error = &errMsg
		delete(asyncResults, id)
		return
	}

	var workerError WorkerError
	if err := json.Unmarshal(raw, &workerError); err == nil && workerError.Status == "error" {
		log.Printf("worker returned error: %s", workerError.Message)
		a.Status = "failed"
		a.Error = &workerError.Message
		delete(asyncResults, id)
		return
	}

	var workerResult WorkerResult
	if err := json.Unmarshal(raw, &workerResult); err != nil {
		log.Printf("unmarshal worker result error: %v", err)
		errMsg := "invalid worker result structure"
		a.Status = "failed"
		a.Error = &errMsg
		delete(asyncResults, id)
		return
	}

	if workerResult.Status == "error" {
		log.Printf("worker result status is error")
		errMsg := "worker processing failed"
		a.Status = "failed"
		a.Error = &errMsg
		delete(asyncResults, id)
		return
	}

	{
		stats := types.Stats{
			Total:    len(workerResult.Reviews),
			Positive: 0,
			Negative: 0,
			Neutral:  0,
		}

		analysisIDStr := strconv.Itoa(a.ID)

		for _, wr := range workerResult.Reviews {
			review := &types.Review{
				ID:         nextReviewID,
				SourceID:   wr.SourceID,
				AnalysisID: analysisIDStr,
				Text:       wr.Text,
				Sentiment:  wr.Sentiment,
				Confidence: wr.Confidence,
				ClusterID:  wr.ClusterID,
				Coords: types.Coord{
					X: wr.Coords.X,
					Y: wr.Coords.Y,
				},
			}
			reviews[nextReviewID] = review
			nextReviewID++

			switch wr.Sentiment {
			case "positive":
				stats.Positive++
			case "negative":
				stats.Negative++
			case "neutral":
				stats.Neutral++
			}
		}

		for _, wc := range workerResult.Clusters {
			cluster := &types.Cluster{
				TrueID:     nextClusterID,
				ID:         wc.ID,
				AnalysisID: a.ID,
				Title:      wc.Title,
				Summary:    wc.Summary,
			}
			clusters[nextClusterID] = cluster
			nextClusterID++
		}

		a.Status = "done"
		a.Stats = &stats
	}

	delete(asyncResults, id)
}

func getAnalysis(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	a, ok := analyses[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	if a.Status == "pending" {
		checkAnalysisResult(id, a)
	}

	c.JSON(http.StatusOK, a)
}

func deleteAnalysis(c *gin.Context) {
	strid := c.Param("id")
	id, _ := strconv.Atoi(strid)

	delete(analyses, id)
	delete(asyncResults, id)

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
	data := []*types.Cluster{}
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
