package dto

import "grocerics-backend/internal/query"

// @Swagger:model User
// @Description: Data Transfer Object for User entity
// @Property id: Unique identifier for the user
// @Property name: Name of the user
// @Property email: Email address of the user
// @Property role: Role of the user (e.g., "client", "client_manager", "admin")
// @Property status: Status of the user (e.g., "active", "disabled")
// @Property company_id: Optional company ID associated with the user
type UserDTO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// UserListItemDTO is the per-row shape returned by GET /v1/users.
type UserListItemDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// UserListResponseDTO is the envelope for GET /v1/users.
type UserListResponseDTO struct {
	Items []UserListItemDTO `json:"items"`
	Meta  query.Meta        `json:"meta"`
}
