package quota

import "math"

// Item represents a quota calculation result.
type Item struct {
	Limit      *int64  `json:"limit"`
	Used       int64   `json:"used"`
	Remaining  *int64  `json:"remaining"`
	Percentage float64 `json:"percentage"`
}

// Calculate computes quota usage for a given limit.
func Calculate(used int64, limit *int64) Item {
	if limit == nil {
		return Item{Limit: nil, Used: used, Remaining: nil, Percentage: 0}
	}

	remaining := *limit - used
	if remaining < 0 {
		remaining = 0
	}

	percentage := 0.0
	if *limit > 0 {
		percentage = (float64(used) / float64(*limit)) * 100
	}

	return Item{
		Limit:      limit,
		Used:       used,
		Remaining:  &remaining,
		Percentage: percentage,
	}
}

// BytesToMB converts bytes to whole megabytes (floor).
func BytesToMB(bytes int64) int64 {
	if bytes <= 0 {
		return 0
	}
	return int64(math.Floor(float64(bytes) / (1024 * 1024)))
}
