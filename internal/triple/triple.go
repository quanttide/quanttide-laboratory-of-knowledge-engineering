package triple

import "time"

type Verification string

const (
	VerificationUnverified  Verification = "unverified"
	VerificationVerified    Verification = "verified"
	VerificationMalformed   Verification = "malformed"
	VerificationPendingMerge Verification = "pending_merge"
	VerificationDiscarded   Verification = "discarded"
)

type Triple struct {
	ID           string            `json:"id"`
	Source       string            `json:"source"`
	Timestamp    string            `json:"timestamp"`
	Confidence   float64           `json:"confidence"`
	Subject      string            `json:"subject"`
	Predicate    string            `json:"predicate"`
	Object       string            `json:"object"`
	Context      map[string]string `json:"context,omitempty"`
	Verification Verification      `json:"verification"`
}

func New(id, source string, confidence float64, subject, predicate, object string) *Triple {
	return &Triple{
		ID:           id,
		Source:       source,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Confidence:   confidence,
		Subject:      subject,
		Predicate:    predicate,
		Object:       object,
		Context:      make(map[string]string),
		Verification: VerificationUnverified,
	}
}

func (t *Triple) MarkMalformed() {
	t.Verification = VerificationMalformed
}

func (t *Triple) MarkVerified() {
	t.Verification = VerificationVerified
}

func (t *Triple) SetContext(key, value string) {
	if t.Context == nil {
		t.Context = make(map[string]string)
	}
	t.Context[key] = value
}
