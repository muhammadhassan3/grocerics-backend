package repository

import (
	"context"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

// ---------- cities ----------

type CityRepository struct{ db *gorm.DB }

func NewCityRepository(db *gorm.DB) *CityRepository { return &CityRepository{db: db} }

func (r *CityRepository) ListEnabled() ([]domain.City, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.City](r.db).
		Where("enabled AND deleted_at IS NULL").
		Order("display_order, name").
		Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_cities_")
	}
	return items, nil
}

func (r *CityRepository) FindByID(id string) (*domain.City, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.City](r.db).Where("id = ? AND deleted_at IS NULL", id).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_cities_")
	}
	return &data, nil
}

// ---------- pincodes ----------

type PincodeRepository struct{ db *gorm.DB }

func NewPincodeRepository(db *gorm.DB) *PincodeRepository { return &PincodeRepository{db: db} }

func (r *PincodeRepository) FindByPincode(pincode string) (*domain.Pincode, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.Pincode](r.db).Where("pincode = ? AND deleted_at IS NULL", pincode).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_pincodes_")
	}
	return &data, nil
}

// ---------- user addresses ----------

type AddressRepository struct{ db *gorm.DB }

func NewAddressRepository(db *gorm.DB) *AddressRepository { return &AddressRepository{db: db} }

func (r *AddressRepository) Create(a *domain.UserAddress) (*domain.UserAddress, error) {
	ctx := context.Background()
	if err := gorm.G[domain.UserAddress](r.db).Create(ctx, a); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_user_addresses_")
	}
	return a, nil
}

func (r *AddressRepository) ListByUser(userID string) ([]domain.UserAddress, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.UserAddress](r.db).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("is_default DESC, created_at DESC").
		Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_user_addresses_")
	}
	return items, nil
}

func (r *AddressRepository) FindByID(id string) (*domain.UserAddress, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.UserAddress](r.db).Where("id = ? AND deleted_at IS NULL", id).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, util.ParseDatabaseError(err, "idx_user_addresses_")
	}
	return &data, nil
}

func (r *AddressRepository) Update(a *domain.UserAddress) (*domain.UserAddress, error) {
	ctx := context.Background()
	if _, err := gorm.G[domain.UserAddress](r.db).Where("id = ? AND deleted_at IS NULL", a.ID).Updates(ctx, *a); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_user_addresses_")
	}
	return a, nil
}

func (r *AddressRepository) UnsetDefaults(userID string) error {
	err := r.db.WithContext(context.Background()).
		Model(&domain.UserAddress{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Update("is_default", false).Error
	if err != nil {
		return util.ParseDatabaseError(err, "idx_user_addresses_")
	}
	return nil
}
