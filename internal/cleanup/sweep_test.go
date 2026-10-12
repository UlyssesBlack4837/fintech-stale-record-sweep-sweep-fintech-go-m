package cleanup

import (
	"testing"
	"time"
)

func TestPolicyEvaluate(t *testing.T) {
	now := time.Date(2026, time.September, 6, 8, 0, 0, 0, time.UTC)
	stale := now.Add(-91 * 24 * time.Hour)
	recent := now.Add(-2 * time.Hour)
	policy := Policy{StaleAfter: 90 * 24 * time.Hour, RiskFloor: 70}

	tests := []struct {
		name   string
		event  PaymentEvent
		want   Action
		reason string
	}{
		{"stale low risk deletes", PaymentEvent{ID: "pay_101", LastUpdated: stale, RiskScore: 12}, Delete, "stale low-risk record is eligible for cleanup"},
		{"stale high risk routes to review", PaymentEvent{ID: "pay_102", LastUpdated: stale, RiskScore: 88}, Review, "stale high-risk record requires review"},
		{"legal hold wins", PaymentEvent{ID: "pay_103", LastUpdated: stale, RiskScore: 3, LegalHold: true}, Keep, "legal hold preserves the record"},
		{"recent record stays", PaymentEvent{ID: "pay_104", LastUpdated: recent, RiskScore: 5}, Keep, "record is inside the retention window"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := policy.Evaluate(tt.event, now)
			if got.Action != tt.want || got.Notification.Action != tt.want {
				t.Fatalf("action = %q, notification action = %q; want %q", got.Action, got.Notification.Action, tt.want)
			}
			if got.Notification.EventID != tt.event.ID || got.Notification.Reason != tt.reason || !got.Notification.EvaluatedAt.Equal(now) {
				t.Fatalf("unexpected audit notification: %+v", got.Notification)
			}
		})
	}
}
