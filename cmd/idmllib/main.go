// Package main provides a CLI tool for working with InDesign IDML files.
// It allows listing stories and exporting them as IDMS snippets.
package main

import (
	"flag"
	"github.com/dimelords/idmllib/types"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dimelords/idmllib/idml"
	"github.com/dimelords/idmllib/idms"
)

func main() {
	// Configure slog with text handler for CLI
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Define command line flags
	idmlPath := flag.String("idml", "", "Path to IDML file (required)")
	listStories := flag.Bool("list", false, "List all stories in the IDML file")
	textFrameID := flag.String("textframe", "", "Textframe Self")
	outputPath := flag.String("output", "", "Output IDMS file path")

	flag.Parse()

	if *idmlPath == "" {
		slog.Error("IDML flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Verify IDML file exists
	if _, err := os.Stat(*idmlPath); os.IsNotExist(err) {
		slog.Error("IDML file not found", "path", *idmlPath)
		os.Exit(1)
	}

	// Open IDML package
	pkg, err := idml.Open(*idmlPath)
	if err != nil {
		slog.Error("Failed to open IDML file", "error", err, "path", *idmlPath)
		os.Exit(1)
	}
	defer func(pkg *idml.Package) {
		_ = pkg.Close()
	}(pkg)

	// Handle list command
	if *listStories {
		listAllStories(pkg)
		return
	}

	// Handle idms export command
	if *textFrameID != "" {
		if *outputPath == "" {
			slog.Error("Output flag is required when exporting a story")
			os.Exit(1)
		}
		exportIDMS(pkg, *textFrameID, *outputPath)
		return
	}

	// If no command specified, show usage
	slog.Error("No command specified. Use -list to list stories or -textframe to export a idms")
	flag.Usage()
}

// listAllStories lists all stories in the IDML with optional content preview
func listAllStories(pkg *idml.Package) {
	slog.Info("Stories found", "count", len(pkg.Stories))

	for i, story := range pkg.Stories {
		slog.Info("Story details",
			"number", i+1,
			"id", story.Self,
			"self", story.Self)
	}
}

// exportIDMS exports a specific story as an IDMS file
func exportIDMS(pkg *idml.Package, textFrameID, outputPath string) {
	// Ensure output path has .idms extension
	if !strings.HasSuffix(strings.ToLower(outputPath), ".idms") {
		outputPath += ".idms"
	}

	// Make output directory if needed
	outputDir := filepath.Dir(outputPath)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			slog.Error("Failed to create output directory", "error", err, "dir", outputDir)
			os.Exit(1)
		}
	}

	slog.Info("Exporting idms", "textFrame", textFrameID, "output", outputPath)

	// Create IDMS exporter with the IDML package as reader
	exporter := idms.NewExporter(pkg)

	// Export the textframe using the IDMS exporter
	predicate := func(tf *types.TextFrame) bool {
		return tf.Self == textFrameID
	}

	exportErr := exporter.ExportXML(outputPath, predicate)

	if exportErr != nil {
		slog.Error("Failed to export textframe", "error", exportErr, "textframeId", textFrameID)
		os.Exit(1)
	}

	slog.Info("Export successful", "file", outputPath)
}
