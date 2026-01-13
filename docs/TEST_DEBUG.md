# Test Debugging Guide

This document explains how to use debug flags to preserve test output files for debugging purposes.

## Debug Flags

### `-preserve-test-output`

When running tests with the `-preserve-test-output` flag, test output files will be preserved in a `debug_test_output/` directory instead of being automatically cleaned up.

```bash
# Run tests with debug output preservation
go test -preserve-test-output ./pkg/idml

# Run specific test with debug output
go test -preserve-test-output -run TestSpecificTest ./pkg/idml

# Run all tests with debug output
go test -preserve-test-output ./...
```

## Debug Output Directory

When debug mode is enabled, test artifacts are saved to:

```
debug_test_output/
├── TestName_output.idml
├── TestName_roundtrip.idms
└── TestName_test.idml
```

The files are named with the pattern: `{TestName}_{filename}`

## Using Debug Functions

### For IDML Files

```go
// Regular temporary file (auto-cleanup)
outputPath := testutil.TempIDML(t, "output.idml")

// Debug-enabled file (preserved if flag is set)
outputPath := testutil.TempIDMLWithDebug(t, "output.idml")

// Write IDML with debug support
outputPath := writeTestIDMLWithDebug(t, pkg, "debug_output.idml")
```

### For ZIP Files

```go
// Regular ZIP file (auto-cleanup)
zipPath := testutil.CreateTestZIP(t, files)

// Debug-enabled ZIP file (preserved if flag is set)
zipPath := testutil.CreateTestZIPWithDebug(t, files, "debug.idml")
```

## Cleanup Behavior

| Scenario | Regular Mode | Debug Mode |
|----------|-------------|------------|
| Test passes | Files cleaned up | Files preserved if `-preserve-test-output` |
| Test fails | Files cleaned up | Files always preserved |
| Manual cleanup | Not needed | Remove `debug_test_output/` directory |

## Best Practices

1. **Use debug functions sparingly**: Only use debug-enabled functions when you need to inspect test output
2. **Clean up debug directory**: Periodically remove the `debug_test_output/` directory
3. **Don't commit debug files**: The `debug_test_output/` directory is in `.gitignore`
4. **Use descriptive names**: Give debug files descriptive names to identify their purpose

## Examples

### Debugging a Roundtrip Test

```go
func TestMyRoundtrip(t *testing.T) {
    pkg := loadTestIDML(t, "example.idml")
    
    // Write with debug preservation
    outputPath := writeTestIDMLWithDebug(t, pkg, "roundtrip_output.idml")
    
    // Read back
    pkg2, err := Read(outputPath)
    if err != nil {
        t.Fatalf("Failed to read back: %v", err)
    }
    
    // Compare...
}
```

### Debugging XML Generation

```go
func TestXMLGeneration(t *testing.T) {
    files := map[string][]byte{
        "designmap.xml": generateDesignmap(),
        "Stories/Story_u1d8.xml": generateStory(),
    }
    
    // Create debug ZIP
    zipPath := testutil.CreateTestZIPWithDebug(t, files, "generated.idml")
    
    // Test the generated ZIP...
}
```

## Troubleshooting

### Debug Files Not Preserved

Make sure you're using the `-preserve-test-output` flag:

```bash
go test -preserve-test-output ./pkg/idml
```

### Debug Directory Not Created

The debug directory is created automatically when needed. If it's not appearing, check:

1. You're using debug-enabled functions (`TempIDMLWithDebug`, etc.)
2. The `-preserve-test-output` flag is set
3. You have write permissions in the current directory

### Too Many Debug Files

Clean up the debug directory periodically:

```bash
rm -rf debug_test_output/
```

Or add it to your test cleanup script.