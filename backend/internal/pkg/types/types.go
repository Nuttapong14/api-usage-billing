package types

import "time"

// Pagination captures common list parameters.
type Pagination struct {
	Limit  int
	Offset int
}

// TimeRange represents a start/end window for queries.
type TimeRange struct {
	Start time.Time
	End   time.Time
}

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)
