package dto

// @Swagger:model ProductVariantUnit
// @Property value: Numeric quantity of the unit
// @Property unit: Unit of measurement
// @Description Volume/weight of a product variant, e.g. "500 gm".
type ProductVariantUnit struct {
	// Numeric quantity of the unit
	Value int `json:"value"`
	// Unit of measurement
	Unit string `json:"unit" enums:"kg,gm,ltr,ml,pcs"`
}

// @Swagger:model ProductVariantItem
// @Property product_id: Parent product's identifier
// @Property product_variant_id: Unique identifier for this variant
// @Property variant_custom_id: Human-readable/SKU-style identifier for this variant
// @Property product_volume: Volume or weight of this variant
// @Description A single sellable variant of a product (e.g. a specific pack size).
// Price and stock are per-platform and per-city, so they are NOT on the variant:
// read them from GET /v1/inventory-management/variants/{variant_id}/prices.
type ProductVariantItem struct {
	// Parent product's identifier
	ProductID string `json:"product_id"`
	// Unique identifier for this variant
	ProductVariantID string `json:"product_variant_id"`
	// Human-readable/SKU-style identifier for this variant
	VariantCustomID string `json:"variant_custom_id"`
	// Volume or weight of this variant
	ProductVolume ProductVariantUnit `json:"product_volume"`
}

type ProductVariantItems struct {
	// Page of product variants
	Variants []ProductVariantItem `json:"variants"`
}
