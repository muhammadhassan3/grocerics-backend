package util

import (
	"fmt"
	"math"

	"grocerics-backend/internal/domain"
)

// FormatPaise renders paise as a ₹ string, e.g. 30000 -> "₹300.00".
func FormatPaise(paise int64) string {
	neg := ""
	if paise < 0 {
		neg = "-"
		paise = -paise
	}
	return fmt.Sprintf("%s₹%d.%02d", neg, paise/100, paise%100)
}

func RupeesToPaise(rupees float64) int64 {
	return int64(math.Round(rupees * 100))
}

func AveragePaise(amounts []int64) (int64, bool) {
	if len(amounts) == 0 {
		return 0, false
	}
	var sum int64
	for _, a := range amounts {
		sum += a
	}
	return int64(math.Round(float64(sum) / float64(len(amounts)))), true
}

func UnitPricePaise(pricePaise int64, volumeValue float64, unit domain.VolumeUnit) (perUnitPaise int64, label string, ok bool) {
	if volumeValue <= 0 {
		return 0, "", false
	}
	p := float64(pricePaise)
	switch unit {
	case domain.VolumeUnitGm:
		return int64(math.Round(p * 100 / volumeValue)), "100gm", true
	case domain.VolumeUnitKg: // 1 kg = 1000 g
		return int64(math.Round(p * 100 / (volumeValue * 1000))), "100gm", true
	case domain.VolumeUnitMl:
		return int64(math.Round(p * 100 / volumeValue)), "100ml", true
	case domain.VolumeUnitLtr: // 1 ltr = 1000 ml
		return int64(math.Round(p * 100 / (volumeValue * 1000))), "100ml", true
	case domain.VolumeUnitPcs:
		return int64(math.Round(p / volumeValue)), "pc", true
	}
	return 0, "", false
}
