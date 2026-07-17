package v1

// In-package: toVariantItemDTO is unexported.

import (
	"encoding/json"
	"testing"

	"grocerics-backend/internal/domain"
)

// Pack sizes are not whole numbers. Blinkit sells Coca-Cola at 2.25 ltr; 1.5 ltr and
// 0.5 kg are equally real. domain.ProductVariant.VolumeValue is float64 and the column
// is numeric, so the DTO must not narrow them.

func TestCreateVariantAcceptsDecimalVolume(t *testing.T) {
	var req CreateVariantRequest
	body := []byte(`{"product_id":"p1","volume":{"value":2.25,"unit":"ltr"}}`)
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("2.25 ltr must bind, got: %v", err)
	}
	if got := float64(req.Volume.Value); got != 2.25 {
		t.Errorf("value = %v, want 2.25", got)
	}
}

func TestUpdateVariantAcceptsDecimalVolume(t *testing.T) {
	var req UpdateVariantRequest
	body := []byte(`{"product_variant_id":"v1","volume":{"value":1.5,"unit":"ltr"}}`)
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("1.5 ltr must bind, got: %v", err)
	}
	if got := float64(req.Volume.Value); got != 1.5 {
		t.Errorf("value = %v, want 1.5", got)
	}
}

// The read path truncated: a 1.5 ltr variant in the DB came back as 1 over the API.
func TestVariantDTOKeepsDecimalVolume(t *testing.T) {
	got := toVariantItemDTO(domain.ProductVariant{
		VolumeValue: 1.5,
		VolumeUnit:  domain.VolumeUnitLtr,
	})
	if v := float64(got.ProductVolume.Value); v != 1.5 {
		t.Errorf("value = %v, want 1.5 — a 1.5 ltr variant must not read back as 1", v)
	}
	if got.ProductVolume.Unit != "ltr" {
		t.Errorf("unit = %q, want ltr", got.ProductVolume.Unit)
	}
}

// Whole numbers must stay clean, not become 2.0000000001 or similar.
func TestWholeVolumesUnaffected(t *testing.T) {
	got := toVariantItemDTO(domain.ProductVariant{
		VolumeValue: 2,
		VolumeUnit:  domain.VolumeUnitLtr,
	})
	if v := float64(got.ProductVolume.Value); v != 2 {
		t.Errorf("value = %v, want 2", v)
	}
	b, err := json.Marshal(got.ProductVolume)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"value":2,"unit":"ltr"}` {
		t.Errorf("whole number must serialise as 2, got %s", b)
	}
}
