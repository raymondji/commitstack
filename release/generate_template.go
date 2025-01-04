package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"
)

var (
	tmpl                 = flag.String("template", "", "Which template to generate")
	version              = flag.String("version", "", "Version of the CLI (required)")
	sampleOutputFilePath = flag.String("sample-output", "", "Path to the sample output file (required)")
)

const (
	templateVersion = "version"
	templateReadme  = "readme"
)

func main() {
	flag.Parse()

	if *tmpl == "" {
		log.Fatal("Error: --template is required")
	}

	switch *tmpl {
	case templateVersion:
		if *version == "" {
			log.Fatal("Error: --version is required")
		}

		if err := generateTemplate("version/version.go.tmpl", "version/version.go", map[string]string{
			"Version": *version,
		}); err != nil {
			log.Fatalf("Failed to generate version file: %v", err)
		}

		log.Println("Generated version/version.go")
	case templateReadme:
		if *version == "" {
			log.Fatal("Error: --version is required")
		}
		if *sampleOutputFilePath == "" {
			log.Fatal("Error: --sample-output is required")
		}

		sampleOutputString, err := readFileContents(*sampleOutputFilePath)
		if err != nil {
			log.Fatalf("Failed to read sample output file: %v", err)
		}

		if err := generateTemplate("README.md.tmpl", "README.md", map[string]string{
			"Version":      *version,
			"SampleOutput": sampleOutputString,
		}); err != nil {
			log.Fatalf("Failed to generate README file: %v", err)
		}

		log.Println("Generated README.md")
	default:
		log.Fatal("Error: invalid template", *tmpl)
	}
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
