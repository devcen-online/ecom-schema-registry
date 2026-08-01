// Package breaking проверяет обратную совместимость JSON Schema
// (Protobuf/JSON Schema registry, FR-004).
//
// Критерии breaking (упрощённая модель, достаточная для registry):
//   - удалено обязательное поле (required без default) — breaking;
//   - добавлено обязательное поле — breaking;
//   - изменён тип существующего поля — breaking;
//   - изменён enum (удалено значение или изменён список) — breaking;
//   - сужен диапазон (minimum увеличен, maximum уменьшен) — breaking.
package breaking

import "encoding/json"

type Schema struct {
	Type       string          `json:"type"`
	Required   []string        `json:"required"`
	Properties map[string]Prop `json:"properties"`
	Enum       []any           `json:"enum,omitempty"`
	Minimum    *float64        `json:"minimum,omitempty"`
	Maximum    *float64        `json:"maximum,omitempty"`
	Items      *Schema         `json:"items,omitempty"`
}

type Prop struct {
	Type    string   `json:"type"`
	Enum    []any    `json:"enum,omitempty"`
	Minimum *float64 `json:"minimum,omitempty"`
	Maximum *float64 `json:"maximum,omitempty"`
	Items   *Schema  `json:"items,omitempty"`
}

type Issue struct {
	Level   string `json:"level"` // "breaking" | "warning"
	Path    string `json:"path"`
	Message string `json:"message"`
}

func Parse(data []byte) (*Schema, error) {
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.Properties == nil {
		s.Properties = map[string]Prop{}
	}
	return &s, nil
}

func Check(oldData, newData []byte) ([]Issue, error) {
	oldS, err := Parse(oldData)
	if err != nil {
		return nil, err
	}
	newS, err := Parse(newData)
	if err != nil {
		return nil, err
	}
	return diff(oldS, newS), nil
}

func diff(oldS, newS *Schema) []Issue {
	var issues []Issue
	issues = append(issues, diffSchema("$", oldS, newS)...)
	return issues
}

func diffSchema(path string, oldS, newS *Schema) []Issue {
	var issues []Issue

	// Изменение типа корня.
	if oldS.Type != newS.Type {
		issues = append(issues, Issue{"breaking", path, "тип корня изменён: " + oldS.Type + " -> " + newS.Type})
	}

	// Enum корня.
	if s := diffEnum(path, oldS.Enum, newS.Enum); len(s) > 0 {
		issues = append(issues, s...)
	}

	// Required: удалённое обязательное поле — breaking; добавленное — breaking.
	oldReq := map[string]bool{}
	for _, r := range oldS.Required {
		oldReq[r] = true
	}
	newReq := map[string]bool{}
	for _, r := range newS.Required {
		newReq[r] = true
	}
	for _, r := range newS.Required {
		if !oldReq[r] {
			issues = append(issues, Issue{"breaking", path + "." + r, "добавлено обязательное поле (несовместимо со старыми продюсерами)"})
		}
	}

	// Свойства.
	for name, oldP := range oldS.Properties {
		newP, exists := newS.Properties[name]
		if !exists {
			if oldReq[name] {
				issues = append(issues, Issue{"breaking", path + "." + name, "удалено обязательное поле"})
			} else {
				issues = append(issues, Issue{"warning", path + "." + name, "удалено необязательное поле"})
			}
			continue
		}
		if oldP.Type != newP.Type {
			issues = append(issues, Issue{"breaking", path + "." + name, "тип изменён: "+oldP.Type+" -> "+newP.Type})
		}
		if s := diffEnum(path+"."+name, oldP.Enum, newP.Enum); len(s) > 0 {
			issues = append(issues, s...)
		}
		if oldP.Minimum != nil && newP.Minimum != nil && *newP.Minimum > *oldP.Minimum {
			issues = append(issues, Issue{"breaking", path + "." + name, "minimum увеличен: " + fmtF(*oldP.Minimum) + " -> " + fmtF(*newP.Minimum)})
		}
		if oldP.Maximum != nil && newP.Maximum != nil && *newP.Maximum < *oldP.Maximum {
			issues = append(issues, Issue{"breaking", path + "." + name, "maximum уменьшен: " + fmtF(*oldP.Maximum) + " -> " + fmtF(*newP.Maximum)})
		}
		if s := diffItems(path+"."+name, oldP.Items, newP.Items); len(s) > 0 {
			issues = append(issues, s...)
		}
	}
	// Добавленные необязательные поля — не breaking (совместимо).
	if s := diffItems(path, oldS.Items, newS.Items); len(s) > 0 {
		issues = append(issues, s...)
	}
	return issues
}

// diffItems рекурсивно сравнивает элементы массивов (items).
func diffItems(path string, oldI, newI *Schema) []Issue {
	if oldI == nil && newI == nil {
		return nil
	}
	if oldI == nil || newI == nil {
		return []Issue{{"breaking", path + ".items", "добавлен или удалён массив (items)"}}
	}
	return diffSchema(path+".items", oldI, newI)
}

func diffEnum(path string, oldE, newE []any) []Issue {
	if len(oldE) == 0 && len(newE) == 0 {
		return nil
	}
	oldSet := map[string]bool{}
	for _, v := range oldE {
		oldSet[anyString(v)] = true
	}
	newSet := map[string]bool{}
	for _, v := range newE {
		newSet[anyString(v)] = true
	}
	var issues []Issue
	for v := range oldSet {
		if !newSet[v] {
			issues = append(issues, Issue{"breaking", path, "удалено значение enum: " + v})
		}
	}
	if len(oldE) > 0 && len(newE) == 0 {
		issues = append(issues, Issue{"breaking", path, "enum удалён целиком"})
	}
	return issues
}

func anyString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func fmtF(f float64) string {
	return anyString(f)
}
