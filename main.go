package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// CodeBlock represents a structure that holds both the content and the language of a code block
type CodeBlock struct {
	Language string
	Content  string
}

// ExtractCodeBlocks extracts code blocks and their languages from the given markdown content
func ExtractCodeBlocks(markdown string) []CodeBlock {
	// Create a goldmark parser
	md := goldmark.New()

	// Convert markdown string to bytes
	source := []byte(markdown)

	// Parse the markdown into a document AST
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var codeBlocks []CodeBlock

	// Walk through the AST and find code blocks
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if codeBlock, ok := n.(*ast.FencedCodeBlock); ok && entering {
			var content bytes.Buffer

			// Loop through all the lines in the code block
			for i := 0; i < codeBlock.Lines().Len(); i++ {
				segment := codeBlock.Lines().At(i)
				content.Write(segment.Value(source))
			}

			// Extract the programming language (if specified)
			language := string(codeBlock.Language(source))

			// Store the complete code block content and the language
			codeBlocks = append(codeBlocks, CodeBlock{
				Language: language,
				Content:  content.String(),
			})
		}
		return ast.WalkContinue, nil
	})

	return codeBlocks
}

// GenerateCodeImage saves the code to a temporary file and uses the "silicon" binary to create a PNG image.
func GenerateCodeImage(code, language, filename string) error {
	// Create a temporary file to store the code block
	tempFile, err := os.CreateTemp("", "codeblock.*."+language)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	fmt.Printf("Created temp file: %s\n", tempFile.Name())
	// defer os.Remove(tempFile.Name()) // Clean up the temp file after it's used

	// Write the code to the temporary file
	if _, err := tempFile.Write([]byte(code)); err != nil {
		return fmt.Errorf("failed to write code to temp file: %v", err)
	}

	// Close the file to ensure data is flushed to disk
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	// Define the output image file path
	outputFilePath := filepath.Join("", filename)

	// Build the command to call the silicon binary
	cmd := exec.Command(
		"silicon", tempFile.Name(),
		"-o", outputFilePath,
		"--language", language,
	)

	// Run the silicon command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run silicon: %v", err)
	}

	fmt.Printf("Image saved as %s\n", outputFilePath)
	return nil
}

func main() {
	markdownFile := os.Args[1]
	filePrefix := os.Args[2]

	markdown, err := os.ReadFile(markdownFile)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	codeBlocks := ExtractCodeBlocks(string(markdown))
	for idx, block := range codeBlocks {
		err := GenerateCodeImage(
			block.Content,
			block.Language,
			fmt.Sprintf("%s-%s-%d.png", filePrefix, block.Language, idx),
		)
		if err != nil {
			fmt.Println("Error generating image:", err)
		}
	}
}
