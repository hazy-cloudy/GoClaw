package proactive

type Intent struct {
	Type        string         `json:"type"`
	Priority    string         `json:"priority"`
	ReasonCodes []string       `json:"reason_codes"`
	Payload     map[string]any `json:"payload"`
}

type Provider interface {
	Name() string
	Evaluate(snapshot Snapshot) (*Intent, bool, error)
}
