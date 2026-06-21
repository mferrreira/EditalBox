package domain

import "time"

type Notice struct {
	ID              int64
	SourceURL       string
	CanonicalURL    string
	Title           string
	Excerpt         string
	BodyText        string
	SourceType      string
	Status          string
	RegistrationEnd *time.Time
	FinalEventAt    *time.Time
	UpdatedAt       time.Time
	CreatedAt       time.Time
}

type NoticeDocument struct {
	ID        int64
	NoticeID  int64
	URL       string
	Kind      string
	Title     string
	UpdatedAt time.Time
}

type NoticeEvent struct {
	ID        int64
	NoticeID  int64
	EventType string
	Label     string
	EventAt   time.Time
	UpdatedAt time.Time
}

type SyncRun struct {
	ID              int64
	StartedAt       time.Time
	FinishedAt      *time.Time
	Status          string
	SeedsVisited    int
	PagesFetched    int
	NoticesUpserted int
	ErrorMessage    string
}

type SessionSummary struct {
	ChatID       int64
	Summary      string
	LastActivity time.Time
}

type CandidateAnswer struct {
	ID       int64    `json:"id"`
	Title    string   `json:"title"`
	URL      string   `json:"url"`
	Status   string   `json:"status"`
	Excerpt  string   `json:"excerpt"`
	BodyText string   `json:"body_text"`
	Keywords []string `json:"keywords"`
}
