package validation

import "strings"

type Violations map[string]string

func (v Violations) Empty() bool { return len(v) == 0 }

// Basic validators
func Required(field, value string, v Violations) {
	if strings.TrimSpace(value) == "" {
		v[field] = "required"
	}
}

func PositiveFloat(field string, val float64, v Violations) {
	if val <= 0 {
		v[field] = "must_be_positive"
	}
}

func RangeFloat(field string, val, minVal, maxVal float64, v Violations) {
	if val < minVal || val > maxVal {
		v[field] = "out_of_range"
	}
}
