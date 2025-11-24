package receipts

import (
	"testing"
	"time"
)

func TestWrapLegacyReceipt(t *testing.T) {
	r := &Receipt{
		ID:               "r1",
		ReceiptType:      "ci.call.v1",
		EventID:          "e1",
		IssuedBy:         "actor-1",
		IssuedAt:         time.Now().UTC(),
		Verified:         true,
		Metadata:         map[string]any{"foo": "bar"},
		OriginDomainID:   "dom-1",
		OriginDomainName: "Example Domain",
	}
	originID := "dom-1"
	originName := "Example Domain"
	env := WrapLegacyReceipt(originID, originName, r)
	if env.ActionPanel == nil {
		t.Fatalf("action panel should be set")
	}
	if env.PolicyPanel["foo"] != "bar" {
		t.Fatalf("policy panel should contain legacy metadata")
	}
	if env.DomainPanel["origin_id"] != "dom-1" {
		t.Fatalf("domain panel should contain origin_id")
	}
}
