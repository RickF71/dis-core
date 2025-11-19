/* --- dev.go (NITPICKED DROP-IN DEV AUTH ENDPOINTS) --- */

package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type DevUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var devUsers = []DevUser{
	{ID: "rick", Name: "Rick (Dev)"},
}

/*
GET /api/dev/users
Returns a static list of dev users.
*/
func GetDevUsers(w http.ResponseWriter, r *http.Request) {
	resp := struct {
		Users []DevUser `json:"users"`
	}{
		Users: devUsers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

/*
POST /auth/dev
Sets a session cookie and authenticates as the chosen dev user.
This is a simplified dev-only login using ONLY a user ID.
*/
func PostDevAuth(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		UserID string `json:"user_id"`
	}
	var req Req
	json.NewDecoder(r.Body).Decode(&req)

	if req.UserID == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}

	// TODO: Replace with your own session / identity logic
	// For now we simply set a dummy cookie containing the user_id
	cookie := &http.Cookie{
		Name:     "dis_session",
		Value:    req.UserID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   false,
	}
	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

/*
RegisterRoutesForDev adds BOTH endpoints to the router.
Call this from your main router setup.
*/
func RegisterRoutesForDev(r chi.Router) {
	r.Get("/api/dev/users", GetDevUsers)
	r.Post("/auth/dev", PostDevAuth)
}
