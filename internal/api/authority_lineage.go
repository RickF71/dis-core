package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

func HandleAuthorityLineage(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	domainID := parts[len(parts)-1]
	resp := AuthorityEngine.BuildLineage(domainID)
	json.NewEncoder(w).Encode(resp)
}
