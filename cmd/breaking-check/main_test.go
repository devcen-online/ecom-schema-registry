package main

import (
	"testing"

	"github.com/devcen-online/ecom-schema-registry/internal/breaking"
)

// TestIssueLineFormat — BDD-032#S-7: CI выводит точную строку
// "[breaking] $.price: удалено обязательное поле".
func TestIssueLineFormat(t *testing.T) {
	got := issueLine(breaking.Issue{Level: "breaking", Path: "$.price", Message: "удалено обязательное поле"})
	want := "[breaking] $.price: удалено обязательное поле"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestIssueLineWarning(t *testing.T) {
	got := issueLine(breaking.Issue{Level: "warning", Path: "$.name", Message: "удалено необязательное поле"})
	want := "[warning] $.name: удалено необязательное поле"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}
