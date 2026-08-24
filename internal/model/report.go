package model

type BatchSummary struct {
	Batch       NightBatch `json:"batch"`
	Specimens   int        `json:"specimens"`
	Classified  int        `json:"classified"`
	Reviews     int        `json:"reviews"`
	Blockers    int        `json:"blockers"`
	Quality     float64    `json:"quality"`
	LastReading *Reading   `json:"last_reading,omitempty"`
}

type DailyRecord struct {
	Day       string  `json:"day"`
	Batches   int     `json:"batches"`
	Specimens int     `json:"specimens"`
	Captures  int     `json:"captures"`
	Quality   float64 `json:"quality"`
}

type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func (p Page) Normalize() Page {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 25
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p
}

type Health struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	QueueSize int    `json:"queue_size"`
	Now       string `json:"now"`
}
