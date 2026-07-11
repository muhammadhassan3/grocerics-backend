package repository

import (
	"context"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

// ---------- notification preferences ----------

type NotificationPreferenceRepository struct{ db *gorm.DB }

func NewNotificationPreferenceRepository(db *gorm.DB) *NotificationPreferenceRepository {
	return &NotificationPreferenceRepository{db: db}
}

func (r *NotificationPreferenceRepository) Get(userID string) (*domain.NotificationPreference, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.NotificationPreference](r.db).
		Where("user_id = ? AND deleted_at IS NULL", userID).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_notification_preferences_")
	}
	return &data, nil
}

func (r *NotificationPreferenceRepository) Upsert(p *domain.NotificationPreference) (*domain.NotificationPreference, error) {
	existing, err := r.Get(p.UserID)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if existing == nil {
		if cErr := gorm.G[domain.NotificationPreference](r.db).Create(ctx, p); cErr != nil {
			return nil, util.ParseDatabaseError(cErr, "idx_notification_preferences_")
		}
		return p, nil
	}
	existing.PriceAlerts = p.PriceAlerts
	existing.Promotions = p.Promotions
	existing.OrderUpdates = p.OrderUpdates
	if _, uErr := gorm.G[domain.NotificationPreference](r.db).Where("id = ?", existing.ID).Updates(ctx, *existing); uErr != nil {
		return nil, util.ParseDatabaseError(uErr, "idx_notification_preferences_")
	}
	return existing, nil
}

// ---------- FCM tokens ----------

type FcmTokenRepository struct{ db *gorm.DB }

func NewFcmTokenRepository(db *gorm.DB) *FcmTokenRepository { return &FcmTokenRepository{db: db} }

func (r *FcmTokenRepository) Upsert(userID, token string, platform domain.DevicePlatform) error {
	ctx := context.Background()
	data, err := gorm.G[domain.FcmToken](r.db).Where("token = ? AND deleted_at IS NULL", token).First(ctx)
	if err != nil && err != gorm.ErrRecordNotFound {
		return util.ParseDatabaseError(err, "idx_fcm_tokens_")
	}
	if err == gorm.ErrRecordNotFound {
		t := &domain.FcmToken{UserID: userID, Token: token, Platform: platform}
		if cErr := gorm.G[domain.FcmToken](r.db).Create(ctx, t); cErr != nil {
			return util.ParseDatabaseError(cErr, "idx_fcm_tokens_")
		}
		return nil
	}
	data.UserID = userID
	data.Platform = platform
	if _, uErr := gorm.G[domain.FcmToken](r.db).Where("id = ?", data.ID).Updates(ctx, data); uErr != nil {
		return util.ParseDatabaseError(uErr, "idx_fcm_tokens_")
	}
	return nil
}

func (r *FcmTokenRepository) Delete(token string) error {
	ctx := context.Background()
	if _, err := gorm.G[domain.FcmToken](r.db).Where("token = ?", token).Delete(ctx); err != nil {
		return util.ParseDatabaseError(err, "idx_fcm_tokens_")
	}
	return nil
}
