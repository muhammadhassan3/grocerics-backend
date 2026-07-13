package domain

type Category struct {
	BaseModel
	Name          string  `gorm:"not null" json:"name"`
	Slug          string  `gorm:"not null" json:"slug"`
	ImageURL      *string `json:"image_url,omitempty"`
	Description   *string `json:"description,omitempty"`
	IsTopCategory bool    `gorm:"not null;default:false" json:"is_top_category"`
	Status        Status  `gorm:"type:varchar;not null;default:'active'" json:"status"`
	DisplayOrder  int     `gorm:"not null;default:0" json:"display_order"`
	Timestamps
	SoftDelete
}

type Subcategory struct {
	BaseModel
	CategoryID       string  `gorm:"type:uuid;not null" json:"category_id"`
	Name             string  `gorm:"not null" json:"name"`
	Slug             *string `json:"slug,omitempty"`
	ImageURL         *string `json:"image_url,omitempty"`
	IsTopSubcategory bool    `gorm:"not null;default:false" json:"is_top_subcategory"`
	Status           Status  `gorm:"type:varchar;not null;default:'active'" json:"status"`
	DisplayOrder     int     `gorm:"not null;default:0" json:"display_order"`
	Timestamps
	SoftDelete
}

type Brand struct {
	BaseModel
	Name       string  `gorm:"not null" json:"name"`
	Slug       *string `json:"slug,omitempty"`
	ImageURL   *string `json:"image_url,omitempty"`
	IsTopBrand bool    `gorm:"not null;default:false" json:"is_top_brand"`
	Status     Status  `gorm:"type:varchar;not null;default:'active'" json:"status"`
	Timestamps
	SoftDelete
}

type Product struct {
	BaseModel
	CategoryID    string  `gorm:"type:uuid;not null" json:"category_id"`
	SubcategoryID *string `gorm:"type:uuid" json:"subcategory_id,omitempty"`
	BrandID       *string `gorm:"type:uuid" json:"brand_id,omitempty"`
	Name          string  `gorm:"not null" json:"name"`
	Description   *string `json:"description,omitempty"`
	ImageURL      *string `json:"image_url,omitempty"`
	IsTopItem     bool    `gorm:"not null;default:false" json:"is_top_item"`
	Status        Status  `gorm:"type:varchar;not null;default:'active'" json:"status"`
	Timestamps
	SoftDelete
}

type ProductImage struct {
	BaseModel
	ProductID    string `gorm:"type:uuid;not null" json:"product_id"`
	ImageURL     string `gorm:"not null" json:"image_url"`
	DisplayOrder int    `gorm:"not null;default:0" json:"display_order"`
	Timestamps
	SoftDelete
}

type ProductVariant struct {
	BaseModel
	ProductID       string     `gorm:"type:uuid;not null" json:"product_id"`
	CustomVariantID *string    `json:"custom_variant_id,omitempty"`
	VolumeValue     float64    `gorm:"type:numeric;not null" json:"volume_value"`
	VolumeUnit      VolumeUnit `gorm:"type:varchar;not null" json:"volume_unit"`
	DisplayOrder    int        `gorm:"not null;default:0" json:"display_order"`
	Timestamps
	SoftDelete
}
