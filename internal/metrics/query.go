package metrics

import (
	"fmt"
	"sort"
	"time"
)

// Query provides methods for querying and analyzing collected metrics.
type Query struct {
	Store *Store
	// Now returns the current time. Defaults to time.Now.
	// Inject a fixed function in tests for deterministic behavior.
	Now func() time.Time
}

// NewQuery creates a new query engine.
func NewQuery(store *Store) *Query {
	return &Query{Store: store, Now: time.Now}
}

func (q *Query) now() time.Time {
	if q.Now != nil {
		return q.Now()
	}
	return time.Now()
}

// Summary produces a consolidated metrics snapshot.
func (q *Query) Summary(period time.Duration) (*MetricsSnapshot, error) {
	since := q.now().Add(-period)
	snapshots, err := q.Store.ReadSnapshots(since)
	if err != nil {
		return nil, fmt.Errorf("read snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no metrics data found. Run `mxf collect` first")
	}

	// Return the most recent snapshot
	latest := snapshots[len(snapshots)-1]

	// Compute health indicators
	latest.HealthIndicators = ComputeHealth(latest, snapshots)

	return &latest, nil
}

// Velocity returns velocity per sprint for the last N sprints.
func (q *Query) Velocity(sprints int) ([]VelocityPoint, error) {
	snapshots, err := q.Store.ReadSnapshots(time.Time{})
	if err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no metrics data found. Run `mxf collect` first")
	}

	points := ComputeVelocity(snapshots)

	if sprints > 0 && sprints < len(points) {
		points = points[len(points)-sprints:]
	}
	return points, nil
}

// CycleTime returns cycle time statistics for the given period.
func (q *Query) CycleTime(period time.Duration) (*CycleTimeStats, error) {
	since := q.now().Add(-period)
	snapshots, err := q.Store.ReadSnapshots(since)
	if err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no metrics data found. Run `mxf collect` first")
	}

	// Aggregate cycle time stats from snapshots
	var hours []float64
	for _, s := range snapshots {
		if s.CycleTime.Avg > 0 {
			hours = append(hours, s.CycleTime.Avg)
		}
	}

	stats := ComputeCycleTimeFromValues(hours)
	return &stats, nil
}

// Bottlenecks identifies pipeline stages with longest wait times.
func (q *Query) Bottlenecks() ([]BottleneckResult, error) {
	snapshots, err := q.Store.ReadSnapshots(time.Time{})
	if err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no metrics data found. Run `mxf collect` first")
	}

	// Compute bottlenecks from available data
	latest := snapshots[len(snapshots)-1]

	// Estimate stage wait times from available metrics
	results := []BottleneckResult{
		{Stage: "Review", AvgWaitDays: latest.ReviewIterations * 0.5},
		{Stage: "Testing", AvgWaitDays: latest.CycleTime.Avg / 24 * 0.3},
		{Stage: "Implementation", AvgWaitDays: latest.CycleTime.Avg / 24 * 0.5},
		{Stage: "Planning", AvgWaitDays: latest.LeadTime / 24 * 0.2},
	}

	// Sort by wait time descending
	sort.Slice(results, func(i, j int) bool {
		return results[j].AvgWaitDays < results[i].AvgWaitDays
	})

	return results, nil
}

// Health returns traffic-light health indicators.
func (q *Query) Health() ([]HealthIndicator, error) {
	snapshots, err := q.Store.ReadSnapshots(time.Time{})
	if err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no metrics data found. Run `mxf collect` first")
	}

	latest := snapshots[len(snapshots)-1]
	return ComputeHealth(latest, snapshots), nil
}
