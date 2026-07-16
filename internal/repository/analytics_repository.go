package repository

import (
	"context"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/query"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

type AnalyticsRepository struct{ db *gorm.DB }

func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

type ProductSearchCount struct {
	ProductID   string `gorm:"column:product_id"`
	SearchCount int    `gorm:"column:search_count"`
}

func (r *AnalyticsRepository) MarkActive(userID string) error {
	err := r.db.WithContext(context.Background()).Exec(
		`INSERT INTO user_activity_daily (user_id, activity_date)
		 VALUES (?, CURRENT_DATE)
		 ON CONFLICT (user_id, activity_date) DO NOTHING`, userID).Error
	if err != nil {
		return util.ParseDatabaseError(err, "idx_user_activity_daily_")
	}
	return nil
}

func (r *AnalyticsRepository) LogSearch(e *domain.SearchEvent) error {
	if err := gorm.G[domain.SearchEvent](r.db).Create(context.Background(), e); err != nil {
		return util.ParseDatabaseError(err, "idx_search_events_")
	}
	return nil
}

func (r *AnalyticsRepository) DAU() (map[int]int, error) {
	var rows []struct {
		Dow int
		Cnt int
	}
	err := r.db.WithContext(context.Background()).Raw(
		`SELECT EXTRACT(ISODOW FROM activity_date)::int AS dow, count(*) AS cnt
		 FROM user_activity_daily
		 WHERE activity_date >= date_trunc('week', CURRENT_DATE)
		   AND activity_date <  date_trunc('week', CURRENT_DATE) + INTERVAL '7 days'
		 GROUP BY dow`).Scan(&rows).Error
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_user_activity_daily_")
	}
	out := make(map[int]int, len(rows))
	for _, x := range rows {
		out[x.Dow] = x.Cnt
	}
	return out, nil
}

func (r *AnalyticsRepository) MAU() (map[int]int, error) {
	var rows []struct {
		Mon int
		Cnt int
	}
	err := r.db.WithContext(context.Background()).Raw(
		`SELECT EXTRACT(MONTH FROM activity_date)::int AS mon, count(DISTINCT user_id) AS cnt
		 FROM user_activity_daily
		 WHERE activity_date >= date_trunc('year', CURRENT_DATE)
		   AND activity_date <  date_trunc('year', CURRENT_DATE) + INTERVAL '1 year'
		 GROUP BY mon`).Scan(&rows).Error
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_user_activity_daily_")
	}
	out := make(map[int]int, len(rows))
	for _, x := range rows {
		out[x.Mon] = x.Cnt
	}
	return out, nil
}

func (r *AnalyticsRepository) SearchStats() (total, thisMonth, lastMonth int, err error) {
	var row struct {
		Total     int
		ThisMonth int
		LastMonth int
	}
	err = r.db.WithContext(context.Background()).Raw(
		`SELECT
		   count(*) AS total,
		   count(*) FILTER (WHERE created_at >= date_trunc('month', now())) AS this_month,
		   count(*) FILTER (WHERE created_at >= date_trunc('month', now()) - INTERVAL '1 month'
		                      AND created_at <  date_trunc('month', now())) AS last_month
		 FROM search_events`).Scan(&row).Error
	if err != nil {
		return 0, 0, 0, util.ParseDatabaseError(err, "idx_search_events_")
	}
	return row.Total, row.ThisMonth, row.LastMonth, nil
}

func (r *AnalyticsRepository) NewUserMonthlyDiff() (int, error) {
	var diff int
	err := r.db.WithContext(context.Background()).Raw(
		`SELECT
		   count(*) FILTER (WHERE created_at >= date_trunc('month', now()))
		   - count(*) FILTER (WHERE created_at >= date_trunc('month', now()) - INTERVAL '1 month'
		                        AND created_at <  date_trunc('month', now())) AS diff
		 FROM users
		 WHERE deleted_at IS NULL`).Scan(&diff).Error
	if err != nil {
		return 0, util.ParseDatabaseError(err, "idx_users_")
	}
	return diff, nil
}

func (r *AnalyticsRepository) AverageBasketItems() (int, error) {
	var avg int
	err := r.db.WithContext(context.Background()).Raw(
		`SELECT COALESCE(ROUND(AVG(item_count)), 0)::int
		 FROM (
		   SELECT cart_id, SUM(quantity) AS item_count
		   FROM cart_items
		   WHERE deleted_at IS NULL
		   GROUP BY cart_id
		 ) t`).Scan(&avg).Error
	if err != nil {
		return 0, util.ParseDatabaseError(err, "idx_cart_items_")
	}
	return avg, nil
}

func (r *AnalyticsRepository) TopSearchedProducts(p query.Page) ([]ProductSearchCount, int64, error) {
	ctx := context.Background()
	var total int64
	if err := r.db.WithContext(ctx).Raw(
		`SELECT count(DISTINCT result_product_id)
		 FROM search_events
		 WHERE result_product_id IS NOT NULL`).Scan(&total).Error; err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_search_events_")
	}
	var rows []ProductSearchCount
	if err := r.db.WithContext(ctx).Raw(
		`SELECT result_product_id AS product_id, count(*) AS search_count
		 FROM search_events
		 WHERE result_product_id IS NOT NULL
		 GROUP BY result_product_id
		 ORDER BY search_count DESC, result_product_id
		 LIMIT ? OFFSET ?`, p.Limit(), p.Offset()).Scan(&rows).Error; err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_search_events_")
	}
	return rows, total, nil
}
