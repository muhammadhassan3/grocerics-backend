package repository

import (
	"context"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/util"

	"gorm.io/gorm"
)

// ---------- carts ----------

type CartRepository struct{ db *gorm.DB }

func NewCartRepository(db *gorm.DB) *CartRepository { return &CartRepository{db: db} }

func (r *CartRepository) FindOrCreateByUser(userID string) (*domain.Cart, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.Cart](r.db).Where("user_id = ? AND deleted_at IS NULL", userID).First(ctx)
	if err == nil {
		return &data, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, util.ParseDatabaseError(err, "idx_carts_")
	}
	cart := &domain.Cart{UserID: userID}
	if err := gorm.G[domain.Cart](r.db).Create(ctx, cart); err != nil {
		return nil, util.ParseDatabaseError(err, "idx_carts_")
	}
	return cart, nil
}

// ---------- cart items ----------

type CartItemRepository struct{ db *gorm.DB }

func NewCartItemRepository(db *gorm.DB) *CartItemRepository { return &CartItemRepository{db: db} }

func (r *CartItemRepository) Upsert(cartID, variantID string, quantity int) (*domain.CartItem, error) {
	ctx := context.Background()
	data, err := gorm.G[domain.CartItem](r.db).
		Where("cart_id = ? AND variant_id = ? AND deleted_at IS NULL", cartID, variantID).First(ctx)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, util.ParseDatabaseError(err, "idx_cart_items_")
	}
	if err == gorm.ErrRecordNotFound {
		item := &domain.CartItem{CartID: cartID, VariantID: variantID, Quantity: quantity}
		if cErr := gorm.G[domain.CartItem](r.db).Create(ctx, item); cErr != nil {
			return nil, util.ParseDatabaseError(cErr, "idx_cart_items_")
		}
		return item, nil
	}
	data.Quantity = quantity
	if _, uErr := gorm.G[domain.CartItem](r.db).Where("id = ?", data.ID).Updates(ctx, data); uErr != nil {
		return nil, util.ParseDatabaseError(uErr, "idx_cart_items_")
	}
	return &data, nil
}

func (r *CartItemRepository) ListByCart(cartID string) ([]domain.CartItem, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.CartItem](r.db).
		Where("cart_id = ? AND deleted_at IS NULL", cartID).
		Order("created_at").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_cart_items_")
	}
	return items, nil
}

func (r *CartItemRepository) UpdateQuantity(itemID string, quantity int) error {
	err := r.db.WithContext(context.Background()).
		Model(&domain.CartItem{}).
		Where("id = ? AND deleted_at IS NULL", itemID).
		Update("quantity", quantity).Error
	if err != nil {
		return util.ParseDatabaseError(err, "idx_cart_items_")
	}
	return nil
}

func (r *CartItemRepository) Delete(itemID string) error {
	ctx := context.Background()
	if _, err := gorm.G[domain.CartItem](r.db).Where("id = ?", itemID).Delete(ctx); err != nil {
		return util.ParseDatabaseError(err, "idx_cart_items_")
	}
	return nil
}

// ---------- wishlist ----------

type WishlistRepository struct{ db *gorm.DB }

func NewWishlistRepository(db *gorm.DB) *WishlistRepository { return &WishlistRepository{db: db} }

// Add saves a variant to a user's wishlist (no-op if already present).
func (r *WishlistRepository) Add(userID, variantID string) error {
	ctx := context.Background()
	_, err := gorm.G[domain.Wishlist](r.db).
		Where("user_id = ? AND variant_id = ? AND deleted_at IS NULL", userID, variantID).First(ctx)
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return util.ParseDatabaseError(err, "idx_wishlists_")
	}
	if cErr := gorm.G[domain.Wishlist](r.db).Create(ctx, &domain.Wishlist{UserID: userID, VariantID: variantID}); cErr != nil {
		return util.ParseDatabaseError(cErr, "idx_wishlists_")
	}
	return nil
}

func (r *WishlistRepository) ListByUser(userID string) ([]domain.Wishlist, error) {
	ctx := context.Background()
	items, err := gorm.G[domain.Wishlist](r.db).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").Find(ctx)
	if err != nil {
		return nil, util.ParseDatabaseError(err, "idx_wishlists_")
	}
	return items, nil
}

func (r *WishlistRepository) Delete(userID, variantID string) error {
	ctx := context.Background()
	if _, err := gorm.G[domain.Wishlist](r.db).
		Where("user_id = ? AND variant_id = ?", userID, variantID).Delete(ctx); err != nil {
		return util.ParseDatabaseError(err, "idx_wishlists_")
	}
	return nil
}
