package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// CircuitBreaker states
const (
	StateClosed = iota
	StateOpen
	StateHalfOpen
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mutex                    sync.RWMutex
	state                    int
	failureCount             int
	failureThreshold         int
	resetTimeout             time.Duration
	lastFailureTime          time.Time
	halfOpenSuccess          int
	halfOpenSuccessThreshold int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration, halfOpenSuccessThreshold int) *CircuitBreaker {
	return &CircuitBreaker{
		state:                    StateClosed,
		failureThreshold:         failureThreshold,
		resetTimeout:             resetTimeout,
		halfOpenSuccessThreshold: halfOpenSuccessThreshold,
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if circuit is open
	if !cb.AllowRequest() {
		return ErrCircuitOpen
	}

	err := fn()

	// Handle result
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// AllowRequest checks if a request can be made
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if reset timeout has expired
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			// Transition to half-open
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = StateHalfOpen
			cb.halfOpenSuccess = 0
			cb.mutex.Unlock()
			cb.mutex.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordFailure records a failed request
func (cb *CircuitBreaker) recordFailure() {
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		cb.state = StateOpen
	}
}

// recordSuccess records a successful request
func (cb *CircuitBreaker) recordSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failureCount = 0
	case StateHalfOpen:
		cb.halfOpenSuccess++
		if cb.halfOpenSuccess >= cb.halfOpenSuccessThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}
