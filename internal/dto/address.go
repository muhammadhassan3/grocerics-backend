package dto

// @Swagger:model AddAddressRequest
// @Property address_line_1: Primary address line, e.g. house/flat number and street
// @Property address_line_2: Secondary address line, e.g. landmark or apartment details
// @Property pincode: Postal/ZIP code of the address
// @Description Request payload for adding a new delivery address.
type AddAddressRequest struct {
	// Primary address line, e.g. house/flat number and street
	AddressLine1 string `json:"address_line_1" binding:"required"`
	// Secondary address line, e.g. landmark or apartment details
	AddressLine2 string `json:"address_line_2"`
	// Postal/ZIP code of the address
	Pincode string `json:"pincode" binding:"required"`
}

// @Swagger:model AddressResponse
// @Property address_id: Unique identifier for the address
// @Property address_line_1: Primary address line, e.g. house/flat number and street
// @Property address_line_2: Secondary address line, e.g. landmark or apartment details
// @Property pincode: Postal/ZIP code of the address
// @Property created_at: Timestamp when the address was added, RFC3339
// @Description Response payload containing details of a delivery address.
type AddressResponse struct {
	AddressID    string `json:"address_id"`
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	Pincode      string `json:"pincode"`
	CreatedAt    string `json:"created_at"`
}
