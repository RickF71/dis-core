package policy

type DomainPolicy struct {
	Gates     string `json:"gates"`
	Freeze    string `json:"freeze"`
	Risk      string `json:"risk"`
	CIRules   string `json:"ci_rules"`
	Redaction string `json:"redaction"`
	Inherit   string `json:"inherit"`
}
