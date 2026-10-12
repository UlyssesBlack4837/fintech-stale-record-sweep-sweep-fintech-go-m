package cleanup

import "time"

type PaymentEvent struct {
	ID          string    `json:"id"`
	State       string    `json:"state"`
	LastUpdated time.Time `json:"last_updated"`
	RiskScore   int       `json:"risk_score"`
	LegalHold   bool      `json:"legal_hold"`
	AmountCents int64     `json:"amount_cents"`
}

type Action string

const (
	Delete Action = "delete"
	Review Action = "review"
	Keep   Action = "keep"
)

type AuditNotification struct {
	EventID     string    `json:"event_id"`
	Action      Action    `json:"action"`
	Reason      string    `json:"reason"`
	EvaluatedAt time.Time `json:"evaluated_at"`
}

type Result struct {
	Event        PaymentEvent      `json:"event"`
	Action       Action            `json:"action"`
	Notification AuditNotification `json:"notification"`
}

type Policy struct {
	StaleAfter time.Duration
	RiskFloor  int
}

func (p Policy) Evaluate(event PaymentEvent, now time.Time) Result {
	action, reason := Keep, "record is inside the retention window"
	if event.LegalHold {
		reason = "legal hold preserves the record"
	} else if now.Sub(event.LastUpdated) >= p.StaleAfter {
		if event.RiskScore >= p.RiskFloor {
			action, reason = Review, "stale high-risk record requires review"
		} else {
			action, reason = Delete, "stale low-risk record is eligible for cleanup"
		}
	}

	return Result{
		Event:  event,
		Action: action,
		Notification: AuditNotification{
			EventID: event.ID, Action: action, Reason: reason, EvaluatedAt: now.UTC(),
		},
	}
}

func (p Policy) Sweep(events []PaymentEvent, now time.Time) []Result {
	results := make([]Result, 0, len(events))
	for _, event := range events {
		results = append(results, p.Evaluate(event, now))
	}
	return results
}
