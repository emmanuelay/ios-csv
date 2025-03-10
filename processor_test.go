package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConvertToIOS(t *testing.T) {
	tempDir := "./"
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")

	if err := os.Mkdir(inputDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(outputDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test CSV
	csvContent := "Key,en,sv\nhello,Hello,Hej\nworld,World,Världen"
	if err := os.WriteFile(filepath.Join(inputDir, "test.csv"), []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	config := Config{
		InPath:    inputDir,
		OutPath:   outputDir,
		Delimiter: ",",
	}
	processor := NewLocalizationProcessor(config)
	if err := processor.ConvertToIOS(); err != nil {
		t.Fatalf("convert to iOS failed: %v", err)
	}

	// Check generated files
	for _, lang := range []string{"en", "sv"} {
		path := filepath.Join(outputDir, lang+".lproj", "Localizable.strings")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}
		expected := "\n/* test.csv */\n\"hello\" = \"" + map[string]string{"en": "Hello", "sv": "Hej"}[lang] + "\";\n\"world\" = \"" + map[string]string{"en": "World", "sv": "Världen"}[lang] + "\";\n"
		if string(content) != expected {
			t.Errorf("unexpected content in %s:\nGot: %s\nWant: %s", path, content, expected)
		}
	}

	if err := os.RemoveAll(inputDir); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(outputDir); err != nil {
		t.Fatal(err)
	}
}

func TestConvertToCSV(t *testing.T) {
	tempDir := "./"
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")

	if err := os.Mkdir(inputDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(outputDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test .lproj files
	for _, lang := range []string{"en", "sv"} {
		langDir := filepath.Join(inputDir, lang+".lproj")
		if err := os.Mkdir(langDir, 0755); err != nil {
			t.Fatal(err)
		}
		content := "\"hello\" = \"" + map[string]string{"en": "Hello", "sv": "Hej"}[lang] + "\";\n\"world\" = \"" + map[string]string{"en": "World", "sv": "Världen"}[lang] + "\";"
		if err := os.WriteFile(filepath.Join(langDir, "Localizable.strings"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	config := Config{
		InPath:    inputDir,
		OutPath:   outputDir,
		Delimiter: ",",
		ToCSV:     true,
	}
	processor := NewLocalizationProcessor(config)
	if err := processor.ConvertToCSV(); err != nil {
		t.Fatalf("ConvertToCSV failed: %v", err)
	}

	// Check generated CSV
	content, err := os.ReadFile(filepath.Join(outputDir, "output.csv"))
	if err != nil {
		t.Fatalf("failed to read output.csv: %v", err)
	}
	expected := "Key,en,sv\nhello,Hello,Hej\nworld,World,Världen\n"
	if string(content) != expected {
		t.Fatalf("unexpected CSV content:\nGot: %s\nWant: %s", content, expected)
	}

	if err := os.RemoveAll(inputDir); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(outputDir); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePath(t *testing.T) {
	tempDir := t.TempDir()
	if err := validatePath(tempDir); err != nil {
		t.Errorf("validatePath(%s) failed: %v", tempDir, err)
	}
	if err := validatePath(filepath.Join(tempDir, "nonexistent")); err == nil {
		t.Error("validatePath should fail for nonexistent path")
	}
}
