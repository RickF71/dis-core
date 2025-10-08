package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"dis-core/internal/config"
	"dis-core/internal/core"
	"dis-core/internal/policy"
)

type APIServer struct {
	db       *sql.DB
	cfg      *config.Config
	pol      *policy.Policy
	polSum   string
	coreHash string
}

func NewServer(db *sql.DB, cfg *config.Config, pol *policy.Policy, polSum string, coreHash string) *APIServer {
	return &APIServer{
		db:       db,
		cfg:      cfg,
		pol:      pol,
		polSum:   polSum,
		coreHash: coreHash,
	}
}

func (s *APIServer) Start(addr string) error {
	http.HandleFunc("/health", s.health)
	http.HandleFunc("/policy", s.getPolicy)
	http.HandleFunc("/receipts", s.getReceipts)
	http.HandleFunc("/act", s.postAct)
	return http.ListenAndServe(addr, nil)
}

func (s *APIServer) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "version": "v0.5", "time": time.Now().UTC()})
}

func (s *APIServer) getPolicy(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"checksum": s.polSum, "policy": s.pol})
}

func (s *APIServer) getReceipts(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(`SELECT id, by_domain, scope, nonce, timestamp, policy_checksum, signature FROM receipts ORDER BY created_at DESC`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	type rec struct {
		ReceiptID int    `json:"receipt_id"`
		By        string `json:"by"`
		Scope     string `json:"scope"`
		Nonce     string `json:"nonce"`
		Timestamp string `json:"timestamp"`
		PolicySum string `json:"policy_checksum"`
		Signature string `json:"signature"`
	}
	list := []rec{}
	for rows.Next() {
		var r rec
		if err := rows.Scan(&r.ReceiptID, &r.By, &r.Scope, &r.Nonce, &r.Timestamp, &r.PolicySum, &r.Signature); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		list = append(list, r)
	}
	writeJSON(w, http.StatusOK, list)
}

type actReq struct {
	By    string `json:"by"`
	Scope string `json:"scope"`
	Nonce string `json:"nonce"`
}

func (s *APIServer) postAct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req actReq
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	by := req.By
	if by == "" {
		by = s.cfg.DefaultDomain
	}
	scope := req.Scope
	if scope == "" {
		scope = s.cfg.DefaultScope
	}
	recID, nonce, ts, sig, err := core.PerformConsentAction(s.db, by, scope, req.Nonce, s.cfg, s.pol, s.polSum)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":          "ok",
		"action":          "consent:grant",
		"by":              by,
		"scope":           scope,
		"timestamp":       ts,
		"nonce":           nonce,
		"policy_checksum": s.polSum,
		"signature":       sig,
		"receipt_id":      recID,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
