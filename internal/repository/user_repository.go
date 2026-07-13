package repository

import (
	"context"
	"strings"

	"grocerics-backend/internal/auth"
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

// normalizeEmail is the single canonicalisation point for user emails.
// Both writes (Create / Update) and reads (FindByEmail) route through
// this so callers cannot accidentally bypass it.
func normalizeEmail(e string) string {
	return strings.ToLower(strings.TrimSpace(e))
}

func (r *UserRepository) Create(user *domain.User) (*domain.User, error) {
	user.Email = normalizeEmail(user.Email)
	ctx := context.Background()
	err := gorm.G[domain.User](r.db).Create(ctx, user)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_users_")
	}
	return user, nil
}

// FindByEmail returns the active (non-soft-deleted) user with this email,
// or nil if none. The partial unique index guarantees there is at most one.
func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.User](r.db).
		Where("email = ? AND deleted_at IS NULL", normalizeEmail(email)).
		First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_users_")
	}
	return &data, nil
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

func (r *UserRepository) Update(user *domain.User) (*domain.User, error) {
	if user.Email != "" {
		user.Email = normalizeEmail(user.Email)
	}
	ctx := context.Background()
	_, err := gorm.G[domain.User](r.db).
		Where("id = ? AND deleted_at IS NULL", user.ID).
		Updates(ctx, *user)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_users_")
	}
	return user, nil
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

func (r *UserRepository) RemoveCompanyFromUser(userID string) error {
	ctx := context.Background()
	_, err := gorm.G[map[string]interface{}](r.db).Table("users").Where("id = ? AND deleted_at IS NULL", userID).Updates(ctx, map[string]interface{}{
		"company_id": nil,
	})
	if err != nil {
		return util.ParseDatabaseError(err, "idx_users_")
	}
	return nil
}

// UserFilters is the read-side filter set used by GET /v1/users. Zero-value
// fields are skipped — every pointer or empty string means "no filter on
// this dimension." Mirrors LeadFilters.
type UserFilters struct {
	CompanyID  *string // admin-override; non-admin enforcement happens in the handler
	Role       *domain.Role
	Status     *domain.UserStatus
	Search     string
	Unassigned bool
}

// List returns users matching the actor's scope + filters, paginated and
// sorted. Soft-deleted rows are always excluded.
func (r *UserRepository) List(s auth.Scope, f UserFilters, p query.Page, srt query.Sort) ([]domain.User, int64, error) {
	ctx := context.Background()

	q := gorm.G[domain.User](r.db).Where("deleted_at IS NULL")
	if clause, args, applied := s.WhereClause("company_id"); applied {
		q = q.Where(clause, args...)
	}
	if f.CompanyID != nil {
		q = q.Where("company_id = ?", *f.CompanyID)
	}
	if f.Unassigned {
		q = q.Where("company_id IS NULL")
	}
	if f.Role != nil {
		q = q.Where("role = ?", string(*f.Role))
	}
	if f.Status != nil {
		q = q.Where("status = ?", string(*f.Status))
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("(name ILIKE ? OR email ILIKE ?)", like, like)
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

// Delete here is a hard delete. Callers wanting soft-delete should set
// DeletedAt/DeletedBy via Update instead.
func (r *UserRepository) Delete(user *domain.User) error {
	ctx := context.Background()
	_, err := gorm.G[domain.User](r.db).Where("id = ?", user.ID).Delete(ctx)
	if err != nil {
		return util.ParseDatabaseError(err, "idx_users_")
	}
	return nil
}
