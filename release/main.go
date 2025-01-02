package main

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/raymondji/commitstack/libgit"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <tag>")
		os.Exit(1)
	}

	tag := os.Args[1]
	fmt.Printf("Setting current version: %s\n", tag)
	err := generateGoFile(tag)
	if err != nil {
		fmt.Printf("Error generating Go file: %v\n", err)
		os.Exit(1)
	}

	git, err := libgit.New()
	if err != nil {
		fmt.Printf("error getting git: %v\n", err)
		os.Exit(1)
	}

	if err := git.Add(); err != nil {
		fmt.Printf("Error adding changes: %v\n", err)
		os.Exit(1)
	}
	if err := git.Commit(fmt.Sprintf("Releasing version %s", tag)); err != nil {
		fmt.Printf("Error committing changes: %v\n", err)
		os.Exit(1)
	}
	if err := git.TagForce(tag); err != nil {
		fmt.Printf("Error tagging commit: %v\n", err)
		os.Exit(1)
	}
	if err := git.PushTag(tag); err != nil {
		fmt.Printf("Error tagging commit: %v\n", err)
		os.Exit(1)
	}
	if err := git.TagForce("newest"); err != nil {
		fmt.Printf("Error tagging commit: %v\n", err)
		os.Exit(1)
	}
	if err := git.PushTag("newest"); err != nil {
		fmt.Printf("Error tagging commit: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done")
}

func generateGoFile(tag string) error {
	// Load the template file
	tmplPath := "release/releasevars/vars.go.tmpl"
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		Version string
	}{
		Version: tag,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	outputPath := "release/releasevars/vars.go"
	err = os.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
