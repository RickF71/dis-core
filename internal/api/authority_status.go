package api

import (
    "net/http"
    "encoding/json"
)

func HandleAuthorityRoot(w http.ResponseWriter, r *http.Request) {
    resp := AuthorityEngine.Status()
    json.NewEncoder(w).Encode(resp)
}
