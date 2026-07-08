package dto

// @Swagger:model DashboardStats
// @Property total_users: Total registered users
// @Property average_basket_size: Average number of items per basket
// @Property total_searches: Total product searches performed
// @Description Headline stat cards shown on the admin dashboard.
type DashboardStats struct {
	// Total registered users
	TotalUsers StatsItem `json:"total_users"`
	// Average number of items per basket
	AverageBasketSize StatsItem `json:"average_basket_size"`
	// Total product searches performed
	TotalSearches StatsItem `json:"total_searches"`
}

// @Swagger:model DailyActiveUsers
// @Property monday: Active users on Monday
// @Property tuesday: Active users on Tuesday
// @Property wednesday: Active users on Wednesday
// @Property thursday: Active users on Thursday
// @Property friday: Active users on Friday
// @Property saturday: Active users on Saturday
// @Property sunday: Active users on Sunday
// @Description Daily active user counts for the current week, keyed by day.
type DailyActiveUsers struct {
	// Active users on Monday
	Monday int `json:"monday"`
	// Active users on Tuesday
	Tuesday int `json:"tuesday"`
	// Active users on Wednesday
	Wednesday int `json:"wednesday"`
	// Active users on Thursday
	Thursday int `json:"thursday"`
	// Active users on Friday
	Friday int `json:"friday"`
	// Active users on Saturday
	Saturday int `json:"saturday"`
	// Active users on Sunday
	Sunday int `json:"sunday"`
}

// @Swagger:model MonthlyActiveUsers
// @Property january: Active users in January
// @Property february: Active users in February
// @Property march: Active users in March
// @Property april: Active users in April
// @Property may: Active users in May
// @Property june: Active users in June
// @Property july: Active users in July
// @Property august: Active users in August
// @Property september: Active users in September
// @Property october: Active users in October
// @Property november: Active users in November
// @Property december: Active users in December
// @Description Monthly active user counts for the current year, keyed by month.
type MonthlyActiveUsers struct {
	// Active users in January
	January int `json:"january"`
	// Active users in February
	February int `json:"february"`
	// Active users in March
	March int `json:"march"`
	// Active users in April
	April int `json:"april"`
	// Active users in May
	May int `json:"may"`
	// Active users in June
	June int `json:"june"`
	// Active users in July
	July int `json:"july"`
	// Active users in August
	August int `json:"august"`
	// Active users in September
	September int `json:"september"`
	// Active users in October
	October int `json:"october"`
	// Active users in November
	November int `json:"november"`
	// Active users in December
	December int `json:"december"`
}

// @Swagger:model DashboardResponse
// @Property stats: Headline stat cards
// @Property daily_active_users: Active user counts for the current week
// @Property monthly_active_users: Active user counts for the current year
// @Description Envelope for the admin dashboard endpoint.
type DashboardResponse struct {
	// Headline stat cards
	Stats DashboardStats `json:"stats"`
	// Active user counts for the current week
	DailyActiveUsers DailyActiveUsers `json:"daily_active_users"`
	// Active user counts for the current year
	MonthlyActiveUsers MonthlyActiveUsers `json:"monthly_active_users"`
}
