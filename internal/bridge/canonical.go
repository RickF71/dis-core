package bridge

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// CanonicalTime returns the canonical RFC3339Nano UTC representation of t.
// Used in all signatures, receipts, and bridge exchanges.
func CanonicalTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// ParseCanonicalTime converts a canonical timestamp string back into time.Time.
func ParseCanonicalTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}

// CanonicalDigest returns the SHA-256 hex digest of the provided bytes.
func CanonicalDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// CanonicalJSON encodes any structure into stable, deterministic JSON.
// This ensures that two logically identical payloads produce identical hashes.
func CanonicalJSON(v interface{}) ([]byte, error) {
	normalized, err := normalizeJSON(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}

// normalizeJSON recursively sorts JSON object keys and normalizes nested structures.
// This prevents nondeterministic key order or float precision differences.
func normalizeJSON(v interface{}) (interface{}, error) {
	switch val := v.(type) {
	case map[string]interface{}:
		// Sort keys
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		sorted := make(map[string]interface{}, len(keys))
		for _, k := range keys {
			norm, err := normalizeJSON(val[k])
			if err != nil {
				return nil, err
			}
			sorted[k] = norm
		}
		return sorted, nil

	case []interface{}:
		// Normalize each array element
		arr := make([]interface{}, len(val))
		for i, elem := range val {
			norm, err := normalizeJSON(elem)
			if err != nil {
				return nil, err
			}
			arr[i] = norm
		}
		return arr, nil

	default:
		// primitive values (string, bool, float, nil) stay as-is
		return val, nil
	}
}

// CanonicalHashJSON takes any Go structure, canonicalizes it, and returns a hex SHA-256 digest.
func CanonicalHashJSON(v interface{}) (string, error) {
	b, err := CanonicalJSON(v)
	if err != nil {
		return "", err
	}
	return CanonicalDigest(b), nil
}

// VerifyCanonicalHash checks that the provided JSON structure matches a known canonical hash.
func VerifyCanonicalHash(v interface{}, expected string) (bool, error) {
	h, err := CanonicalHashJSON(v)
	if err != nil {
		return false, err
	}
	return h == expected, nil
}

// CanonicalReceipt represents the minimal schema of a bridge receipt.
type CanonicalReceipt struct {
	SchemaRef string `json:"schema_ref"`
	By        string `json:"by"`
	Scope     string `json:"scope"`
	Action    string `json:"action"`
	Nonce     string `json:"nonce"`
	CreatedAt string `json:"created_at"`
	Signature string `json:"signature"`
}

// Hash returns the canonical digest of the receipt's core fields.
func (r *CanonicalReceipt) Hash() (string, error) {
	b, err := CanonicalJSON(map[string]string{
		"schema_ref": r.SchemaRef,
		"by":         r.By,
		"scope":      r.Scope,
		"action":     r.Action,
		"nonce":      r.Nonce,
		"created_at": r.CreatedAt,
	})
	if err != nil {
		return "", err
	}
	return CanonicalDigest(b), nil
}

// Equal checks canonical equality of two receipts.
func (r *CanonicalReceipt) Equal(other *CanonicalReceipt) bool {
	h1, _ := r.Hash()
	h2, _ := other.Hash()
	return h1 == h2 && r.Signature == other.Signature
}

// PrettyJSON prints stable, human-friendly JSON for inspection or debugging.
func PrettyJSON(v interface{}) string {
	b, err := CanonicalJSON(v)
	if err != nil {
		return fmt.Sprintf("{\"error\": %q}", err.Error())
	}
	var out bytes.Buffer
	_ = json.Indent(&out, b, "", "  ")
	return out.String()
}
