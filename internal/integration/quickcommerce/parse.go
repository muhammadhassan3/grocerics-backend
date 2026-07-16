package quickcommerce

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

// The QuickCommerce API is inconsistent across platforms: prices arrive as a
// number (BlinkIt 17), a string ("72"), or a float (84.0); stock arrives as an
// integer, an object {inStock, lowStockText}, a bare bool, or null. The helpers
// below normalise all of those at the decode boundary so the rest of the code
// sees clean Go types.

type flexFloat float64

func (f *flexFloat) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*f = 0
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			*f = 0
			return nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		*f = flexFloat(v)
		return nil
	}
	var v float64
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*f = flexFloat(v)
	return nil
}

// flexBool decodes "available", which arrives as a real bool, a string
// ("true" / "1" / "in stock"), or a number (0/1) depending on the platform.
// Unrecognised values decode to false: claiming an item is in stock when we
// can't confirm it is the worse failure (the user goes to buy something gone).
type flexBool bool

func (f *flexBool) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*f = false
		return nil
	}
	switch b[0] {
	case 't', 'f':
		var v bool
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*f = flexBool(v)
	case '"':
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*f = flexBool(parseBoolish(s))
	default: // number: 0 = out of stock, anything else = in stock
		var n float64
		if err := json.Unmarshal(b, &n); err != nil {
			return err
		}
		*f = flexBool(n != 0)
	}
	return nil
}

func parseBoolish(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return false
	}
	if v, err := strconv.ParseBool(s); err == nil { // true/false/1/0/t/f
		return v
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n != 0
	}
	switch s {
	case "yes", "y", "in stock", "instock", "in_stock", "available":
		return true
	}
	return false
}

type flexInventory struct {
	present   bool
	Available bool
	Count     *int
	Label     string
}

func (fi *flexInventory) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	fi.present = true
	switch b[0] {
	case '{': // object: {inStock, lowStockText}
		var o struct {
			InStock      *bool  `json:"inStock"`
			LowStockText string `json:"lowStockText"`
		}
		if err := json.Unmarshal(b, &o); err != nil {
			return err
		}
		fi.Available = o.InStock == nil || *o.InStock
		fi.Label = o.LowStockText
	case 't', 'f': // bare bool
		return json.Unmarshal(b, &fi.Available)
	case '"': // string — maybe a numeric string
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			fi.Count, fi.Available = &n, n > 0
		} else {
			fi.present = false
		}
	default: // number (int or float)
		var fl float64
		if err := json.Unmarshal(b, &fl); err != nil {
			return err
		}
		n := int(fl)
		fi.Count, fi.Available = &n, n > 0
	}
	return nil
}

var (
	multipackBefore = regexp.MustCompile(`(?i)(\d+)\s*x`)
	multipackAfter  = regexp.MustCompile(`(?i)x\s*(\d+)\b`)
)

func parseMultipack(qty string) int {
	if m := multipackBefore.FindStringSubmatch(qty); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			return n
		}
	}
	if m := multipackAfter.FindStringSubmatch(qty); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			return n
		}
	}
	return 1
}

var firstIntRe = regexp.MustCompile(`\d+`)

func firstInt(s string) int {
	if m := firstIntRe.FindString(s); m != "" {
		n, _ := strconv.Atoi(m)
		return n
	}
	return 0
}
