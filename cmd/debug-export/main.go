package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dimelords/idmllib/pkg/idml"
	"github.com/dimelords/idmllib/pkg/idms"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: debug-export <idml-file> [text-frame-id]")
		fmt.Println("\nExample:")
		fmt.Println("  debug-export example.idml ue3")
		fmt.Println("\nIf no text-frame-id is provided, will list all available text frames")
		os.Exit(1)
	}

	idmlPath := os.Args[1]

	fmt.Printf("📖 Loading IDML file: %s\n", idmlPath)
	pkg, err := idml.Read(idmlPath)
	if err != nil {
		log.Fatalf("❌ Failed to load IDML: %v", err)
	}
	fmt.Println("✅ IDML loaded successfully")

	// Load all spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		log.Fatalf("❌ Failed to load spreads: %v", err)
	}

	// If no text frame ID provided, list all available text frames
	if len(os.Args) < 3 {
		fmt.Println("\n📋 Available TextFrames:")
		fmt.Println("   ID          | ParentStory | Content Preview")
		fmt.Println("   " + strings.Repeat("-", 80))

		for spreadFile, spread := range spreads {
			for _, tf := range spread.InnerSpread.TextFrames {
				preview := "[no story]"
				if tf.ParentStory != "" {
					storyFile := "Stories/Story_" + tf.ParentStory + ".xml"
					story, err := pkg.Story(storyFile)
					if err == nil {
						// Get first few characters of content
						if len(story.StoryElement.ParagraphStyleRanges) > 0 {
							if len(story.StoryElement.ParagraphStyleRanges[0].CharacterStyleRanges) > 0 {
								children := story.StoryElement.ParagraphStyleRanges[0].CharacterStyleRanges[0].Children
								// Concatenate all text from Content elements
								var textBuilder strings.Builder
								for _, child := range children {
									if child.Content != nil {
										textBuilder.WriteString(child.Content.Text)
									}
								}
								text := textBuilder.String()
								if len(text) > 30 {
									preview = text[:30] + "..."
								} else {
									preview = text
								}
							}
						}
					} else {
						preview = fmt.Sprintf("[error: %v]", err)
					}
				}

				fmt.Printf("   %-12s| %-12s| %s (in %s)\n",
					tf.Self,
					tf.ParentStory,
					preview,
					spreadFile)
			}
		}

		fmt.Println("\n💡 Run with a text frame ID to debug export:")
		fmt.Println("   debug-export " + idmlPath + " <text-frame-id>")
		return
	}

	frameID := os.Args[2]
	fmt.Printf("\n🔍 Debugging export for TextFrame: %s\n", frameID)

	// Find the text frame
	tf, err := pkg.SelectTextFrameByID(frameID)
	if err != nil {
		log.Fatalf("❌ Failed to find text frame: %v", err)
	}

	fmt.Println("\n📝 TextFrame Details:")
	fmt.Printf("   Self: %s\n", tf.Self)
	fmt.Printf("   ParentStory: %s\n", tf.ParentStory)
	fmt.Printf("   AppliedObjectStyle: %s\n", tf.AppliedObjectStyle)
	fmt.Printf("   ItemLayer: %s\n", tf.ItemLayer)

	// Create selection
	sel := idml.NewSelection()
	sel.AddTextFrame(tf)

	fmt.Println("\n✅ Selection created with 1 text frame")

	// Create exporter
	exporter := idms.NewExporter(pkg)
	fmt.Println("✅ Exporter created")

	// Export (this will analyze dependencies)
	fmt.Println("\n🔄 Starting export...")
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		log.Fatalf("❌ Export failed: %v", err)
	}

	// Get dependencies
	deps := exporter.Dependencies()

	fmt.Println("\n📊 Dependency Analysis:")
	fmt.Printf("   Stories: %d\n", len(deps.Stories))
	if len(deps.Stories) > 0 {
		fmt.Println("   Story files:")
		for storyFile := range deps.Stories {
			fmt.Printf("     - %s\n", storyFile)

			// Try to load the story to verify it exists
			story, err := pkg.Story(storyFile)
			if err != nil {
				fmt.Printf("       ⚠️  ERROR loading story: %v\n", err)
			} else {
				fmt.Printf("       ✅ Story loaded: %d paragraph ranges\n", len(story.StoryElement.ParagraphStyleRanges))
			}
		}
	}

	fmt.Printf("   ParagraphStyles: %d\n", len(deps.ParagraphStyles))
	fmt.Printf("   CharacterStyles: %d\n", len(deps.CharacterStyles))
	fmt.Printf("   ObjectStyles: %d\n", len(deps.ObjectStyles))
	fmt.Printf("   Colors: %d\n", len(deps.Colors))
	fmt.Printf("   Swatches: %d\n", len(deps.Swatches))
	fmt.Printf("   Layers: %d\n", len(deps.Layers))

	// Check result document
	fmt.Println("\n📦 Export Result:")
	fmt.Printf("   SnippetType: %s\n", result.SnippetType())
	fmt.Printf("   InlineSpreads: %d\n", len(result.Document.InlineSpreads))

	if len(result.Document.InlineSpreads) > 0 {
		spread := result.Document.InlineSpreads[0]
		fmt.Printf("   TextFrames in spread: %d\n", len(spread.TextFrames))
		if len(spread.TextFrames) > 0 {
			fmt.Printf("     First frame ID: %s\n", spread.TextFrames[0].Self)
			fmt.Printf("     First frame ParentStory: %s\n", spread.TextFrames[0].ParentStory)
		}
	}

	fmt.Printf("   InlineStories: %d\n", len(result.Document.InlineStories))

	if len(result.Document.InlineStories) > 0 {
		fmt.Println("   Stories included:")
		for _, story := range result.Document.InlineStories {
			content := "[no content]"
			if len(story.ParagraphStyleRanges) > 0 {
				if len(story.ParagraphStyleRanges[0].CharacterStyleRanges) > 0 {
					children := story.ParagraphStyleRanges[0].CharacterStyleRanges[0].Children
					// Concatenate all text from Content elements
					var textBuilder strings.Builder
					for _, child := range children {
						if child.Content != nil {
							textBuilder.WriteString(child.Content.Text)
						}
					}
					text := textBuilder.String()
					if len(text) > 50 {
						content = text[:50] + "..."
					} else {
						content = text
					}
				}
			}
			fmt.Printf("     - %s: %s\n", story.Self, content)
		}
	} else {
		fmt.Println("   ⚠️  WARNING: No stories in export!")
		fmt.Println("\n🔍 Debugging why stories are missing:")

		// Check if dependencies were collected
		if len(deps.Stories) == 0 {
			fmt.Println("   ❌ No story dependencies were collected during analysis")
			if tf.ParentStory == "" {
				fmt.Println("   💡 ROOT CAUSE: TextFrame has no ParentStory reference!")
			} else {
				fmt.Println("   💡 ROOT CAUSE: Dependency tracking failed despite TextFrame having ParentStory")
			}
		} else {
			fmt.Println("   ✅ Story dependencies were collected")
			fmt.Println("   ❌ But stories were not added to the document")
			fmt.Println("   💡 ROOT CAUSE: Story extraction or document building failed")
		}
	}

	// Marshal to see final XML
	data, err := idms.Marshal(result)
	if err != nil {
		log.Fatalf("❌ Marshal failed: %v", err)
	}

	outputFile := "debug-output.idms"
	if err := os.WriteFile(outputFile, data, 0600); err != nil {
		log.Fatalf("❌ Failed to write output: %v", err)
	}

	fmt.Printf("\n💾 Exported to: %s\n", outputFile)
	fmt.Printf("   Size: %d bytes\n", len(data))

	// Show a preview of the XML
	preview := string(data)
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	fmt.Println("\n📄 XML Preview:")
	fmt.Println(preview)

	// Detailed JSON output of key parts
	fmt.Println("\n🔬 Document.InlineStories structure:")
	if len(result.Document.InlineStories) > 0 {
		jsonData, err := json.MarshalIndent(result.Document.InlineStories, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal: %v", err)
		} else {
			jsonStr := string(jsonData)
			if len(jsonStr) > 1000 {
				jsonStr = jsonStr[:1000] + "\n... (truncated)"
			}
			fmt.Println(jsonStr)
		}
	} else {
		fmt.Println("   (empty array)")
	}
}
