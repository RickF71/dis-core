package auth

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateChallengeResponse is the response for POST /api/auth/challenge
type CreateChallengeResponse struct {
	ChallengeID string `json:"challenge_id"`
	QRPayload   string `json:"qr_payload"`
}

// ChallengeStatusResponse is the response for GET /api/auth/challenge/:id/status
type ChallengeStatusResponse struct {
	Status string `json:"status"`
}

// CompleteChallengeRequest is the request for POST /api/auth/qr-complete
type CompleteChallengeRequest struct {
	ChallengeID string `json:"challenge_id"`
	UserID      string `json:"user_id"`
}

// CompleteChallengeResponse is the response for POST /api/auth/qr-complete
type CompleteChallengeResponse struct {
	OK bool `json:"ok"`
}

// HandleCreateChallenge creates a new authentication challenge
// Endpoint: POST /api/auth/challenge
//
// Purpose: Browser calls this from NoneSpace to create a new auth challenge.
// DIS-Core ties the challenge to the current browser session.
func HandleCreateChallenge(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Legacy QR challenge endpoint is no longer supported", http.StatusGone)
	}
}

// HandleChallengeStatus returns the current status of a challenge
// Endpoint: GET /api/auth/challenge/:id/status
//
// Purpose: Browser polls to see if the challenge is authenticated.
func HandleChallengeStatus(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Legacy QR challenge status endpoint is no longer supported", http.StatusGone)
	}
}

// HandleCompleteChallenge marks a challenge as authenticated
// Endpoint: POST /api/auth/qr-complete
//
// Purpose: Phone (or external agent) calls this after scanning QR and authenticating user.
func HandleCompleteChallenge(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Legacy QR challenge complete endpoint is no longer supported", http.StatusGone)
	}
}
