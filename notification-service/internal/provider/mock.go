package provider

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

// MockProvider simulates a real notification provider.
// It introduces random latency and occasional failures for testing retry logic.
type MockProvider struct {
	failRate float64
}

func NewMockProvider() *MockProvider {
	return &MockProvider{failRate: 0.3}
}

func (m *MockProvider) Send(to, subject, body string) error {
	// Simulate network latency (100ms - 500ms).
	delay := time.Duration(100+rand.Intn(400)) * time.Millisecond
	time.Sleep(delay)

	// Simulate random failures.
	if rand.Float64() < m.failRate {
		return fmt.Errorf("mock provider: simulated transient failure sending to %s", to)
	}

	log.Printf("[MockProvider] Email sent to=%s subject=%q body=%q (latency=%v)",
		to, subject, body, delay)
	return nil
}
