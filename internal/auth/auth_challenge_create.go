package auth

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type ChallengeCreateHandler struct {
	Store ChallengeStore
}

type challengeCreateRequest struct {
	ExternalUserID string `json:"external_user_id"`
}

type challengeCreateResponse struct {
	ChallengeID string `json:"challenge_id"`
	Status      string `json:"status"`
}

func NewChallengeCreateHandler(store ChallengeStore) http.Handler {
	return &ChallengeCreateHandler{Store: store}
}

func (h *ChallengeCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req challengeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	externalUserID, err := uuid.Parse(req.ExternalUserID)
	if err != nil {
		http.Error(w, "invalid external_user_id", http.StatusBadRequest)
		return
	}

	challenge := NewChallenge(externalUserID)

	if err := h.Store.SaveChallenge(r.Context(), challenge); err != nil {
		http.Error(w, "failed to create challenge", http.StatusInternalServerError)
		return
	}

	resp := challengeCreateResponse{
		ChallengeID: challenge.ID.String(),
		Status:      string(challenge.Status),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
