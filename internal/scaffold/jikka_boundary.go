package scaffold

import "time"

type JikkaBoundary struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`
	Target    string    `json:"target"`
	Mutual    bool      `json:"mutual"`
	Strength  float64   `json:"strength"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
