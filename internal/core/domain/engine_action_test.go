package domain

import (
	"context"
	"testing"

	"dis-core/internal/receipts"
)

type mockDB struct{}

func (m *mockDB) LoadDomainByID(ctx context.Context, id string) (any, error) {
	return &Domain{ID: id, Name: "domain.lima"}, nil
}

type mockPolicy struct{}

func (m *mockPolicy) EvalFn(ctx context.Context, ns string, input map[string]any) (map[string]any, error) {
	return map[string]any{"allow": true}, nil
}

type mockReceipts struct {
	last *receipts.ReceiptEnvelope
}

func (m *mockReceipts) SaveEnvelope(ctx context.Context, env any) error {
	m.last = env.(*receipts.ReceiptEnvelope)
	return nil
}

func TestHandleActionCreatesEnvelope(t *testing.T) {
	e := NewEngine(&mockDB{}, &mockPolicy{}, &mockReceipts{})

	dom := &Domain{ID: "123", Name: "domain.lima"}
	ctx := context.WithValue(context.Background(), "actor_id", "actor-1")

	err := e.HandleAction(ctx, dom, "test.action", map[string]any{"x": 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
