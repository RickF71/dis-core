package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func domainAssetPath(domain, file string) string {
	if domain == "" {
		domain = "domain.terra"
	}
	if file == "" {
		file = "App.css"
	}
	return filepath.Join("disyaml", "canon", domain, "ui", file)
}

func ServeDomainAsset(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	file := r.URL.Query().Get("file")
	path := domainAssetPath(domain, file)

	f, err := os.Open(path)
	if err != nil {
		http.Error(w, "Asset not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	// Basic content type
	switch filepath.Ext(file) {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js", ".jsx":
		w.Header().Set("Content-Type", "application/javascript")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// ETag based on content
	h := sha256.New()
	if _, err := io.Copy(h, f); err == nil {
		etag := hex.EncodeToString(h.Sum(nil))
		w.Header().Set("ETag", etag)
	}
	// re-open to stream (since we consumed it hashing)
	f.Seek(0, 0)
	io.Copy(w, f)
}

type assetMeta struct {
	ETag       string    `json:"etag"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
}

func ServeDomainAssetMeta(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	file := r.URL.Query().Get("file")
	path := domainAssetPath(domain, file)

	st, err := os.Stat(path)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	b, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "Read error", http.StatusInternalServerError)
		return
	}
	sum := sha256.Sum256(b)

	resp := assetMeta{
		ETag:       hex.EncodeToString(sum[:]),
		ModifiedAt: st.ModTime(),
		Size:       st.Size(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
