package ownership_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"driftwatch/internal/ownership"
)

func TestHandler_ListAndAdd(t *testing.T) {
	reg := ownership.New()
	h := ownership.Handler(reg)

	body, _ := json.Marshal(ownership.Entry{Service: "api", Team: "platform"})
	req := httptest.NewRequest(http.MethodPost, "/ownership", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("POST: got %d, want 204", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/ownership", nil)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	var entries []ownership.Entry
	if err := json.NewDecoder(w2.Body).Decode(&entries); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(entries) != 1 || entries[0].Team != "platform" {
		t.Errorf("unexpected entries: %+v", entries)
	}
}

func TestHandler_GetAndDelete(t *testing.T) {
	reg := ownership.New()
	_ = reg.Set(ownership.Entry{Service: "svc-x", Team: "ops"})
	h := ownership.Handler(reg)

	req := httptest.NewRequest(http.MethodGet, "/ownership/svc-x", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET: got %d, want 200", w.Code)
	}

	del := httptest.NewRequest(http.MethodDelete, "/ownership/svc-x", nil)
	wd := httptest.NewRecorder()
	h.ServeHTTP(wd, del)
	if wd.Code != http.StatusNoContent {
		t.Fatalf("DELETE: got %d, want 204", wd.Code)
	}

	get2 := httptest.NewRequest(http.MethodGet, "/ownership/svc-x", nil)
	wg := httptest.NewRecorder()
	h.ServeHTTP(wg, get2)
	if wg.Code != http.StatusNotFound {
		t.Fatalf("GET after delete: got %d, want 404", wg.Code)
	}
}

func TestConcurrentSetAndGet(t *testing.T) {
	reg := ownership.New()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = reg.Set(ownership.Entry{Service: "svc", Team: "team"})
			reg.Get("svc")
		}(i)
	}
	wg.Wait()
}
