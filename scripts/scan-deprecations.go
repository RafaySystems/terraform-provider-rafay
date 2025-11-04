package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Deprecation represents a detected deprecation in the code
type Deprecation struct {
	Resource string `json:"resource"`
	Field    string `json:"field,omitempty"`
	Message  string `json:"message"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// DeprecationResult holds all detected deprecations
type DeprecationResult struct {
	Deprecations []Deprecation `json:"deprecations"`
}

var (
	pathFlag    = flag.String("path", ".", "Path to scan for Go files")
	outputFlag  = flag.String("output", "deprecations.json", "Output JSON file path")
	verboseFlag = flag.Bool("verbose", false, "Verbose output")
)

func main() {
	flag.Parse()

	deprecations, err := scanDirectory(*pathFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	result := DeprecationResult{
		Deprecations: deprecations,
	}

	// Write JSON output
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputFlag, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	if *verboseFlag {
		fmt.Printf("Found %d deprecation(s)\n", len(deprecations))
		fmt.Printf("Output written to %s\n", *outputFlag)
	}

	// Also print to stdout for pipeline consumption
	fmt.Println(string(data))
}

func scanDirectory(root string) ([]Deprecation, error) {
	var deprecations []Deprecation

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files and test files
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip vendor and internal generated files
		if strings.Contains(path, "/vendor/") || strings.Contains(path, "_gen.go") {
			return nil
		}

		deps, err := scanFile(path)
		if err != nil {
			if *verboseFlag {
				fmt.Fprintf(os.Stderr, "Warning: Error scanning %s: %v\n", path, err)
			}
			return nil // Continue scanning other files
		}

		deprecations = append(deprecations, deps...)
		return nil
	})

	return deprecations, err
}

func scanFile(filename string) ([]Deprecation, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var deprecations []Deprecation
	resourceName := extractResourceName(filename)

	// Inspect the AST
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			// Look for resource/data source declarations
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Resource" {
					// Check for DeprecationMessage in resource definition
					if deps := findDeprecationMessage(x, resourceName, filename, fset); len(deps) > 0 {
						deprecations = append(deprecations, deps...)
					}
				}
			}

		case *ast.CompositeLit:
			// Look for schema definitions with Deprecated field
			if deps := findDeprecatedInSchema(x, resourceName, filename, fset); len(deps) > 0 {
				deprecations = append(deprecations, deps...)
			}
		}
		return true
	})

	return deprecations, nil
}

func findDeprecationMessage(call *ast.CallExpr, resourceName, filename string, fset *token.FileSet) []Deprecation {
	var deprecations []Deprecation

	// Look through the arguments for composite literals (struct initialization)
	for _, arg := range call.Args {
		if comp, ok := arg.(*ast.CompositeLit); ok {
			for _, elt := range comp.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok {
					if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == "DeprecationMessage" {
						if lit, ok := kv.Value.(*ast.BasicLit); ok {
							message := strings.Trim(lit.Value, `"`)
							pos := fset.Position(lit.Pos())
							deprecations = append(deprecations, Deprecation{
								Resource: resourceName,
								Message:  message,
								File:     filename,
								Line:     pos.Line,
							})
						}
					}
				}
			}
		}
	}

	return deprecations
}

func findDeprecatedInSchema(comp *ast.CompositeLit, resourceName, filename string, fset *token.FileSet) []Deprecation {
	var deprecations []Deprecation
	var fieldName string
	var deprecatedMessage string
	var line int

	// Check if this is a schema definition
	for _, elt := range comp.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				switch ident.Name {
				case "Deprecated":
					if lit, ok := kv.Value.(*ast.BasicLit); ok {
						deprecatedMessage = strings.Trim(lit.Value, `"`)
						pos := fset.Position(lit.Pos())
						line = pos.Line
					}
				}
			}
		}
	}

	// If we found a Deprecated field, try to find the field name from parent context
	if deprecatedMessage != "" {
		// Try to extract field name from nearby context
		fieldName = extractFieldNameFromContext(comp, fset)

		deprecations = append(deprecations, Deprecation{
			Resource: resourceName,
			Field:    fieldName,
			Message:  deprecatedMessage,
			File:     filename,
			Line:     line,
		})
	}

	return deprecations
}

func extractFieldNameFromContext(comp *ast.CompositeLit, fset *token.FileSet) string {
	// This is a heuristic to find the field name
	// In schema maps, the field name is usually the key in the parent map
	// This is a simplified version - a more robust implementation would track
	// the AST traversal path

	// Look for string literals in the composite literal that might be field names
	for _, elt := range comp.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			// Check for "Type" field to confirm this is a schema
			if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == "Type" {
				// This is likely a schema field, but we need the parent map key
				// For now, return empty and let the changelog generator handle it
				return ""
			}
		}
	}

	return ""
}

func extractResourceName(filename string) string {
	// Extract resource name from filename
	// Examples:
	// rafay/resource_eks_cluster.go -> rafay_eks_cluster
	// rafay/data_source_clusters.go -> rafay_clusters (data source)

	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, ".go")

	// Remove common prefixes
	base = strings.TrimPrefix(base, "resource_")
	base = strings.TrimPrefix(base, "data_source_")

	// Get package name (directory)
	dir := filepath.Base(filepath.Dir(filename))

	// Construct resource name
	if dir == "rafay" || dir == "provider" {
		return fmt.Sprintf("rafay_%s", base)
	}

	return base
}

// Additional helper function to scan for deprecation patterns in string literals
func scanForDeprecationPatterns(content string) []string {
	var patterns []string

	// Common deprecation patterns
	deprecationRegex := regexp.MustCompile(`(?i)(deprecated|deprecation).*?["\']([^"\']+)["\']`)
	matches := deprecationRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 2 {
			patterns = append(patterns, match[2])
		}
	}

	return patterns
}
