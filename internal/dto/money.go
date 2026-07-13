package dto

import "grocerics-backend/internal/util"

// MoneyDTO carries money as both integer paise
// and a ₹ string (for display)
type MoneyDTO struct {
	Paise   int64  `json:"paise"`
	Display string `json:"display"`
}

func Money(paise int64) MoneyDTO {
	return MoneyDTO{Paise: paise, Display: util.FormatPaise(paise)}
}

func MoneyPtr(paise *int64) *MoneyDTO {
	if paise == nil {
		return nil
	}
	m := Money(*paise)
	return &m
}
