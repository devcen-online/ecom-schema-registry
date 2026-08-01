package breaking

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNoChangesNoIssues(t *testing.T) {
	issues, err := Check(mustMarshal(offerSchemaV1()), mustMarshal(offerSchemaV1()))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("идентичные схемы не должны давать замечаний: %v", issues)
	}
}

func TestRemovedRequiredField(t *testing.T) {
	oldS := offerSchemaV1()
	newS := offerSchemaV1()
	delete(newS.Properties, "category_id")
	newS.Required = newS.Required[:len(newS.Required)-1]

	issues, _ := Check(mustMarshal(oldS), mustMarshal(newS))
	if !hasBreaking(issues, "category_id") {
		t.Fatalf("удаление обязательного поля должно быть breaking: %v", issues)
	}
}

func TestAddedRequiredField(t *testing.T) {
	oldS := offerSchemaV1()
	newS := offerSchemaV1()
	newS.Properties["weight_g"] = Prop{Type: "integer"}
	newS.Required = append(newS.Required, "weight_g")

	issues, _ := Check(mustMarshal(oldS), mustMarshal(newS))
	if !hasBreaking(issues, "weight_g") {
		t.Fatalf("добавление обязательного поля должно быть breaking: %v", issues)
	}
}

func TestTypeChange(t *testing.T) {
	oldS := offerSchemaV1()
	newS := offerSchemaV1()
	newS.Properties["price"] = Prop{Type: "string"}

	issues, _ := Check(mustMarshal(oldS), mustMarshal(newS))
	if !hasBreaking(issues, "price") {
		t.Fatalf("изменение типа должно быть breaking: %v", issues)
	}
}

func TestRemovedOptionalFieldIsWarning(t *testing.T) {
	oldS := offerSchemaV1()
	newS := offerSchemaV1()
	delete(newS.Properties, "name")

	issues, _ := Check(mustMarshal(oldS), mustMarshal(newS))
	if hasBreaking(issues, "name") {
		t.Fatalf("удаление необязательного поля не должно быть breaking: %v", issues)
	}
	found := false
	for _, i := range issues {
		if i.Level == "warning" && i.Path == "$.name" {
			found = true
		}
	}
	if !found {
		t.Fatalf("ожидалось warning для $.name: %v", issues)
	}
}

func TestEnumValueRemoved(t *testing.T) {
	oldS := offerSchemaV1()
	newS := offerSchemaV1()
	oldS.Properties["status"] = Prop{Type: "string", Enum: []any{"active", "suspended", "blocked"}}
	newS.Properties["status"] = Prop{Type: "string", Enum: []any{"active", "suspended"}}

	issues, _ := Check(mustMarshal(oldS), mustMarshal(newS))
	if !hasBreaking(issues, "status") {
		t.Fatalf("удаление значения enum должно быть breaking: %v", issues)
	}
}

func TestMinimumRaised(t *testing.T) {
	oldS := offerSchemaV1()
	newS := offerSchemaV1()
	min0 := 0.0
	oldS.Properties["price"] = Prop{Type: "integer", Minimum: &min0}
	min100 := 100.0
	newS.Properties["price"] = Prop{Type: "integer", Minimum: &min100}

	issues, _ := Check(mustMarshal(oldS), mustMarshal(newS))
	if !hasBreaking(issues, "price") {
		t.Fatalf("увеличение minimum должно быть breaking: %v", issues)
	}
}

func mustMarshal(m any) []byte {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

func hasBreaking(issues []Issue, pathPart string) bool {
	for _, i := range issues {
		if i.Level == "breaking" && strings.Contains(i.Path, pathPart) {
			return true
		}
	}
	return false
}

func offerSchemaV1() *Schema {
	return &Schema{
		Type: "object",
		Required: []string{
			"event_id", "aggregate_id", "aggregate_version",
			"offer_id", "product_id", "unit_id", "price", "category_id",
		},
		Properties: map[string]Prop{
			"event_id":          {Type: "string"},
			"aggregate_id":      {Type: "string"},
			"aggregate_version": {Type: "integer"},
			"offer_id":          {Type: "integer"},
			"product_id":        {Type: "integer"},
			"unit_id":           {Type: "string"},
			"seller_id":         {Type: "string"},
			"name":              {Type: "string"},
			"price":             {Type: "integer"},
			"stock":             {Type: "integer"},
			"category_id":       {Type: "integer"},
		},
	}
}
