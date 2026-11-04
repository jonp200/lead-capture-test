package leads_test

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"leadcapture/internal/leads"
)

func openStore(t *testing.T) *leads.Store {
	t.Helper()
	store, err := leads.Open(t.Context(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal("open store:", err)
	}
	t.Cleanup(func() {
		if closeErr := store.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	})
	return store
}

func call(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, req)
	return recorder
}

func TestAPI(t *testing.T) {
	store := openStore(t)
	handler := leads.Handler(store, fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("frontend")}}, testLogger())
	response := call(handler, http.MethodGet, "/api/leads", "")
	if response.Code != http.StatusOK || response.Body.String() != "[]" {
		t.Fatal("empty list must be []")
	}
	input := `{"first_name":"Zoë","last_name":"Martin","email":"ZOE@example.com","finance_amount":38500.5,"salary":null,"state":"qld","vehicle":{"make":"Kia","model":"EV6","year":null}}`
	response = call(handler, http.MethodPost, "/api/leads", input)
	if response.Code != http.StatusCreated {
		t.Fatalf("create status: %d", response.Code)
	}
	var saved leads.Lead
	if err := json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.ID == 0 || saved.CreatedAt == "" || saved.State != "QLD" || saved.Vehicle == nil ||
		saved.Vehicle.Year != nil {
		t.Fatal("incorrect saved metadata or optional fields")
	}
	response = call(handler, http.MethodGet, "/api/leads", "")
	var list []leads.Lead
	if err := json.Unmarshal(response.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != saved.ID {
		t.Fatal("saved ID did not round trip")
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("PII must not be cached")
	}
	response = call(
		handler,
		http.MethodPost,
		"/api/leads",
		strings.ReplaceAll(input, "ZOE@example.com", "zoe@example.com"),
	)
	if response.Code != http.StatusConflict {
		t.Fatal("expected case-insensitive duplicate conflict")
	}
	response = call(handler, http.MethodPost, "/api/leads", `{}`)
	if response.Code != http.StatusUnprocessableEntity ||
		!strings.Contains(response.Body.String(), `"finance_amount"`) {
		t.Fatal("missing field errors")
	}
	for index, body := range []string{`{`, `null`, `[]`, `{} {}`, `{"finance_amount":"abc"}`, `{"unknown":1}`, `{"vehicle":{"year":2020.5}}`} {
		t.Run(fmt.Sprintf("malformed-%d", index), func(t *testing.T) {
			if result := call(handler, http.MethodPost, "/api/leads", body); result.Code != http.StatusBadRequest {
				t.Errorf("status: %d", result.Code)
			}
		})
	}
	if result := call(
		handler,
		http.MethodGet,
		"/",
		"",
	); result.Code != http.StatusOK ||
		result.Body.String() != "frontend" {
		t.Fatal("frontend not served")
	}
	if result := call(handler, http.MethodGet, "/api/missing", ""); result.Code != http.StatusNotFound {
		t.Fatal("unknown API should return 404")
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestPersistence(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "persistent.db")
	store, err := leads.Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Add(t.Context(), validLead())
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := leads.Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	items, err := reopened.List(t.Context())
	if err != nil || len(items) != 1 || items[0].ID == 0 {
		t.Fatal("lead not persisted")
	}
}
