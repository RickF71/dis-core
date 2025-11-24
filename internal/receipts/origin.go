package receipts

// NewDomainReceipt wraps a base Receipt and sets the origin domain fields
// so that every receipt clearly indicates the domain where the actor acted.
// Accept origin ID and name strings to avoid importing domain types.
func NewDomainReceipt(originID, originName string, base *Receipt) *Receipt {
	if base == nil {
		base = &Receipt{}
	}
	if originID != "" {
		base.OriginDomainID = originID
	}
	if originName != "" {
		base.OriginDomainName = originName
	}
	return base
}
