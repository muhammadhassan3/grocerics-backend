package v1

// In-package: seed() is unexported.

import "testing"

func i64p(v int64) *int64 { return &v }
func boolp(v bool) *bool  { return &v }

func TestSeedAbsentFallsBackToGetItem(t *testing.T) {
	s, err := ConfirmLinkRequest{}.seed()
	if err != nil {
		t.Fatalf("no seed fields is not an error: %v", err)
	}
	if s != nil {
		t.Errorf("want nil seed (GetItem fallback), got %+v", s)
	}
}

// price and availability describe one observation; half of it is not usable.
func TestSeedRequiresPriceAndAvailableTogether(t *testing.T) {
	if _, err := (ConfirmLinkRequest{PricePaise: i64p(3800)}).seed(); err == nil {
		t.Error("price without available should be rejected")
	}
	if _, err := (ConfirmLinkRequest{Available: boolp(true)}).seed(); err == nil {
		t.Error("available without price should be rejected")
	}
}

// Real case, verified against live QC: Blinkit item 10557 (Coke 2 ltr) reports
// price 0 / available false in Mumbai. Those are observations, not absent fields,
// so they must seed rather than fall back -- which is why the request uses pointers.
func TestSeedZeroPriceAndUnavailableAreRealValues(t *testing.T) {
	s, err := ConfirmLinkRequest{PricePaise: i64p(0), Available: boolp(false)}.seed()
	if err != nil {
		t.Fatalf("0/false is a real observation: %v", err)
	}
	if s == nil {
		t.Fatal("want a seed, got nil -- 0/false must not read as 'not sent'")
	}
	if s.PricePaise != 0 || s.Available {
		t.Errorf("got price=%d available=%v, want 0/false", s.PricePaise, s.Available)
	}
}

func TestSeedCarriesFields(t *testing.T) {
	inv := 14
	s, err := ConfirmLinkRequest{
		PricePaise: i64p(9000),
		MRPPaise:   i64p(9900),
		Available:  boolp(true),
		Inventory:  &inv,
		DeepLink:   "https://blinkit.com/prn/x/prid/10557",
	}.seed()
	if err != nil {
		t.Fatal(err)
	}
	if s.PricePaise != 9000 || s.MRPPaise != 9900 || !s.Available {
		t.Errorf("bad seed: %+v", s)
	}
	if s.Inventory == nil || *s.Inventory != 14 {
		t.Errorf("inventory not carried: %+v", s.Inventory)
	}
	if s.DeepLink == "" {
		t.Error("deeplink not carried")
	}
}

func TestSeedRejectsNegativeMoney(t *testing.T) {
	if _, err := (ConfirmLinkRequest{PricePaise: i64p(-1), Available: boolp(true)}).seed(); err == nil {
		t.Error("negative price should be rejected")
	}
	if _, err := (ConfirmLinkRequest{PricePaise: i64p(100), MRPPaise: i64p(-1), Available: boolp(true)}).seed(); err == nil {
		t.Error("negative mrp should be rejected")
	}
}
