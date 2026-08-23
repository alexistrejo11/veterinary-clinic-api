package addresses

import "testing"

func TestParseCountryAcceptsSupportedCountries(t *testing.T) {
	tests := map[string]Country{
		"USA":           USA,
		"united states": USA,
		"Mexico":        Mexico,
		"canada":        Canada,
	}

	for input, want := range tests {
		got, err := ParseCountry(input)
		if err != nil || got != want {
			t.Fatalf("ParseCountry(%q) = %q, %v; want %q, nil", input, got, err, want)
		}
	}
}

func TestParseCountryRejectsUnsupportedCountry(t *testing.T) {
	if _, err := ParseCountry("Spain"); err == nil {
		t.Fatal("ParseCountry accepted an unsupported country")
	}
}
