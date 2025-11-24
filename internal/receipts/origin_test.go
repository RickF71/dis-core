package receipts

import "testing"

func TestDomainReceiptSetsOrigin(t *testing.T) {
	originID := "42"
	originName := "domain.lima"
	r := NewDomainReceipt(originID, originName, &Receipt{})
	if r.OriginDomainID != "42" || r.OriginDomainName != "domain.lima" {
		t.Fatalf("origin not set correctly: got id=%q name=%q", r.OriginDomainID, r.OriginDomainName)
	}
}
