package leads

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"maps"
	"strconv"
	"strings"
	"time"
)

// Rejection identifies a source line and safe validation messages for correction.
type Rejection struct {
	Line   int               `json:"line"`
	Errors map[string]string `json:"errors"`
}

// ImportResult accounts for every accepted and rejected CSV record.
type ImportResult struct {
	Imported int         `json:"imported"`
	Rejected []Rejection `json:"rejected"`
}

// Import processes each CSV record independently. Fatal I/O and database errors stop the import.
func (store *Store) Import(ctx context.Context, input io.Reader) (ImportResult, error) {
	result := ImportResult{Rejected: make([]Rejection, 0)}
	reader := csv.NewReader(input)
	header, err := reader.Read()
	if err != nil {
		return result, errors.New("CSV header is missing or malformed")
	}
	columns, err := csvColumns(header)
	if err != nil {
		return result, err
	}
	for {
		row, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			return result, nil
		}
		if err = ctx.Err(); err != nil {
			return result, err
		}
		if readErr != nil {
			parseErr, ok := errors.AsType[*csv.ParseError](readErr)
			if !ok {
				return result, errors.New("could not read CSV")
			}
			result.Rejected = append(
				result.Rejected,
				Rejection{
					Line:   parseErr.StartLine,
					Errors: map[string]string{"row": "malformed CSV or wrong column count"},
				},
			)
			continue
		}
		line, _ := reader.FieldPos(0)
		lead, fields := parseRow(row, columns)
		if err = store.saveImported(ctx, lead, fields); err != nil {
			return result, err
		}
		if len(fields) > 0 {
			result.Rejected = append(result.Rejected, Rejection{Line: line, Errors: fields})
		} else {
			result.Imported++
		}
	}
}

func csvColumns(header []string) (map[string]int, error) {
	columns := make(map[string]int, len(header))
	for index, name := range header {
		name = strings.TrimSpace(strings.TrimPrefix(name, "\ufeff"))
		if _, exists := columns[name]; exists {
			return nil, errors.New("CSV contains duplicate column headers")
		}
		columns[name] = index
	}
	for _, name := range []string{"first_name", "last_name", "email", "finance_amount", "salary", "state", "vehicle_make", "vehicle_model", "vehicle_year"} {
		if _, ok := columns[name]; !ok {
			return nil, fmt.Errorf("CSV is missing column %s", name)
		}
	}
	return columns, nil
}

func (store *Store) saveImported(ctx context.Context, lead Lead, fields map[string]string) error {
	if len(fields) != 0 {
		return nil
	}
	_, err := store.Add(ctx, lead)
	if errors.Is(err, ErrDuplicateEmail) {
		fields["email"] = "already exists"
		return nil
	}
	return err
}

func parseRow(row []string, columns map[string]int) (Lead, map[string]string) {
	value := func(name string) string { return strings.TrimSpace(row[columns[name]]) }
	parseErrors := make(map[string]string)
	finance := csvAmount(value("finance_amount"), "finance_amount", parseErrors)
	lead := Lead{
		FirstName:     value("first_name"),
		LastName:      value("last_name"),
		Email:         value("email"),
		State:         value("state"),
		FinanceAmount: finance,
	}
	if value("salary") != "" {
		salary := csvAmount(value("salary"), "salary", parseErrors)
		lead.Salary = &salary
	}
	vehicle := &Vehicle{Make: value("vehicle_make"), Model: value("vehicle_model")}
	if value("vehicle_year") != "" {
		year, err := strconv.Atoi(value("vehicle_year"))
		if err != nil {
			parseErrors["vehicle.year"] = "must be a whole year"
		} else {
			vehicle.Year = &year
		}
	}
	lead.Vehicle = vehicle
	lead.Normalize()
	fields := lead.Validate(time.Now())
	maps.Copy(fields, parseErrors)
	return lead, fields
}

func csvAmount(raw, field string, fields map[string]string) float64 {
	// Partner exports may include dollar symbols and thousands separators.
	clean := strings.ReplaceAll(strings.TrimPrefix(raw, "$"), ",", "")
	amount, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		fields[field] = "must be a number in dollars"
	}
	return amount
}
