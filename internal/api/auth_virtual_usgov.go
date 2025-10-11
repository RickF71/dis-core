package api

import (
	"dis-core/internal/db"
	"encoding/json"
	"net/http"
)

type VirtualUSGovCredential struct {
	CredentialID  string   `json:"credential_id"`
	HolderUID     string   `json:"holder_uid"`
	ValidityScope string   `json:"validity_scope"`
	LinkedDomains []string `json:"linked_domains,omitempty"`
	JikkaRef      string   `json:"jikka_ref,omitempty"`
	Signature     string   `json:"signature"`
}

func HandleVirtualUSGovCredential(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var cred VirtualUSGovCredential
		if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		cred.CredentialID = "vusa-" + db.NowRFC3339Nano()
		resp := map[string]any{
			"status":     "issued",
			"credential": cred,
		}
		json.NewEncoder(w).Encode(resp)
	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}
