package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/raymondji/git-stack-cli/version"
)

var (
	sampleOutputFilePath = flag.String("sample-output", "", "Path to the sample output file (required)")
)

func main() {
	flag.Parse()

	if *sampleOutputFilePath == "" {
		log.Fatal("Error: --sample-output is required")
	}

	sampleOutputString, err := readFileContents(*sampleOutputFilePath)
	if err != nil {
		log.Fatalf("Failed to read sample output file: %v", err)
	}

	if err := generateTemplate("README.md.tmpl", "readme/README.md", map[string]string{
		"Version":      version.Version,
		"SampleOutput": sampleOutputString,
	}); err != nil {
		log.Fatalf("Failed to generate README file: %v", err)
	}

	log.Println("Generated README.md")
}

func readFileContents(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("could not read file %s: %w", filePath, err)
	}
	return string(content), nil
}

func generateTemplate(tmplPath string, outputPath string, vars map[string]string) error {
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v, err: %v", tmplPath, err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v, err: %v", outputPath, err)
	}
	defer outFile.Close()

	err = tmpl.Execute(outFile, vars)
	if err != nil {
		return fmt.Errorf("failed to execute template, path: %s, err: %v", tmplPath, err)
	}
	return nil
}
