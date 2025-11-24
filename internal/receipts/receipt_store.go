package receipts

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var ledgerLock sync.Mutex

// SaveEnvelope persists a ReceiptEnvelope to disk (individual file + ledger),
// computes PrevHash/Hash chain and enforces presence of origin domain.
func SaveEnvelope(env *ReceiptEnvelope) error {
	ledgerLock.Lock()
	defer ledgerLock.Unlock()

	dir := "receipts"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Ensure origin domain is present
	if env.DomainID == "" {
		// Also check domain panel for fallback
		if idv, ok := env.DomainPanel["origin_id"]; ok {
			if s, ok2 := idv.(string); ok2 {
				env.DomainID = s
			}
		}
	}
	if env.DomainID == "" {
		return fmt.Errorf("envelope missing origin domain")
	}

	// Try to obtain previous hash from ledger
	ledgerFile := filepath.Join(dir, "ledger.jsonl")
	prevHash := ""
	if f, err := os.Open(ledgerFile); err == nil {
		defer f.Close()
		// Read last non-empty line
		var lastLine string
		r := bufio.NewReader(f)
		for {
			line, err := r.ReadString('\n')
			if err != nil && err != io.EOF {
				break
			}
			if len(line) > 1 {
				lastLine = line
			}
			if err == io.EOF {
				break
			}
		}
		if lastLine != "" {
			var last map[string]any
			if err := json.Unmarshal([]byte(lastLine), &last); err == nil {
				if h, ok := last["hash"].(string); ok {
					prevHash = h
				}
			}
		}
	}
	env.PrevHash = prevHash

	// Compute current hash (sha256 of marshaled envelope without the hash field set)
	env.Hash = "" // ensure empty during computation
	raw, err := json.Marshal(env)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(raw)
	env.Hash = hex.EncodeToString(sum[:])

	// --- 1️⃣ Save individual file ---
	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}
	filename := filepath.Join(dir, fmt.Sprintf("%s.json", env.ID))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return err
	}

	// --- 2️⃣ Append to rolling ledger file ---
	lf, err := os.OpenFile(ledgerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer lf.Close()

	flatData, _ := json.Marshal(env)
	if _, err := lf.Write(append(flatData, '\n')); err != nil {
		return err
	}

	log.Printf("📜 Saved envelope → %s", filename)
	return nil
}

// SaveReceipt remains for backward compatibility: it wraps the legacy Receipt
// into a ReceiptEnvelope and delegates to SaveEnvelope.
func SaveReceipt(r *Receipt) error {
	env := WrapLegacyReceipt(r.OriginDomainID, r.OriginDomainName, r)
	// Preserve origin if present on legacy receipt
	if r.OriginDomainID != "" {
		env.DomainID = r.OriginDomainID
		env.DomainPanel["origin_id"] = r.OriginDomainID
	}
	if r.OriginDomainName != "" {
		env.DomainPanel["origin_name"] = r.OriginDomainName
	}
	return SaveEnvelope(env)
}

// SaveRawReceipt preserves backward compatibility for any legacy JSON
// payloads that are pre-marshaled.
func SaveRawReceipt(data []byte) error {
	var r Receipt
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	return SaveReceipt(&r)
}
