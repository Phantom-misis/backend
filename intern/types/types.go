package types

import "time"

type LoginForm struct {
	User     string `form:"user" binding:"required"`
	Password string `form:"password" binding:"required"`
}

type Analysis struct {
	ID        int       `json:"id"`
	Status    string    `json:"status"`
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"created_at"`
	Error     *string   `json:"error"`
	Stats     *Stats    `json:"stats"`
	TaskID    string    `json:"-"`
}

type Stats struct {
	Total    int `json:"total"`
	Positive int `json:"positive"`
	Negative int `json:"negative"`
	Neutral  int `json:"neutral"`
}

type Review struct {
	ID         int     `json:"id"`
	AnalysisID string  `json:"analysis_id"`
	SourceID   string  `json:"source_id"`
	Text       string  `json:"text"`
	Sentiment  string  `json:"sentiment"`
	Confidence float64 `json:"confidence"`
	ClusterID  int     `json:"cluster_id"`
	Coords     Coord   `json:"coords"`
}

type Coord struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Cluster struct {
	TrueID     int    `json:"-"`
	ID         int    `json:"id"`
	AnalysisID int    `json:"analysis_id"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
}
