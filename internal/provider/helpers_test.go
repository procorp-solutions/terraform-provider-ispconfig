package provider

import (
	"testing"
)

func TestBoolToYN(t *testing.T) {
	tests := []struct {
		input bool
		want  string
	}{
		{true, "y"},
		{false, "n"},
	}
	for _, tt := range tests {
		got := boolToYN(tt.input)
		if got != tt.want {
			t.Errorf("boolToYN(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestYnToBool(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"y", true},
		{"Y", true},
		{"n", false},
		{"N", false},
		{"", false},
		{"yes", false},
		{"no", false},
	}
	for _, tt := range tests {
		got := ynToBool(tt.input)
		if got != tt.want {
			t.Errorf("ynToBool(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestMbToAPIQuota(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  int64
	}{
		{"positive", 100, 100 * 1024 * 1024},
		{"one MB", 1, 1024 * 1024},
		{"zero", 0, 0},
		{"unlimited", -1, -1},
		{"negative", -5, -5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mbToAPIQuota(tt.input)
			if got != tt.want {
				t.Errorf("mbToAPIQuota(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestApiQuotaToMB(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  int64
	}{
		{"positive", 100 * 1024 * 1024, 100},
		{"one MB", 1024 * 1024, 1},
		{"zero", 0, 0},
		{"unlimited", -1, -1},
		{"negative", -5, -5},
		{"partial MB truncates", 1500000, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := apiQuotaToMB(tt.input)
			if got != tt.want {
				t.Errorf("apiQuotaToMB(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestMbToAPIQuota_RoundTrip(t *testing.T) {
	for _, mb := range []int64{0, 1, 10, 100, 1024, -1} {
		got := apiQuotaToMB(mbToAPIQuota(mb))
		if got != mb {
			t.Errorf("round-trip failed for %d: got %d", mb, got)
		}
	}
}

func TestParseCronSchedule(t *testing.T) {
	tests := []struct {
		name                                        string
		input                                       string
		wantMin, wantHour, wantMday, wantMon, wantW string
		wantErr                                     bool
	}{
		{
			name:     "standard",
			input:    "*/5 * * * *",
			wantMin:  "*/5",
			wantHour: "*",
			wantMday: "*",
			wantMon:  "*",
			wantW:    "*",
		},
		{
			name:     "specific values",
			input:    "0 3 15 6 1",
			wantMin:  "0",
			wantHour: "3",
			wantMday: "15",
			wantMon:  "6",
			wantW:    "1",
		},
		{
			name:     "extra whitespace collapsed",
			input:    "0  3  15  6  1",
			wantMin:  "0",
			wantHour: "3",
			wantMday: "15",
			wantMon:  "6",
			wantW:    "1",
		},
		{
			name:    "too few fields",
			input:   "* * *",
			wantErr: true,
		},
		{
			name:    "too many fields",
			input:   "* * * * * *",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, hour, mday, mon, wday, err := parseCronSchedule(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if min != tt.wantMin || hour != tt.wantHour || mday != tt.wantMday || mon != tt.wantMon || wday != tt.wantW {
				t.Errorf("got (%q,%q,%q,%q,%q), want (%q,%q,%q,%q,%q)",
					min, hour, mday, mon, wday,
					tt.wantMin, tt.wantHour, tt.wantMday, tt.wantMon, tt.wantW)
			}
		})
	}
}

func TestBuildCronSchedule(t *testing.T) {
	got := buildCronSchedule("*/5", "*", "*", "*", "*")
	want := "*/5 * * * *"
	if got != want {
		t.Errorf("buildCronSchedule() = %q, want %q", got, want)
	}
}

func TestCronScheduleRoundTrip(t *testing.T) {
	schedule := "30 2 15 6 3"
	min, hour, mday, mon, wday, err := parseCronSchedule(schedule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := buildCronSchedule(min, hour, mday, mon, wday)
	if got != schedule {
		t.Errorf("round-trip: got %q, want %q", got, schedule)
	}
}
