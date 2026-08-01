// Command breaking-check проверяет обратную совместимость JSON Schema
// событий (registry). Использование:
//
//	breaking-check schemas/            # проверить все схемы: old vs new (git diff)
//	breaking-check old.json new.json   # сравнить две схемы
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/devcen-online/ecom-schema-registry/internal/breaking"
)

const exitOK = 0

func main() {
	args := os.Args[1:]
	if len(args) == 2 {
		os.Exit(checkPair(args[0], args[1]))
	}
	if len(args) == 1 {
		os.Exit(checkDir(args[0]))
	}
	fmt.Fprintln(os.Stderr, "usage: breaking-check <old.json> <new.json> | breaking-check <dir>")
	os.Exit(2)
}

func checkPair(oldPath, newPath string) int {
	oldB, err := os.ReadFile(oldPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", oldPath, err)
		return 1
	}
	newB, err := os.ReadFile(newPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", newPath, err)
		return 1
	}
	issues, err := breaking.Check(oldB, newB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse: %v\n", err)
		return 1
	}
	return report(oldPath, issues)
}

func checkDir(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "readdir %s: %v\n", dir, err)
		return 1
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		fmt.Fprintf(os.Stderr, "no schemas in %s\n", dir)
		return 1
	}
	// В режиме каталога проверяем, что каждая схема парсится и имеет $id/type.
	fail := false
	for _, n := range names {
		path := filepath.Join(dir, n)
		b, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
			fail = true
			continue
		}
		s, err := breaking.Parse(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", path, err)
			fail = true
			continue
		}
		var meta struct {
			ID string `json:"$id"`
		}
		_ = json.Unmarshal(b, &meta)
		if meta.ID == "" || s.Type == "" {
			fmt.Fprintf(os.Stderr, "%s: missing $id or type\n", path)
			fail = true
			continue
		}
		fmt.Printf("ok %s\n", path)
	}
	if fail {
		return 1
	}
	return exitOK
}

func report(name string, issues []breaking.Issue) int {
	fail := false
	for _, i := range issues {
		level := i.Level
		if level == "breaking" {
			fail = true
		}
		fmt.Printf("%s\n", issueLine(i))
	}
	if fail {
		fmt.Printf("%s: BREAKING CHANGES DETECTED\n", name)
		return 1
	}
	fmt.Printf("%s: ok\n", name)
	return exitOK
}

// issueLine — формат строки замечания, зафиксированный в BDD-032#S-7:
// "[breaking] $.price: удалено обязательное поле" (без префикса файла).
func issueLine(i breaking.Issue) string {
	return fmt.Sprintf("[%s] %s: %s", i.Level, i.Path, i.Message)
}
