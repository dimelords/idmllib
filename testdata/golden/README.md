# Golden Files

This directory contains "golden" files used for regression testing.

## What Are Golden Files?

Golden files are known-good test outputs that we compare against during testing. If a test output doesn't match its golden file, the test fails.

## Directory Structure

```
golden/
├── README.md                    # This file
├── plain_idml_roundtrip.golden  # Expected output for plain.idml roundtrip
└── example_idml_roundtrip.golden # Expected output for example.idml roundtrip
```

## How Golden Files Work

1. **First Run (Creating Golden Files):**
   ```bash
   go test ./pkg/idml/... -run TestGolden -update
   ```
   This creates/updates golden files with current output.

2. **Subsequent Runs (Comparing Against Golden):**
   ```bash
   go test ./pkg/idml/... -run TestGolden
   ```
   This compares current output against golden files.

3. **If Test Fails:**
   - Review the diff shown in test output
   - If output is correct, update golden: `go test -update`
   - If output is wrong, fix the code

## When to Update Golden Files

✅ **DO update when:**
- You've intentionally changed the output format
- You've fixed a bug that changes output
- You've verified the new output is correct

❌ **DON'T update when:**
- Tests are failing (fix the code first!)
- You haven't reviewed the changes
- You're unsure if the new output is correct

## Best Practices

1. **Review Before Updating**
   Always look at the diff before running `-update`

2. **Commit Golden Files**
   Golden files should be checked into version control

3. **Document Changes**
   When updating goldens, document why in your commit message

4. **Small Golden Files**
   Keep golden files reasonably sized for readability

## Example Usage

```go
golden := testutil.NewGoldenFile(t, "testdata/golden")

// Compare output against golden file
golden.Assert(t, "my_test", actualOutput)

// Update golden file (run with -update flag)
golden.Update(t, "my_test", actualOutput)
```

## Troubleshooting

### Golden File Doesn't Exist
Run the test with `-update` flag to create it:
```bash
go test ./pkg/idml/... -run TestGolden -update
```

### Golden File Out of Date
If intentional, update it:
```bash
go test ./pkg/idml/... -run TestGolden -update
```

### Want to See What Changed
Run with verbose flag:
```bash
go test ./pkg/idml/... -run TestGolden -v
```

---

**Note:** Golden files are binary files (IDML/ZIP format). Use `unzip -l` to inspect their contents if needed.
