package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type LocalizationProcessor struct {
	config Config
}

func NewLocalizationProcessor(config Config) *LocalizationProcessor {
	return &LocalizationProcessor{config: config}
}

func (p *LocalizationProcessor) ConvertToIOS() error {
	if err := validatePath(p.config.InPath); err != nil {
		return fmt.Errorf("source path validation failed: %v", err)
	}
	if err := validatePath(p.config.OutPath); err != nil {
		return fmt.Errorf("target path validation failed: %v", err)
	}

	outputDir := p.config.OutPath

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %v", outputDir, err)
	}

	log.Printf("created output directory: %s", outputDir)
	return p.generateIOSKeys(p.config.InPath, outputDir)
}

func (p *LocalizationProcessor) generateIOSKeys(sourcePath, output string) error {
	var langKeys []string
	var allWrites []*os.File
	defer func() {
		for _, w := range allWrites {
			w.Close()
		}
	}()

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".csv" {
			return err
		}

		keys, writers, err := p.setupLanguageFiles(path, output)
		if err != nil {
			return err
		}
		if len(langKeys) == 0 {
			langKeys = keys
			allWrites = writers
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking directory: %v", err)
	}

	return p.startLocalizeIOS(sourcePath, allWrites)
}

func (p *LocalizationProcessor) setupLanguageFiles(csvPath, output string) ([]string, []*os.File, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, nil, fmt.Errorf("empty CSV file")
	}

	parts := strings.Split(scanner.Text(), p.config.Delimiter)
	langKeys := parts[1:]
	var allWrites []*os.File

	for _, lang := range langKeys {
		langPath := filepath.Join(output, fmt.Sprintf("%s.lproj", lang))
		if err := os.MkdirAll(langPath, 0755); err != nil {
			return nil, nil, err
		}
		outPath := filepath.Join(langPath, "Localizable.strings")
		f, err := os.Create(outPath)
		if err != nil {
			return nil, nil, err
		}
		allWrites = append(allWrites, f)
	}
	return langKeys, allWrites, nil
}

func (p *LocalizationProcessor) startLocalizeIOS(sourcePath string, allWrites []*os.File) error {
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".csv" {
			return err
		}

		log.Printf("Localizing: %s to iOS", info.Name())
		return p.processCSVToIOS(path, info.Name(), allWrites)
	})
}

func (p *LocalizationProcessor) processCSVToIOS(path, filename string, allWrites []*os.File) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan() // Skip header

	for _, w := range allWrites {
		fmt.Fprintf(w, "\n/* %s */\n", filename)
	}

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), p.config.Delimiter)
		key := strings.ReplaceAll(parts[0], " ", "")
		if strings.HasPrefix(key, "//") {
			continue
		}
		values := parts[1:]

		for i, w := range allWrites {
			fmt.Fprintf(w, "\"%s\" = \"%s\";\n", key, values[i])
		}
	}
	return scanner.Err()
}

func (p *LocalizationProcessor) ConvertToCSV() error {
	if err := validatePath(p.config.InPath); err != nil {
		return fmt.Errorf("source path validation failed: %v", err)
	}

	outputFile := filepath.Join(p.config.OutPath, "output.csv")
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	languages, languageOrder, err := p.readLocalizedFiles()
	if err != nil {
		return fmt.Errorf("error processing localized files: %v", err)
	}

	if err := p.writeCSV(writer, languages, languageOrder); err != nil {
		return fmt.Errorf("error writing CSV: %v", err)
	}

	log.Printf("generated CSV file: %s", outputFile)
	return nil
}

func (p *LocalizationProcessor) readLocalizedFiles() (map[string]map[string]string, []string, error) {
	languages := make(map[string]map[string]string)
	languageOrder := []string{}

	err := filepath.Walk(p.config.InPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(info.Name(), ".lproj") || !info.IsDir() {
			return err
		}

		lang := strings.TrimSuffix(info.Name(), ".lproj")
		languageOrder = append(languageOrder, lang)
		languages[lang] = make(map[string]string)

		stringsPath := filepath.Join(path, "Localizable.strings")
		if _, err := os.Stat(stringsPath); os.IsNotExist(err) {
			return nil
		}

		file, err := os.Open(stringsPath)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "//") {
				continue
			}
			parts := strings.SplitN(line, " = ", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.Trim(parts[0], `"`)
			value := strings.Trim(strings.TrimSuffix(parts[1], ";"), `"`)
			languages[lang][key] = value
		}
		return scanner.Err()
	})
	return languages, languageOrder, err
}

func (p *LocalizationProcessor) writeCSV(writer *bufio.Writer, languages map[string]map[string]string, languageOrder []string) error {
	// Write header
	header := append([]string{"Key"}, languageOrder...)
	fmt.Fprintln(writer, strings.Join(header, p.config.Delimiter))

	// Collect all unique keys
	keys := make(map[string]bool)
	for _, langMap := range languages {
		for key := range langMap {
			keys[key] = true
		}
	}

	var keySlice []string
	for k := range keys {
		keySlice = append(keySlice, k)
	}

	sort.Strings(keySlice)

	for _, key := range keySlice {
		row := []string{key}
		for _, lang := range languageOrder {
			value := languages[lang][key]
			row = append(row, value)
		}
		fmt.Fprintln(writer, strings.Join(row, p.config.Delimiter))
	}
	return nil
}

func validatePath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path not found: %s", path)
	}
	return nil
}
