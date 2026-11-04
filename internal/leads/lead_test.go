package leads_test

import (
	"math"
	"testing"
	"time"

	"leadcapture/internal/leads"
)

func validLead() leads.Lead {
	return leads.Lead{
		FirstName:     "Zoë",
		LastName:      "Martin",
		Email:         "zoe@example.com",
		FinanceAmount: 38500.5,
		State:         leads.StateQLD,
	}
}

func TestValidation(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.November, 4, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		change func(*leads.Lead)
		field  string
	}{
		{name: "valid", change: func(_ *leads.Lead) {}},
		{name: "first name", change: func(lead *leads.Lead) { lead.FirstName = "  " }, field: "first_name"},
		{name: "last name", change: func(lead *leads.Lead) { lead.LastName = "" }, field: "last_name"},
		{name: "missing email", change: func(lead *leads.Lead) { lead.Email = "" }, field: "email"},
		{name: "email", change: func(lead *leads.Lead) { lead.Email = "bad-address" }, field: "email"},
		{name: "display name", change: func(lead *leads.Lead) { lead.Email = "Zoe <zoe@example.com>" }, field: "email"},
		{name: "zero amount", change: func(lead *leads.Lead) { lead.FinanceAmount = 0 }, field: "finance_amount"},
		{name: "negative amount", change: func(lead *leads.Lead) { lead.FinanceAmount = -1 }, field: "finance_amount"},
		{
			name:   "nonfinite amount",
			change: func(lead *leads.Lead) { lead.FinanceAmount = math.Inf(1) },
			field:  "finance_amount",
		},
		{
			name:   "NaN salary",
			change: func(lead *leads.Lead) { salary := math.NaN(); lead.Salary = &salary },
			field:  "salary",
		},
		{name: "zero salary", change: func(lead *leads.Lead) { salary := 0.0; lead.Salary = &salary }, field: "salary"},
		{
			name:   "negative salary",
			change: func(lead *leads.Lead) { salary := -1.0; lead.Salary = &salary },
			field:  "salary",
		},
		{name: "positive salary", change: func(lead *leads.Lead) { salary := 100000.0; lead.Salary = &salary }},
		{name: "invalid state", change: func(lead *leads.Lead) { lead.State = "XX" }, field: "state"},
		{name: "missing state", change: func(lead *leads.Lead) { lead.State = "" }, field: "state"},
		{name: "make only", change: func(lead *leads.Lead) { lead.Vehicle = &leads.Vehicle{Make: "Kia"} }},
		{name: "empty vehicle", change: func(lead *leads.Lead) { lead.Vehicle = &leads.Vehicle{} }},
		{
			name:   "model without make",
			change: func(lead *leads.Lead) { lead.Vehicle = &leads.Vehicle{Model: "EV6"} },
			field:  "vehicle.make",
		},
		{
			name:   "year without make",
			change: func(lead *leads.Lead) { year := 2026; lead.Vehicle = &leads.Vehicle{Year: &year} },
			field:  "vehicle.make",
		},
		{
			name:   "old year",
			change: func(lead *leads.Lead) { year := 1989; lead.Vehicle = &leads.Vehicle{Make: "Kia", Year: &year} },
			field:  "vehicle.year",
		},
		{
			name:   "future year",
			change: func(lead *leads.Lead) { year := 2028; lead.Vehicle = &leads.Vehicle{Make: "Kia", Year: &year} },
			field:  "vehicle.year",
		},
		{
			name:   "minimum year",
			change: func(lead *leads.Lead) { year := 1990; lead.Vehicle = &leads.Vehicle{Make: "Kia", Year: &year} },
		},
		{
			name:   "next year",
			change: func(lead *leads.Lead) { year := 2027; lead.Vehicle = &leads.Vehicle{Make: "Kia", Year: &year} },
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			lead := validLead()
			test.change(&lead)
			lead.Normalize()
			fields := lead.Validate(now)
			if test.field == "" && len(fields) != 0 {
				t.Fatalf("unexpected validation errors: %v", fields)
			}
			if test.field != "" && fields[test.field] == "" {
				t.Fatalf("expected error for %s", test.field)
			}
		})
	}
}

func TestNormalization(t *testing.T) {
	t.Parallel()
	for input, expected := range map[string]string{
		"ACT": leads.StateACT, "NSW": leads.StateNSW, "NT": leads.StateNT,
		"QLD": leads.StateQLD, "SA": leads.StateSA, "TAS": leads.StateTAS,
		"VIC": leads.StateVIC, "WA": leads.StateWA,
		" Australian Capital Territory ": leads.StateACT,
		" new south wales ":              leads.StateNSW,
		" Northern Territory ":           leads.StateNT,
		" Queensland ":                   leads.StateQLD,
		" South Australia ":              leads.StateSA,
		" Tasmania ":                     leads.StateTAS,
		" Victoria ":                     leads.StateVIC,
		" Western Australia ":            leads.StateWA,
	} {
		lead := validLead()
		lead.State = input
		lead.Email = " ZOE@EXAMPLE.COM "
		lead.Normalize()
		if lead.State != expected {
			t.Errorf("Normalize state %q = %q, want %q", input, lead.State, expected)
		}
		if lead.Email != "zoe@example.com" || len(lead.Validate(time.Now())) != 0 {
			t.Fatal("normalization failed")
		}
	}
}
