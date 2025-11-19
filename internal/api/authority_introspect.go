package api

import (
	"encoding/json"
	"net/http"
)

func HandleAuthorityIntrospect(w http.ResponseWriter, r *http.Request) {
	resp := AuthorityEngine.Introspect()
	json.NewEncoder(w).Encode(resp)
}
