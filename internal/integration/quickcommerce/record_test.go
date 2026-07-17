package quickcommerce

// In-package: doGet and emit are unexported, same reason parse_test.go lives here.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func serving(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRecordsRawResponse(t *testing.T) {
	srv := serving(t, 200, `{"credits_remaining":7,"data":{"eta":"10 Mins","open":true}}`)
	var got []RawCall
	c := NewHTTPClient(Config{APIKey: "k", BaseURL: srv.URL, Record: func(rc RawCall) { got = append(got, rc) }})

	if _, err := c.ETA(context.Background(), "BlinkIt", Location{Lat: 28.698, Lon: 77.149, Pincode: "110035"}); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 recorded call, got %d", len(got))
	}
	rc := got[0]
	if rc.Endpoint != "/eta" {
		t.Errorf("endpoint = %q, want /eta", rc.Endpoint)
	}
	if rc.StatusCode != 200 {
		t.Errorf("status = %d, want 200", rc.StatusCode)
	}
	if !strings.Contains(string(rc.Body), "10 Mins") {
		t.Errorf("body not captured verbatim: %q", rc.Body)
	}
	if rc.Err != "" {
		t.Errorf("Err = %q, want empty", rc.Err)
	}
	if rc.Params.Get("platform") != "BlinkIt" || rc.Params.Get("pincode") != "110035" {
		t.Errorf("params not captured: %v", rc.Params)
	}
}

// The ML person wants the real distribution, so failures are data too.
func TestRecordsFailures(t *testing.T) {
	srv := serving(t, 500, `{"error":"boom"}`)
	var got []RawCall
	c := NewHTTPClient(Config{APIKey: "k", BaseURL: srv.URL, Record: func(rc RawCall) { got = append(got, rc) }})

	if _, err := c.Credits(context.Background()); err == nil {
		t.Fatal("want an error on 500")
	}
	if len(got) != 1 {
		t.Fatalf("want the failure recorded, got %d calls", len(got))
	}
	if got[0].StatusCode != 500 {
		t.Errorf("status = %d, want 500", got[0].StatusCode)
	}
	if got[0].Err == "" {
		t.Error("want Err populated on a 500")
	}
	if !strings.Contains(string(got[0].Body), "boom") {
		t.Errorf("error body not captured: %q", got[0].Body)
	}
}

// The sink is an observer. If it explodes, the admin's search still succeeds.
func TestRecorderPanicDoesNotBreakCall(t *testing.T) {
	srv := serving(t, 200, `{"credits_remaining":7}`)
	c := NewHTTPClient(Config{APIKey: "k", BaseURL: srv.URL, Record: func(RawCall) { panic("ml sink exploded") }})

	res, err := c.Credits(context.Background())
	if err != nil {
		t.Fatalf("recorder panic must not fail the QC call: %v", err)
	}
	if res.Remaining != 7 {
		t.Errorf("remaining = %d, want 7", res.Remaining)
	}
}

func TestNilRecorderIsFine(t *testing.T) {
	srv := serving(t, 200, `{"credits_remaining":7}`)
	c := NewHTTPClient(Config{APIKey: "k", BaseURL: srv.URL})
	if _, err := c.Credits(context.Background()); err != nil {
		t.Fatal(err)
	}
}
