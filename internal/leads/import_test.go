package leads_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"leadcapture/internal/leads"
)

const csvHeader = "first_name,last_name,email,finance_amount,salary,state,vehicle_make,vehicle_model,vehicle_year\n"

func TestImportSample(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	file, err := os.Open("../../leads.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	result, err := store.Import(t.Context(), file)
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 36 || len(result.Rejected) != 4 {
		t.Fatalf("imported %d; rejected %d", result.Imported, len(result.Rejected))
	}
	for index, expected := range []struct {
		line  int
		field string
	}{{25, "state"}, {29, "email"}, {33, "vehicle.make"}, {37, "email"}} {
		if result.Rejected[index].Line != expected.line || result.Rejected[index].Errors[expected.field] == "" {
			t.Fatalf("incorrect rejection: %v", result.Rejected[index])
		}
	}
	items, err := store.List(t.Context())
	if err != nil || len(items) != 36 {
		t.Fatal("incorrect stored count")
	}
}

func TestImportInvalidRows(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	input := csvHeader + "Bad,Amount,bad@example.com,nope,,VIC,,,\n" +
		"Bad,Salary,salary@example.com,100,NaN,VIC,,,\n" +
		"Bad,Year,year@example.com,100,,VIC,Kia,,2020.5\n" +
		"too,few,columns\n" +
		"Good,Lead,good@example.com,100,,VIC,,,\n"
	result, err := store.Import(t.Context(), strings.NewReader(input))
	if err != nil || result.Imported != 1 || len(result.Rejected) != 4 {
		t.Fatalf("incorrect row handling: %v, %v", result, err)
	}
	if strings.Contains(fmt.Sprint(result), "@example.com") {
		t.Fatal("report leaked PII")
	}
}

func TestImportHeadersAndMultiline(t *testing.T) {
	t.Parallel()
	for _, input := range []string{"", "email\n", csvHeader + "", strings.Replace(csvHeader, "last_name", "first_name", 1)} {
		store := openStore(t)
		_, err := store.Import(t.Context(), strings.NewReader(input))
		if input != csvHeader && err == nil {
			t.Fatal("expected header error")
		}
	}
	store := openStore(t)
	input := csvHeader + "\"Multi\nLine\",Lead,first@example.com,100,,VIC,,,\n" + "Bad,Lead,,100,,VIC,,,\n"
	result, err := store.Import(t.Context(), strings.NewReader(input))
	if err != nil || result.Imported != 1 || len(result.Rejected) != 1 || result.Rejected[0].Line != 4 {
		t.Fatal("incorrect physical line number")
	}
}

func TestImportUpload(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	handler := leads.Handler(store, nil, testLogger())
	response := call(handler, "POST", "/api/leads/import", csvHeader+"Good,Lead,good@example.com,100,,VIC,,,\n")
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"imported":1`) {
		t.Fatal("upload failed")
	}
}

type failedReader struct{}

func (failedReader) Read(_ []byte) (int, error) { return 0, errors.New("read failed") }

func TestImportReadFailureKeepsProgress(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	input := io.MultiReader(strings.NewReader(csvHeader+"Good,Lead,good@example.com,100,,VIC,,,\n"), failedReader{})
	result, err := store.Import(t.Context(), input)
	if err == nil || result.Imported != 1 {
		t.Fatal("read failure must return committed progress")
	}
}

func TestUploadLimit(t *testing.T) {
	t.Parallel()
	store := openStore(t)
	handler := leads.Handler(store, nil, testLogger())
	const limit = 10 << 20
	response := call(handler, "POST", "/api/leads/import", csvHeader+strings.Repeat("x", limit))
	if response.Code != 400 || !strings.Contains(response.Body.String(), `"result"`) {
		t.Fatal("oversized upload must return a safe error and progress")
	}
}
