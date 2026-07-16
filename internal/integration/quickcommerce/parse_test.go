package quickcommerce

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestFlexFloat(t *testing.T) {
	cases := map[string]float64{`72`: 72, `"72"`: 72, `84.0`: 84, `"38.5"`: 38.5, `null`: 0, `""`: 0}
	for in, want := range cases {
		var f flexFloat
		if err := json.Unmarshal([]byte(in), &f); err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if float64(f) != want {
			t.Errorf("flexFloat(%s) = %v, want %v", in, float64(f), want)
		}
	}
}

func TestFlexInventory(t *testing.T) {

	var fi flexInventory
	mustJSON(t, `18`, &fi)
	if !fi.present || fi.Count == nil || *fi.Count != 18 || !fi.Available {
		t.Errorf("int: %+v", fi)
	}

	fi = flexInventory{}
	mustJSON(t, `{"inStock":true,"lowStockText":"Only 2 left"}`, &fi)
	if !fi.present || fi.Count != nil || !fi.Available || fi.Label != "Only 2 left" {
		t.Errorf("object: %+v", fi)
	}

	fi = flexInventory{}
	mustJSON(t, `{"inStock":false}`, &fi)
	if fi.Available {
		t.Errorf("object false should be unavailable: %+v", fi)
	}

	fi = flexInventory{}
	mustJSON(t, `0`, &fi)
	if !fi.present || fi.Count == nil || *fi.Count != 0 || fi.Available {
		t.Errorf("zero: %+v", fi)
	}

	fi = flexInventory{}
	mustJSON(t, `null`, &fi)
	if fi.present {
		t.Errorf("null should be absent: %+v", fi)
	}
}

// The QC API returns `available` as a bool on some platforms and a string on
// others — a real GetItem response with "available":"true" used to crash the
// decode and fail the whole link confirmation.
func TestFlexBool(t *testing.T) {
	cases := map[string]bool{
		`true`: true, `false`: false,
		`"true"`: true, `"false"`: false,
		`"True"`: true, `"FALSE"`: false,
		`"1"`: true, `"0"`: false,
		`1`: true, `0`: false,
		`"yes"`: true, `"no"`: false,
		`"in stock"`: true, `"out of stock"`: false,
		`null`: false, `""`: false,
		`"sold out"`: false, // unknown/negative -> false, never claim phantom stock
	}
	for in, want := range cases {
		var b flexBool
		if err := json.Unmarshal([]byte(in), &b); err != nil {
			t.Fatalf("flexBool(%s): %v", in, err)
		}
		if bool(b) != want {
			t.Errorf("flexBool(%s) = %v, want %v", in, bool(b), want)
		}
	}
}

// A string `available` must survive all the way through to a parsed item.
func TestRawItemAvailableAsString(t *testing.T) {
	var it rawItem
	mustJSON(t, `{"item_id":"283","name":"Coke","quantity":"750 ml","available":"true","price":"38","mrp":40}`, &it)
	if !bool(it.Available) {
		t.Fatalf("string available should decode true, got %v", it.Available)
	}
}

func TestParseMultipack(t *testing.T) {
	cases := map[string]int{
		"750 ml": 1, "750 ml x 12": 12, "6 x 300 ml": 6, "2 x 750 ml": 2,
		"1 pc (750 ml)": 1, "300 ml X 6": 6, "1 Combo": 1,
	}
	for in, want := range cases {
		if got := parseMultipack(in); got != want {
			t.Errorf("parseMultipack(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestResolveStock(t *testing.T) {

	p := rawProduct{Available: true, Inventory: parseInv(t, `{"inStock":true}`)}.toProduct()
	if !p.Available || p.Inventory != nil {
		t.Errorf("object stock: available=%v inv=%v", p.Available, p.Inventory)
	}

	p = rawProduct{Available: true, Inventory: parseInv(t, `0`)}.toProduct()
	if p.Available {
		t.Errorf("zero inventory should force unavailable")
	}
}

func mustJSON(t *testing.T, s string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(s), v); err != nil {
		t.Fatalf("unmarshal %s: %v", s, err)
	}
}

func parseInv(t *testing.T, s string) flexInventory {
	t.Helper()
	var fi flexInventory
	mustJSON(t, s, &fi)
	return fi
}

func TestLiveGroupSearch(t *testing.T) {
	key := os.Getenv("QC_API_KEY")
	if key == "" {
		t.Skip("QC_API_KEY not set; skipping live QuickCommerce test")
	}
	c := New(Config{APIKey: key})
	loc := Location{Lat: 28.6980, Lon: 77.1490, Pincode: "110035"}
	res, err := c.GroupSearch(context.Background(), "coca cola", loc, []string{"BlinkIt", "Zepto", "Swiggy"})
	if err != nil {
		t.Fatalf("group search: %v", err)
	}
	if len(res.ByPlatform) == 0 {
		t.Fatal("expected at least one platform in results")
	}
	for plat, prods := range res.ByPlatform {
		if len(prods) == 0 {
			continue
		}
		p := prods[0]
		if p.ID == "" || p.PricePaise <= 0 {
			t.Errorf("%s: bad parse id=%q price=%d (string-price platforms must parse)", plat, p.ID, p.PricePaise)
		}
		t.Logf("%s: %d results, first=%q ₹%d.%02d inv=%v pack=%d", plat, len(prods), p.Name, p.PricePaise/100, p.PricePaise%100, p.Inventory, p.Multipack)
	}
}
