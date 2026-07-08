package dto

import "grocerics-backend/internal/query"

// @Swagger:model UserDTO
// @Property id: Unique identifier for the user
// @Property name: Name of the user
// @Property email: Email address of the user
// @Property role: Role of the user
// @Property status: Status of the user
// @Property last_active_at: Timestamp of the user's last activity, RFC3339
// @Property location: Free-text location of the user
// @Property account_status: Account-level status, distinct from the user's active/disabled status
// @Property created_at: Creation timestamp, RFC3339
// @Description Data Transfer Object for a User entity.
type UserDTO struct {
	// Unique identifier for the user
	ID string `json:"id"`
	// Name of the user
	Name string `json:"name"`
	// Email address of the user
	Email string `json:"email"`
	// Role of the user
	Role string `json:"role" enums:"admin,client_manager,client"`
	// Status of the user
	Status string `json:"status" enums:"active,disabled"`
	// Timestamp of the user's last activity, RFC3339
	LastActiveAt string `json:"last_active_at"`
	// Free-text location of the user
	Location string `json:"location"`
	// Account-level status, distinct from the user's active/disabled status
	AccountStatus string `json:"account_status" enums:"active,suspended"`
	// Creation timestamp, RFC3339
	CreatedAt string `json:"created_at"`
}

// @Swagger:model UserListItemDTO
// @Property id: Unique identifier for the user
// @Property name: Name of the user
// @Property email: Email address of the user
// @Property role: Role of the user
// @Property status: Status of the user
// @Property created_at: Creation timestamp, RFC3339
// @Description Per-row shape returned by GET /v1/users.
type UserListItemDTO struct {
	// Unique identifier for the user
	ID string `json:"id"`
	// Name of the user
	Name string `json:"name"`
	// Email address of the user
	Email string `json:"email"`
	// Role of the user
	Role string `json:"role" enums:"admin,client_manager,client"`
	// Status of the user
	Status string `json:"status" enums:"active,disabled"`
	// Creation timestamp, RFC3339
	CreatedAt string `json:"created_at"`
}

// @Swagger:model UserListResponseDTO
// @Property items: Page of users matching the current filters
// @Property meta: Pagination metadata
// @Description Paginated envelope for GET /v1/users.
type UserListResponseDTO struct {
	// Page of users matching the current filters
	Items []UserListItemDTO `json:"items"`
	// Pagination metadata
	Meta query.Meta `json:"meta"`
}
