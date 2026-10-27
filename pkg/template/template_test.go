package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	texttemplate "text/template"
)

func TestParseWithData(t *testing.T) {
	tmpl, err := NewTemplate(`Hello {{ .Name }}`, WithDefaultNgrams(false))
	if err != nil {
		t.Fatalf("NewTemplate returned an error: %v", err)
	}

	text, err := tmpl.ParseString(WithData(map[string]string{"Name": "Aaron"}))
	if err != nil {
		t.Fatalf("ParseString returned an error: %v", err)
	}
	if text != "Hello Aaron" {
		t.Fatalf("unexpected text: %q", text)
	}
}

func TestDefaultNgramsLazyLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping default corpus load in short mode")
	}

	tmpl, err := NewTemplate(`{{ randomText 3 }}`)
	if err != nil {
		t.Fatalf("NewTemplate returned an error: %v", err)
	}

	text, err := tmpl.ParseString()
	if err != nil {
		t.Fatalf("ParseString returned an error: %v", err)
	}
	if len(strings.Fields(text)) != 3 {
		t.Fatalf("expected 3 words, got %d: %q", len(strings.Fields(text)), text)
	}
}

func TestMissingKeyError(t *testing.T) {
	tmpl, err := NewTemplate(
		`{{ .Missing }}`,
		WithDefaultNgrams(false),
		WithMissingKeyError(true),
	)
	if err != nil {
		t.Fatalf("NewTemplate returned an error: %v", err)
	}

	if _, err := tmpl.ParseString(WithData(map[string]string{})); err == nil {
		t.Fatal("expected an error for a missing map key")
	}
}

func TestExtraFuncsOverrideDefaults(t *testing.T) {
	tmpl, err := NewTemplate(
		`{{ randomName }}`,
		WithDefaultNgrams(false),
		WithExtraFuncs(texttemplate.FuncMap{
			"randomName": func() string { return "Custom" },
		}),
	)
	if err != nil {
		t.Fatalf("NewTemplate returned an error: %v", err)
	}

	text, err := tmpl.ParseString()
	if err != nil {
		t.Fatalf("ParseString returned an error: %v", err)
	}
	if text != "Custom" {
		t.Fatalf("unexpected text: %q", text)
	}
}

func TestRandomChoiceListWithTypedSlice(t *testing.T) {
	tmpl, err := NewTemplate(`{{ randomChoiceList .Names }}`, WithDefaultNgrams(false))
	if err != nil {
		t.Fatalf("NewTemplate returned an error: %v", err)
	}

	text, err := tmpl.ParseString(WithData(map[string]any{"Names": []string{"only"}}))
	if err != nil {
		t.Fatalf("ParseString returned an error: %v", err)
	}
	if text != "only" {
		t.Fatalf("unexpected text: %q", text)
	}
}

func TestNestedTemplates(t *testing.T) {
	tmpl, err := NewTemplate(
		`{{ template "child" . }}`,
		WithDefaultNgrams(false),
		WithNestedTemplates(map[string]string{"child": `Hi {{ .Name }}`}),
	)
	if err != nil {
		t.Fatalf("NewTemplate returned an error: %v", err)
	}

	text, err := tmpl.ParseString(WithData(map[string]string{"Name": "Aaron"}))
	if err != nil {
		t.Fatalf("ParseString returned an error: %v", err)
	}
	if text != "Hi Aaron" {
		t.Fatalf("unexpected text: %q", text)
	}

	text, err = tmpl.ParseString(
		WithExecuteTemplate("child"),
		WithData(map[string]string{"Name": "Aaron"}),
	)
	if err != nil {
		t.Fatalf("ParseString returned an error: %v", err)
	}
	if text != "Hi Aaron" {
		t.Fatalf("unexpected text: %q", text)
	}
}

func TestNestedTemplateNameConflict(t *testing.T) {
	_, err := NewTemplate(
		`root`,
		WithName("root"),
		WithDefaultNgrams(false),
		WithNestedTemplates(map[string]string{"root": `child`}),
	)
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestChecksumIsStable(t *testing.T) {
	first, err := NewTemplate(`Hello {{ .Name }}`, WithDefaultNgrams(false))
	if err != nil {
		t.Fatalf("NewTemplate returned an error: %v", err)
	}
	second, err := NewTemplate(`Hello {{ .Name }}`, WithDefaultNgrams(false))
	if err != nil {
		t.Fatalf("NewTemplate returned an error: %v", err)
	}
	if first.Checksum() != second.Checksum() {
		t.Fatalf("checksums differ: %d != %d", first.Checksum(), second.Checksum())
	}

	different, err := NewTemplate(`Hello {{ .Other }}`, WithDefaultNgrams(false))
	if err != nil {
		t.Fatalf("NewTemplate returned an error: %v", err)
	}
	if first.Checksum() == different.Checksum() {
		t.Fatal("expected different checksums")
	}
}

func TestInvalidFuncReturnsError(t *testing.T) {
	_, err := NewTemplate(
		`template`,
		WithDefaultNgrams(false),
		WithExtraFuncs(texttemplate.FuncMap{"invalid": 42}),
	)
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestFileReadRestrictions(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("hello"), 0o600); err != nil {
		t.Fatalf("WriteFile returned an error: %v", err)
	}

	outside := filepath.Join(filepath.Dir(dir), "outside.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatalf("WriteFile returned an error: %v", err)
	}

	tmpl, err := NewTemplate(
		`template`,
		WithDefaultNgrams(false),
		WithBaseDir(dir),
		WithMaxFileSize(10),
	)
	if err != nil {
		t.Fatalf("NewTemplate returned an error: %v", err)
	}

	data, err := tmpl.ReadFileBytes("file.txt")
	if err != nil {
		t.Fatalf("ReadFileBytes returned an error: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected file content: %q", data)
	}

	if _, err := tmpl.ReadFileBytes(filepath.Join("..", filepath.Base(outside))); err == nil {
		t.Fatal("expected an error for a path outside the base directory")
	}

	large := filepath.Join(dir, "large.txt")
	if err := os.WriteFile(large, []byte("01234567890"), 0o600); err != nil {
		t.Fatalf("WriteFile returned an error: %v", err)
	}
	if _, err := tmpl.ReadFileBytes("large.txt"); err == nil {
		t.Fatal("expected an error for a file larger than the maximum size")
	}
}

func TestBase64Helpers(t *testing.T) {
	tmpl, err := NewTemplate(`template`, WithDefaultNgrams(false))
	if err != nil {
		t.Fatalf("NewTemplate returned an error: %v", err)
	}

	if got := tmpl.Base64String("hello"); got != "aGVsbG8=" {
		t.Fatalf("unexpected encoded value: %q", got)
	}
	decoded, err := tmpl.Base64Decode("aGVsbG8=")
	if err != nil {
		t.Fatalf("Base64Decode returned an error: %v", err)
	}
	if string(decoded) != "hello" {
		t.Fatalf("unexpected decoded value: %q", decoded)
	}
}
