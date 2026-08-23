package employees

import (
	"testing"
	"time"
)

func TestParseWeekday(t *testing.T) {
	day, err := parseWeekday("Wednesday")
	if err != nil {
		t.Fatal(err)
	}
	if day != time.Wednesday {
		t.Fatalf("parseWeekday returned %v, want %v", day, time.Wednesday)
	}

	if _, err := parseWeekday("weekday"); err == nil {
		t.Fatal("parseWeekday accepted an invalid day")
	}
}
