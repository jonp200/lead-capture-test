// Package leads validates, imports, and stores partner leads.
package leads

import (
	"math"
	"net/mail"
	"strings"
	"time"
)

// Vehicle contains the optional vehicle details.
type Vehicle struct {
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  *int   `json:"year"`
}

// Lead is the shared API and storage representation. Amounts are in dollars.
type Lead struct {
	ID            int64    `json:"id"`
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	Email         string   `json:"email"`
	FinanceAmount float64  `json:"finance_amount"`
	Salary        *float64 `json:"salary"`
	State         string   `json:"state"`
	Vehicle       *Vehicle `json:"vehicle"`
	CreatedAt     string   `json:"created_at"`
}

// Normalize trims input and canonicalizes email and Australian state names.
func (lead *Lead) Normalize() {
	lead.FirstName = strings.TrimSpace(lead.FirstName)
	lead.LastName = strings.TrimSpace(lead.LastName)
	lead.Email = strings.ToLower(strings.TrimSpace(lead.Email))
	lead.State = strings.ToUpper(strings.TrimSpace(lead.State))
	aliases := map[string]string{
		"NEW SOUTH WALES": "NSW", "VICTORIA": "VIC", "QUEENSLAND": "QLD",
		"SOUTH AUSTRALIA": "SA", "WESTERN AUSTRALIA": "WA", "TASMANIA": "TAS",
		"NORTHERN TERRITORY": "NT", "AUSTRALIAN CAPITAL TERRITORY": "ACT",
	}
	if state, ok := aliases[lead.State]; ok {
		lead.State = state
	}
	if lead.Vehicle != nil {
		lead.Vehicle.Make = strings.TrimSpace(lead.Vehicle.Make)
		lead.Vehicle.Model = strings.TrimSpace(lead.Vehicle.Model)
		if lead.Vehicle.Make == "" && lead.Vehicle.Model == "" && lead.Vehicle.Year == nil {
			lead.Vehicle = nil
		}
	}
}

// Validate returns field-specific errors without including submitted values.
func (lead *Lead) Validate(now time.Time) map[string]string {
	errors := make(map[string]string)
	if lead.FirstName == "" {
		errors["first_name"] = "is required"
	}
	if lead.LastName == "" {
		errors["last_name"] = "is required"
	}
	address, err := mail.ParseAddress(lead.Email)
	if err != nil || address.Address != lead.Email {
		errors["email"] = "must be a valid email address"
	}
	if !positive(lead.FinanceAmount) {
		errors["finance_amount"] = "must be greater than 0"
	}
	if lead.Salary != nil && !positive(*lead.Salary) {
		errors["salary"] = "must be greater than 0"
	}
	switch lead.State {
	case "NSW", "VIC", "QLD", "SA", "WA", "TAS", "NT", "ACT":
	default:
		errors["state"] = "must be a valid Australian state or territory"
	}
	validateVehicle(lead.Vehicle, now.Year()+1, errors)
	return errors
}

func positive(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func validateVehicle(vehicle *Vehicle, maxYear int, errors map[string]string) {
	const minYear = 1990
	if vehicle == nil {
		return
	}
	if vehicle.Make == "" {
		errors["vehicle.make"] = "is required when a model or year is given"
	}
	if vehicle.Year != nil && (*vehicle.Year < minYear || *vehicle.Year > maxYear) {
		errors["vehicle.year"] = "must be between 1990 and next year"
	}
}
