package client

import (
	"encoding/json"
	"testing"
)

func TestFlexInt_UnmarshalJSON_Int(t *testing.T) {
	var fi FlexInt
	if err := json.Unmarshal([]byte("42"), &fi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fi != 42 {
		t.Errorf("got %d, want 42", fi)
	}
}

func TestFlexInt_UnmarshalJSON_NegativeInt(t *testing.T) {
	var fi FlexInt
	if err := json.Unmarshal([]byte("-1"), &fi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fi != -1 {
		t.Errorf("got %d, want -1", fi)
	}
}

func TestFlexInt_UnmarshalJSON_Zero(t *testing.T) {
	var fi FlexInt
	if err := json.Unmarshal([]byte("0"), &fi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fi != 0 {
		t.Errorf("got %d, want 0", fi)
	}
}

func TestFlexInt_UnmarshalJSON_String(t *testing.T) {
	var fi FlexInt
	if err := json.Unmarshal([]byte(`"123"`), &fi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fi != 123 {
		t.Errorf("got %d, want 123", fi)
	}
}

func TestFlexInt_UnmarshalJSON_EmptyString(t *testing.T) {
	var fi FlexInt
	if err := json.Unmarshal([]byte(`""`), &fi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fi != 0 {
		t.Errorf("got %d, want 0", fi)
	}
}

func TestFlexInt_UnmarshalJSON_NonNumericString(t *testing.T) {
	var fi FlexInt
	err := json.Unmarshal([]byte(`"abc"`), &fi)
	if err == nil {
		t.Fatal("expected error for non-numeric string, got nil")
	}
}

func TestFlexInt_UnmarshalJSON_Boolean(t *testing.T) {
	var fi FlexInt
	err := json.Unmarshal([]byte("true"), &fi)
	if err == nil {
		t.Fatal("expected error for boolean, got nil")
	}
}

func TestFlexInt_UnmarshalJSON_Null(t *testing.T) {
	// Go's json.Unmarshal treats null as the zero value for numeric types,
	// so FlexInt correctly returns 0 with no error.
	var fi FlexInt
	if err := json.Unmarshal([]byte("null"), &fi); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fi != 0 {
		t.Errorf("got %d, want 0", fi)
	}
}

func TestFlexInt_UnmarshalJSON_Array(t *testing.T) {
	var fi FlexInt
	err := json.Unmarshal([]byte("[1,2]"), &fi)
	if err == nil {
		t.Fatal("expected error for array, got nil")
	}
}

func TestFlexInt_InStruct(t *testing.T) {
	type S struct {
		ID   FlexInt `json:"id"`
		Name string  `json:"name"`
	}
	input := `{"id":"99","name":"test"}`
	var s S
	if err := json.Unmarshal([]byte(input), &s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != 99 {
		t.Errorf("got ID %d, want 99", s.ID)
	}
	if s.Name != "test" {
		t.Errorf("got Name %q, want %q", s.Name, "test")
	}
}

func TestParseResponseID_Float64(t *testing.T) {
	id, err := parseResponseID(float64(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("got %d, want 42", id)
	}
}

func TestParseResponseID_String(t *testing.T) {
	id, err := parseResponseID("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 123 {
		t.Errorf("got %d, want 123", id)
	}
}

func TestParseResponseID_InvalidString(t *testing.T) {
	_, err := parseResponseID("abc")
	if err == nil {
		t.Fatal("expected error for non-numeric string, got nil")
	}
}

func TestParseResponseID_UnexpectedType(t *testing.T) {
	_, err := parseResponseID(true)
	if err == nil {
		t.Fatal("expected error for bool type, got nil")
	}
}

func TestUnmarshalResponse(t *testing.T) {
	resp := map[string]interface{}{
		"database_name": "testdb",
		"active":        "y",
	}
	var target struct {
		DatabaseName string `json:"database_name"`
		Active       string `json:"active"`
	}
	if err := unmarshalResponse(resp, &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.DatabaseName != "testdb" {
		t.Errorf("got %q, want %q", target.DatabaseName, "testdb")
	}
	if target.Active != "y" {
		t.Errorf("got %q, want %q", target.Active, "y")
	}
}
