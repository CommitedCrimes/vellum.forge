package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

func main() {
	// Get the github-dark style
	style := styles.Get("github-dark")
	if style == nil {
		fmt.Fprintf(os.Stderr, "Style 'github-dark' not found\n")
		os.Exit(1)
	}

	// Create HTML formatter with classes
	formatter := html.New(html.WithClasses(true), html.WithLineNumbers(true))

	// Generate CSS
	css, err := formatter.Format(style, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating CSS: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	err = os.WriteFile("assets/static/css/chroma.css", css, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSS file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Chroma CSS generated successfully!")
}
