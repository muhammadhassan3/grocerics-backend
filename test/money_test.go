package test

import (
	"testing"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/util"
)

func TestFormatPaise(t *testing.T) {
	cases := map[int64]string{
		30000: "₹300.00",
		5:     "₹0.05",
		99:    "₹0.99",
		100:   "₹1.00",
		-2550: "-₹25.50",
	}
	for in, want := range cases {
		if got := util.FormatPaise(in); got != want {
			t.Errorf("FormatPaise(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestAveragePaise(t *testing.T) {
	if _, ok := util.AveragePaise(nil); ok {
		t.Error("empty slice should return ok=false")
	}
	// 30000, 30001, 30002 -> mean 30001
	if got, ok := util.AveragePaise([]int64{30000, 30001, 30002}); !ok || got != 30001 {
		t.Errorf("got %d ok=%v, want 30001 true", got, ok)
	}
	// rounding: 100, 101 -> 100.5 -> 101
	if got, _ := util.AveragePaise([]int64{100, 101}); got != 101 {
		t.Errorf("got %d, want 101", got)
	}
}

func TestUnitPricePaise(t *testing.T) {
	// ₹360 for 500gm -> ₹72 / 100gm = 7200 paise
	if got, label, ok := util.UnitPricePaise(36000, 500, domain.VolumeUnitGm); !ok || got != 7200 || label != "100gm" {
		t.Errorf("gm: got %d %q ok=%v, want 7200 100gm true", got, label, ok)
	}
	// ₹200 for 1kg -> ₹20 / 100gm = 2000 paise
	if got, label, _ := util.UnitPricePaise(20000, 1, domain.VolumeUnitKg); got != 2000 || label != "100gm" {
		t.Errorf("kg: got %d %q, want 2000 100gm", got, label)
	}
	// ₹50 for 2 pcs -> ₹25 / pc = 2500 paise
	if got, label, _ := util.UnitPricePaise(5000, 2, domain.VolumeUnitPcs); got != 2500 || label != "pc" {
		t.Errorf("pcs: got %d %q, want 2500 pc", got, label)
	}
	// zero volume -> not ok
	if _, _, ok := util.UnitPricePaise(5000, 0, domain.VolumeUnitGm); ok {
		t.Error("zero volume should return ok=false")
	}
}
