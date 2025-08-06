package internal

import "testing"

func TestValidateActionYML_Required(t *testing.T) {
	t.Parallel()

	a := &ActionYML{
		Name:        "",
		Description: "",
		Runs:        map[string]any{},
	}
	res := ValidateActionYML(a)
	if len(res.MissingFields) == 0 {
		t.Error("should detect missing fields")
	}
}

func TestValidateActionYML_Valid(t *testing.T) {
	t.Parallel()
	a := &ActionYML{
		Name:        "MyAction",
		Description: "desc",
		Runs:        map[string]any{"using": "node12"},
	}
	res := ValidateActionYML(a)
	if len(res.MissingFields) != 0 {
		t.Errorf("expected no missing fields, got %v", res.MissingFields)
	}
}
