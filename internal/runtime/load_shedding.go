package runtime

import (
	"fmt"
	"sync"
	"time"
)

// LoadShedder provides adaptive load shedding
type LoadShedder struct {
	threshold       int
	currentLoad     int
	rejectionRate   float64
	mu              sync.RWMutex
	metrics         *LoadMetrics
	rejectionPolicy RejectionPolicy
}

// RejectionPolicy defines how requests are rejected
type RejectionPolicy string

const (
	RejectionPolicyLinear      RejectionPolicy = "linear"
	RejectionPolicyCubic       RejectionPolicy = "cubic"
	RejectionPolicyExponential RejectionPolicy = "exponential"
)

// LoadMetrics tracks load metrics
type LoadMetrics struct {
	RequestCount     int64
	RejectedCount    int64
	AcceptedCount    int64
	AvgResponseTime  time.Duration
	P95ResponseTime  time.Duration
	P99ResponseTime  time.Duration
	LastUpdate       time.Time
	TotalProcessTime time.Duration
	BypassCount      int64
}

// NewLoadShedder creates a new load shedder
func NewLoadShedder(threshold int) *LoadShedder {
	return &LoadShedder{
		threshold:       threshold,
		currentLoad:     0,
		rejectionRate:   0.0,
		rejectionPolicy: RejectionPolicyLinear,
		metrics:         &LoadMetrics{LastUpdate: time.Now()},
	}
}

// SetRejectionPolicy sets the rejection policy
func (ls *LoadShedder) SetRejectionPolicy(policy RejectionPolicy) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.rejectionPolicy = policy
}

// ShouldReject determines if a request should be rejected
func (ls *LoadShedder) ShouldReject() bool {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	if ls.currentLoad < ls.threshold {
		return false
	}

	// Calculate rejection rate based on overload and policy
	overloadRatio := float64(ls.currentLoad-ls.threshold) / float64(ls.threshold)

	var rejectionRate float64
	switch ls.rejectionPolicy {
	case RejectionPolicyCubic:
		rejectionRate = overloadRatio * overloadRatio * overloadRatio * 0.5
	case RejectionPolicyExponential:
		rejectionRate = 1.0 - (1.0 / (1.0 + overloadRatio*overloadRatio))
	default: // Linear
		rejectionRate = overloadRatio * 0.1
	}

	// Clamp to [0, 1)
	if rejectionRate > 0.99 {
		rejectionRate = 0.99
	}

	ls.rejectionRate = rejectionRate
	return rejectionRate > 0.5
}

// GetRejectionRate returns the current rejection rate
func (ls *LoadShedder) GetRejectionRate() float64 {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.rejectionRate
}

// RecordRequest records a request
func (ls *LoadShedder) RecordRequest(rejected bool, responseTime time.Duration) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.metrics.RequestCount++
	ls.metrics.TotalProcessTime += responseTime

	if rejected {
		ls.metrics.RejectedCount++
	} else {
		ls.metrics.AcceptedCount++
		// Update average response time (exponential moving average)
		if ls.metrics.AvgResponseTime == 0 {
			ls.metrics.AvgResponseTime = responseTime
		} else {
			ls.metrics.AvgResponseTime = (ls.metrics.AvgResponseTime*3 + responseTime) / 4
		}
	}

	ls.metrics.LastUpdate = time.Now()
}

// UpdateLoad updates the current load
func (ls *LoadShedder) UpdateLoad(load int) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.currentLoad = load
}

// GetCurrentLoad returns the current load
func (ls *LoadShedder) GetCurrentLoad() int {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.currentLoad
}

// GetMetrics returns current metrics
func (ls *LoadShedder) GetMetrics() *LoadMetrics {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	return &LoadMetrics{
		RequestCount:     ls.metrics.RequestCount,
		RejectedCount:    ls.metrics.RejectedCount,
		AcceptedCount:    ls.metrics.AcceptedCount,
		AvgResponseTime:  ls.metrics.AvgResponseTime,
		P95ResponseTime:  ls.metrics.P95ResponseTime,
		P99ResponseTime:  ls.metrics.P99ResponseTime,
		LastUpdate:       ls.metrics.LastUpdate,
		TotalProcessTime: ls.metrics.TotalProcessTime,
		BypassCount:      ls.metrics.BypassCount,
	}
}

// SetThreshold sets a new threshold
func (ls *LoadShedder) SetThreshold(threshold int) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.threshold = threshold
}

// AdaptiveLoadShedder provides adaptive load shedding with dynamic thresholds
type AdaptiveLoadShedder struct {
	*LoadShedder
	adaptiveThreshold int
	history           []int
	maxHistory        int
	minThreshold      int
	maxThreshold      int
	mu                sync.RWMutex
	adaptInterval     time.Duration
	lastAdaptTime     time.Time
}

// NewAdaptiveLoadShedder creates a new adaptive load shedder
func NewAdaptiveLoadShedder(initialThreshold int) *AdaptiveLoadShedder {
	return &AdaptiveLoadShedder{
		LoadShedder:       NewLoadShedder(initialThreshold),
		adaptiveThreshold: initialThreshold,
		history:           make([]int, 0),
		maxHistory:        100,
		minThreshold:      int(float64(initialThreshold) * 0.5),
		maxThreshold:      int(float64(initialThreshold) * 2.0),
		adaptInterval:     10 * time.Second,
		lastAdaptTime:     time.Now(),
	}
}

// SetAdaptationInterval sets the adaptation interval
func (als *AdaptiveLoadShedder) SetAdaptationInterval(interval time.Duration) {
	als.mu.Lock()
	defer als.mu.Unlock()
	als.adaptInterval = interval
}

// Adapt adjusts the threshold based on historical load
func (als *AdaptiveLoadShedder) Adapt() {
	als.mu.Lock()
	defer als.mu.Unlock()

	// Check if enough time has passed
	if time.Since(als.lastAdaptTime) < als.adaptInterval {
		return
	}

	if len(als.history) < 10 {
		return // Need more data
	}

	// Calculate average and percentiles
	var sum int
	for _, load := range als.history {
		sum += load
	}
	avgLoad := sum / len(als.history)

	// Adjust threshold based on average
	threshold80 := int(float64(als.adaptiveThreshold) * 0.8)
	threshold60 := int(float64(als.adaptiveThreshold) * 0.6)

	oldThreshold := als.adaptiveThreshold

	if avgLoad > threshold80 {
		// System is hot, increase threshold
		als.adaptiveThreshold = int(float64(als.adaptiveThreshold) * 1.1)
	} else if avgLoad < threshold60 {
		// System is cool, decrease threshold
		als.adaptiveThreshold = int(float64(als.adaptiveThreshold) * 0.9)
	}

	// Clamp to min/max
	if als.adaptiveThreshold < als.minThreshold {
		als.adaptiveThreshold = als.minThreshold
	} else if als.adaptiveThreshold > als.maxThreshold {
		als.adaptiveThreshold = als.maxThreshold
	}

	als.LoadShedder.threshold = als.adaptiveThreshold
	als.lastAdaptTime = time.Now()

	if oldThreshold != als.adaptiveThreshold {
		fmt.Printf("LoadShedder: Threshold adapted from %d to %d (avgLoad=%d)\n",
			oldThreshold, als.adaptiveThreshold, avgLoad)
	}
}

// RecordLoad records load for adaptation
func (als *AdaptiveLoadShedder) RecordLoad(load int) {
	als.mu.Lock()
	defer als.mu.Unlock()

	als.history = append(als.history, load)
	if len(als.history) > als.maxHistory {
		als.history = als.history[1:]
	}

	als.LoadShedder.currentLoad = load
}

// GetHistoryStats returns statistics about load history
func (als *AdaptiveLoadShedder) GetHistoryStats() map[string]interface{} {
	als.mu.Lock()
	defer als.mu.Unlock()

	if len(als.history) == 0 {
		return map[string]interface{}{}
	}

	var sum, min, max int
	min = als.history[0]
	max = als.history[0]

	for _, load := range als.history {
		sum += load
		if load < min {
			min = load
		}
		if load > max {
			max = load
		}
	}

	avg := sum / len(als.history)

	return map[string]interface{}{
		"count":             len(als.history),
		"min":               min,
		"max":               max,
		"avg":               avg,
		"current_threshold": als.adaptiveThreshold,
		"min_threshold":     als.minThreshold,
		"max_threshold":     als.maxThreshold,
	}
}
