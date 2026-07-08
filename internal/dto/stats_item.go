package dto

// @Swagger:model StatsItem
// @Property value: Current value of the metric
// @Property diff_last_month: Difference versus the previous month's value
// @Description A single dashboard metric with its month-over-month change.
type StatsItem struct {
	// Current value of the metric
	Value int `json:"value"`
	// Difference versus the previous month's value
	DiffLastMonth int `json:"diff_last_month"`
}
