package runtime

import (
	"sync"
	"time"
)

// LoadShedder provides adaptive load shedding
type LoadShedder struct {
	threshold      int
	currentLoad    int
	rejectionRate  float64
	mu             sync.RWMutex
	metrics        *LoadMetrics
}

// LoadMetrics tracks load metrics
type LoadMetrics struct {
	RequestCount    int64
	RejectedCount   int64
	AvgResponseTime time.Duration
	LastUpdate      time.Time
}

// NewLoadShedder creates a new load shedder
func NewLoadShedder(threshold int) *LoadShedder {
	return &LoadShedder{
		threshold:     threshold,
		currentLoad:   0,
		rejectionRate: 0.0,
		metrics: &LoadMetrics{
			LastUpdate: time.Now(),
		},
	}
}

// ShouldReject determines if a request should be rejected
func (ls *LoadShedder) ShouldReject() bool {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	
	if ls.currentLoad < ls.threshold {
		return false
	}
	
	// Calculate rejection rate based on overload
	overload := float64(ls.currentLoad-ls.threshold) / float64(ls.threshold)
	ls.rejectionRate = overload * 0.1 // Reject 10% per 100% overload
	
	// Simple probabilistic rejection
	return ls.rejectionRate > 0.5
}

// RecordRequest records a request
func (ls *LoadShedder) RecordRequest(rejected bool, responseTime time.Duration) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	
	ls.metrics.RequestCount++
	if rejected {
		ls.metrics.RejectedCount++
	} else {
		// Update average response time
		if ls.metrics.AvgResponseTime == 0 {
			ls.metrics.AvgResponseTime = responseTime
		} else {
			ls.metrics.AvgResponseTime = (ls.metrics.AvgResponseTime + responseTime) / 2
		}
	}
}

// UpdateLoad updates the current load
func (ls *LoadShedder) UpdateLoad(load int) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.currentLoad = load
}

// GetMetrics returns current metrics
func (ls *LoadShedder) GetMetrics() *LoadMetrics {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	
	return &LoadMetrics{
		RequestCount:    ls.metrics.RequestCount,
		RejectedCount:   ls.metrics.RejectedCount,
		AvgResponseTime: ls.metrics.AvgResponseTime,
		LastUpdate:      ls.metrics.LastUpdate,
	}
}

// AdaptiveLoadShedder provides adaptive load shedding with dynamic thresholds
type AdaptiveLoadShedder struct {
	*LoadShedder
	adaptiveThreshold int
	history           []int
	maxHistory        int
	mu                sync.RWMutex
}

// NewAdaptiveLoadShedder creates a new adaptive load shedder
func NewAdaptiveLoadShedder(initialThreshold int) *AdaptiveLoadShedder {
	return &AdaptiveLoadShedder{
		LoadShedder:      NewLoadShedder(initialThreshold),
		adaptiveThreshold: initialThreshold,
		history:          make([]int, 0),
		maxHistory:       100,
	}
}

// Adapt adjusts the threshold based on historical load
func (als *AdaptiveLoadShedder) Adapt() {
	als.mu.Lock()
	defer als.mu.Unlock()
	
	if len(als.history) < 10 {
		return // Need more data
	}
	
	// Calculate average load
	var sum int
	for _, load := range als.history {
		sum += load
	}
	avgLoad := sum / len(als.history)
	
	// Adjust threshold based on average
	threshold80 := int(float64(als.adaptiveThreshold) * 0.8)
	threshold50 := int(float64(als.adaptiveThreshold) * 0.5)
	if avgLoad > threshold80 {
		als.adaptiveThreshold = int(float64(als.adaptiveThreshold) * 1.1)
	} else if avgLoad < threshold50 {
		als.adaptiveThreshold = int(float64(als.adaptiveThreshold) * 0.9)
	}
	
	als.LoadShedder.threshold = als.adaptiveThreshold
}

// RecordLoad records load for adaptation
func (als *AdaptiveLoadShedder) RecordLoad(load int) {
	als.mu.Lock()
	defer als.mu.Unlock()
	
	als.history = append(als.history, load)
	if len(als.history) > als.maxHistory {
		als.history = als.history[1:]
	}
	
	als.UpdateLoad(load)
}

