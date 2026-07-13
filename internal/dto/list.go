package dto

import "grocerics-backend/internal/query"

type ProductCardListDTO struct {
	Items []ProductCardDTO `json:"items"`
	Meta  query.Meta       `json:"meta"`
}
