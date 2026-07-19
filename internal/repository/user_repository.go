package repository

import (
	"context"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/query"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *domain.User) (*domain.User, error) {
	ctx := context.Background()
	err := gorm.G[domain.User](r.db).Create(ctx, user)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_users_")
	}
	return user, nil
}

func (r *UserRepository) FindByID(id string) (*domain.User, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.User](r.db).
		Where("id = ? AND deleted_at IS NULL", id).
		First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_users_")
	}
	return &data, nil
}

func (r *UserRepository) FindByPhone(phone string) (*domain.User, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.User](r.db).
		Where("phone = ? AND deleted_at IS NULL", phone).
		First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_users_")
	}
	return &data, nil
}

func (r *UserRepository) Update(user *domain.User) (*domain.User, error) {
	ctx := context.Background()
	_, err := gorm.G[domain.User](r.db).
		Where("id = ? AND deleted_at IS NULL", user.ID).
		Updates(ctx, *user)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_users_")
	}
	return user, nil
}

func (r *UserRepository) Count() (int64, error) {
	n, err := gorm.G[domain.User](r.db).Where("deleted_at IS NULL").Count(context.Background(), "*")
	if err != nil {
		return 0, util.ParseDatabaseError(err, "idx_users_")
	}
	return n, nil
}

func (r *UserRepository) SetStatus(userID string, status domain.UserStatus) (int64, error) {
	res := r.db.WithContext(context.Background()).
		Model(&domain.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("status", status)
	if res.Error != nil {
		return 0, util.ParseDatabaseError(res.Error, "idx_users_")
	}
	return res.RowsAffected, nil
}

func (r *UserRepository) SetCurrentCity(userID string, cityID *string) error {
	err := r.db.WithContext(context.Background()).
		Model(&domain.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("current_city_id", cityID).Error
	if err != nil {
		return util.ParseDatabaseError(err, "idx_users_")
	}
	return nil
}

type UserFilters struct {
	Status *domain.UserStatus
	Search string
}

func (r *UserRepository) List(f UserFilters, p query.Page, srt query.Sort) ([]domain.User, int64, error) {
	ctx := context.Background()

	q := gorm.G[domain.User](r.db).Where("deleted_at IS NULL")
	if f.Status != nil {
		q = q.Where("status = ?", string(*f.Status))
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("(name ILIKE ? OR phone ILIKE ?)", like, like)
	}

	total, err := q.Count(ctx, "*")
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_users_")
	}

	items, err := q.
		Order(srt.OrderClause()).
		Limit(p.Limit()).
		Offset(p.Offset()).
		Find(ctx)
	if err != nil {
		return nil, 0, util.ParseDatabaseError(err, "idx_users_")
	}

	return items, total, nil
}

func (r *UserRepository) Delete(user *domain.User) error {
	ctx := context.Background()
	_, err := gorm.G[domain.User](r.db).Where("id = ?", user.ID).Delete(ctx)
	if err != nil {
		return util.ParseDatabaseError(err, "idx_users_")
	}
	return nil
}
