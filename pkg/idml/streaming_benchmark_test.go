package idml

import (
	"os"
	"testing"
)

// Streaming mode strips embedded image payloads while parsing a spread, so it
// changes nothing for a package that is only read: Read stores the raw file
// bytes either way. The benchmarks below therefore read and then parse, which
// is where the difference shows. BenchmarkReadOnly_Big documents that reading
// alone is unaffected.

// bigOpts raises the size limits for the large fixture, whose single spread
// exceeds the default per-file cap.
func bigOpts(streaming bool) *ReadOptions {
	return &ReadOptions{
		MaxTotalSize:  2 << 30,
		MaxFileSize:   1 << 30,
		StreamingMode: streaming,
	}
}

// readAndParse reads a package and parses every spread, which is the point at
// which embedded image payloads would otherwise be materialized a second time.
func readAndParse(b *testing.B, path string, opts *ReadOptions) {
	b.Helper()
	if _, err := os.Stat(path); err != nil {
		b.Skipf("fixture not present: %v", err)
	}
	pkg, err := ReadWithOptions(path, opts)
	if err != nil {
		b.Fatalf("ReadWithOptions failed: %v", err)
	}
	if _, err := pkg.Spreads(); err != nil {
		b.Fatalf("Spreads failed: %v", err)
	}
}

// BenchmarkReadNormal_Example reads and parses example.idml in normal mode.
func BenchmarkReadNormal_Example(b *testing.B) {
	for b.Loop() {
		readAndParse(b, "../../testdata/example.idml", &ReadOptions{})
	}
}

// BenchmarkReadStreaming_Example reads and parses example.idml in StreamingMode.
func BenchmarkReadStreaming_Example(b *testing.B) {
	for b.Loop() {
		readAndParse(b, "../../testdata/example.idml", &ReadOptions{StreamingMode: true})
	}
}

// BenchmarkReadNormal_Tripple reads and parses tripple.idml (3 spreads) in normal mode.
func BenchmarkReadNormal_Tripple(b *testing.B) {
	for b.Loop() {
		readAndParse(b, "../../testdata/tripple.idml", &ReadOptions{})
	}
}

// BenchmarkReadStreaming_Tripple reads and parses tripple.idml in StreamingMode.
func BenchmarkReadStreaming_Tripple(b *testing.B) {
	for b.Loop() {
		readAndParse(b, "../../testdata/tripple.idml", &ReadOptions{StreamingMode: true})
	}
}

// BenchmarkReadNormal_Big reads and parses example_big.idml, whose single
// spread is 121 MB of embedded image payload, in normal mode.
func BenchmarkReadNormal_Big(b *testing.B) {
	for b.Loop() {
		readAndParse(b, "../../testdata/example_big.idml", bigOpts(false))
	}
}

// BenchmarkReadStreaming_Big reads and parses example_big.idml in StreamingMode.
func BenchmarkReadStreaming_Big(b *testing.B) {
	for b.Loop() {
		readAndParse(b, "../../testdata/example_big.idml", bigOpts(true))
	}
}

// BenchmarkReadOnly_Big shows that reading without parsing costs the same in
// both modes, since the raw bytes are stored either way.
func BenchmarkReadOnly_Big(b *testing.B) {
	for _, streaming := range []bool{false, true} {
		name := "normal"
		if streaming {
			name = "streaming"
		}
		b.Run(name, func(b *testing.B) {
			if _, err := os.Stat("../../testdata/example_big.idml"); err != nil {
				b.Skipf("fixture not present: %v", err)
			}
			for b.Loop() {
				if _, err := ReadWithOptions("../../testdata/example_big.idml", bigOpts(streaming)); err != nil {
					b.Fatalf("ReadWithOptions failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkWriteUntouched_Big compares writing a large document back after
// reading it. A lazily read package copies untouched entries in their
// compressed form, so nothing is decompressed or recompressed.
func BenchmarkWriteUntouched_Big(b *testing.B) {
	for _, lazy := range []bool{false, true} {
		name := "eager"
		if lazy {
			name = "lazy"
		}
		b.Run(name, func(b *testing.B) {
			if _, err := os.Stat("../../testdata/example_big.idml"); err != nil {
				b.Skipf("fixture not present: %v", err)
			}
			out := b.TempDir() + "/out.idml"
			for b.Loop() {
				opts := bigOpts(false)
				opts.Lazy = lazy
				pkg, err := ReadWithOptions("../../testdata/example_big.idml", opts)
				if err != nil {
					b.Fatalf("read: %v", err)
				}
				if err := Write(pkg, out); err != nil {
					b.Fatalf("write: %v", err)
				}
				_ = pkg.Close()
			}
		})
	}
}
