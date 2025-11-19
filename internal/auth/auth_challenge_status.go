package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ChallengeStatusHandler struct {
	Store ChallengeStore
}

type challengeStatusResponse struct {
	ChallengeID string `json:"challenge_id"`
	Status      string `json:"status"`
}

func NewChallengeStatusHandler(store ChallengeStore) http.Handler {
	return &ChallengeStatusHandler{Store: store}
}

func (h *ChallengeStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "missing challenge ID", http.StatusBadRequest)
		return
	}

	challengeID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid challenge ID", http.StatusBadRequest)
		return
	}

	challenge, err := h.Store.GetChallenge(r.Context(), challengeID)
	if err != nil || challenge == nil {
		// Intentionally generic to avoid leaking existence details.
		http.Error(w, "challenge not found", http.StatusNotFound)
		return
	}

	resp := challengeStatusResponse{
		ChallengeID: challenge.ID.String(),
		Status:      string(challenge.Status),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
