package idml

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/dimelords/idmllib/v3/internal/xmlutil"
)

func readFixtureStreaming(path string) (*Package, error) {
	return ReadWithOptions(path, &ReadOptions{
		MaxTotalSize:  2 << 30,
		MaxFileSize:   1 << 30,
		StreamingMode: true,
	})
}

// TestStreamingWriteMatchesNormalUnmodified checks that reading in streaming
// mode and writing without changes produces exactly the same archive as the
// normal path, for every fixture.
func TestStreamingWriteMatchesNormalUnmodified(t *testing.T) {
	for _, path := range fidelityFixtures(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			skipIfLarge(t, path)
			checkStreamingUnmodified(t, path)
		})
	}
}

// checkStreamingUnmodified reads path normally and in streaming mode, writes
// both back without changes and requires identical archives.
func checkStreamingUnmodified(t *testing.T, path string) {
	t.Helper()
	normal, err := readFixture(path)
	if err != nil {
		t.Fatal(err)
	}
	streamed, err := readFixtureStreaming(path)
	if err != nil {
		t.Fatal(err)
	}
	if !streamed.StreamingMode() {
		t.Error("StreamingMode() = false")
	}
	parseEverything(t, normal)
	parseEverything(t, streamed)

	dir := t.TempDir()
	normalOut := filepath.Join(dir, "normal.idml")
	streamedOut := filepath.Join(dir, "streamed.idml")
	if err := Write(normal, normalOut); err != nil {
		t.Fatal(err)
	}
	if err := Write(streamed, streamedOut); err != nil {
		t.Fatal(err)
	}
	a, b := readZipFiles(t, normalOut), readZipFiles(t, streamedOut)
	if len(a) != len(b) {
		t.Fatalf("file count differs: %d vs %d", len(a), len(b))
	}
	for name, want := range a {
		if !bytes.Equal(want, b[name]) {
			t.Errorf("%s differs between normal and streaming output (%d vs %d bytes)", name, len(want), len(b[name]))
		}
	}
}

// TestStreamingWriteMatchesNormalModified does the same after marking every
// spread modified, which forces a re-marshal and therefore exercises payload
// restoration.
func TestStreamingWriteMatchesNormalModified(t *testing.T) {
	for _, path := range fidelityFixtures(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			skipIfLarge(t, path)
			checkStreamingModified(t, path)
		})
	}
}

// checkStreamingModified reads path normally and in streaming mode, marks
// every spread modified so it is re-marshaled, and requires the written
// spreads to be identical.
func checkStreamingModified(t *testing.T, path string) {
	t.Helper()
	outputs := map[bool][]byte{}
	for _, streaming := range []bool{false, true} {
		var pkg *Package
		var err error
		if streaming {
			pkg, err = readFixtureStreaming(path)
		} else {
			pkg, err = readFixture(path)
		}
		if err != nil {
			t.Fatal(err)
		}
		spreads, err := pkg.Spreads()
		if err != nil {
			t.Fatal(err)
		}
		masters, err := pkg.MasterSpreads()
		if err != nil {
			t.Fatal(err)
		}
		for name := range spreads {
			pkg.MarkModified(name)
		}
		for name := range masters {
			pkg.MarkModified(name)
		}
		out := filepath.Join(t.TempDir(), "out.idml")
		if err := Write(pkg, out); err != nil {
			t.Fatal(err)
		}
		files := readZipFiles(t, out)
		names := make([]string, 0, len(files))
		for name := range files {
			if IsSpreadPath(name) || IsMasterSpreadPath(name) {
				names = append(names, name)
			}
		}
		sort.Strings(names) // map order is random; compare deterministically
		for _, name := range names {
			outputs[streaming] = append(outputs[streaming], files[name]...)
		}
	}
	if !bytes.Equal(outputs[false], outputs[true]) {
		t.Errorf("re-marshaled spreads differ between normal and streaming mode (%d vs %d bytes)",
			len(outputs[false]), len(outputs[true]))
	}
}

// TestStreamingKeepsPayloadsOutOfParsedSpread verifies the point of the mode:
// the payload is absent from the parsed document but still reachable and
// still written.
func TestStreamingKeepsPayloadsOutOfParsedSpread(t *testing.T) {
	fixture := requireBigFixture(t)
	pkg, err := readFixtureStreaming(fixture)
	if err != nil {
		t.Fatal(err)
	}
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatal(err)
	}
	var spreadPath string
	for name := range spreads {
		spreadPath = name
	}

	raw, err := pkg.FileData(spreadPath)
	if err != nil {
		t.Fatal(err)
	}
	_, payloads := xmlutil.StripContents(raw)
	if len(payloads) == 0 {
		t.Fatalf("%s has no embedded payloads; fixture assumption broken", spreadPath)
	}
	for self, want := range payloads {
		got, ok := pkg.ImageContents(spreadPath, self)
		if !ok {
			t.Errorf("ImageContents(%s, %s) not found", spreadPath, self)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("payload for %s differs (%d vs %d bytes)", self, len(got), len(want))
		}
	}
	if _, ok := pkg.ImageContents(spreadPath, "nosuchitem"); ok {
		t.Error("ImageContents reported an unknown item as present")
	}
}

// TestStreamingReducesMemory measures the actual benefit: holding a parsed
// image-heavy package must cost meaningfully less in streaming mode.
func TestStreamingReducesMemory(t *testing.T) {
	fixture := requireBigFixture(t)
	measure := func(streaming bool) uint64 {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		var pkg *Package
		var err error
		if streaming {
			pkg, err = readFixtureStreaming(fixture)
		} else {
			pkg, err = readFixture(fixture)
		}
		if err != nil {
			t.Fatal(err)
		}
		spreads, err := pkg.Spreads()
		if err != nil {
			t.Fatal(err)
		}
		runtime.GC()
		runtime.ReadMemStats(&after)
		runtime.KeepAlive(pkg)
		runtime.KeepAlive(spreads)
		if after.HeapAlloc < before.HeapAlloc {
			return 0
		}
		return after.HeapAlloc - before.HeapAlloc
	}
	normal := measure(false)
	streamed := measure(true)
	t.Logf("held heap: normal %.1f MB, streaming %.1f MB", float64(normal)/(1<<20), float64(streamed)/(1<<20))
	if streamed >= normal {
		t.Errorf("streaming mode did not reduce memory: %d vs %d bytes", streamed, normal)
	}
	// The payloads are ~121 MB of a ~122 MB package, so the saving should be
	// large. Require at least 30% to stay robust against allocator noise.
	if saved := float64(normal-streamed) / float64(normal); saved < 0.30 {
		t.Errorf("streaming saved only %.1f%% of held heap, want at least 30%%", saved*100)
	}
}

// epsSpreadItems are page items that each give one owner two <Contents>
// elements, or an empty one, the way InDesign writes them: an embedded EPS
// with its payload and the XMP packet of its MetadataPacketPreference, an
// image whose only Contents is its XMP packet, and an image with an empty
// Contents between items that have payloads.
const epsSpreadItems = `<Rectangle Self="ueps" ContentType="GraphicType">` +
	`<EPS Self="ueps1" Space="RGB"><Properties><Contents><![CDATA[RVBTIFBBWUxPQUQ=]]></Contents></Properties>` +
	`<MetadataPacketPreference><Properties><Contents><![CDATA[<?xpacket begin="" id="eps"?><x:xmpmeta xmlns:x="adobe:ns:meta/"/>]]></Contents></Properties></MetadataPacketPreference>` +
	`</EPS></Rectangle>` +
	`<Rectangle Self="uempty" ContentType="GraphicType"><Image Self="uempty1"><Properties><Contents></Contents></Properties></Image></Rectangle>` +
	`<Rectangle Self="uxmp" ContentType="GraphicType"><Image Self="uxmp1">` +
	`<MetadataPacketPreference><Properties><Contents><![CDATA[<?xpacket begin="" id="img"?><x:xmpmeta xmlns:x="adobe:ns:meta/"/>]]></Contents></Properties></MetadataPacketPreference>` +
	`</Image></Rectangle>`

// epsFixture writes a copy of plain.idml whose spread also holds
// epsSpreadItems, and returns its path and the spread's path in the archive.
func epsFixture(t *testing.T) (path, spreadPath string) {
	t.Helper()
	src, err := zip.OpenReader(filepath.Join("..", "..", "testdata", "plain.idml"))
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	path = filepath.Join(t.TempDir(), "eps.idml")
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	zw := zip.NewWriter(out)

	for _, f := range src.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		if IsSpreadPath(f.Name) && spreadPath == "" {
			i := bytes.LastIndex(data, []byte("</Spread>"))
			if i < 0 {
				t.Fatalf("%s has no </Spread>", f.Name)
			}
			data = append(append(append([]byte{}, data[:i]...), epsSpreadItems...), data[i:]...)
			spreadPath = f.Name
		}
		w, err := zw.CreateHeader(&zip.FileHeader{Name: f.Name, Method: f.Method, Modified: f.Modified})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if spreadPath == "" {
		t.Fatal("plain.idml has no spread")
	}
	return path, spreadPath
}

// TestStreamingHandlesSeveralContentsPerOwner covers a spread where an owner
// has a second <Contents> that has to stay inline, which used to drop the
// bytes around it and leave the spread unparsable in streaming mode.
func TestStreamingHandlesSeveralContentsPerOwner(t *testing.T) {
	path, spreadPath := epsFixture(t)

	t.Run("unmodified", func(t *testing.T) { checkStreamingUnmodified(t, path) })
	t.Run("modified", func(t *testing.T) { checkStreamingModified(t, path) })

	t.Run("payloads", func(t *testing.T) {
		pkg, err := readFixtureStreaming(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pkg.Spread(spreadPath); err != nil {
			t.Fatalf("Spread(%s) in streaming mode: %v", spreadPath, err)
		}
		got, ok := pkg.ImageContents(spreadPath, "ueps1")
		if want := "<![CDATA[RVBTIFBBWUxPQUQ=]]>"; !ok || string(got) != want {
			t.Errorf("ImageContents(ueps1) = %q, %v; want %q", got, ok, want)
		}
		got, ok = pkg.ImageContents(spreadPath, "uxmp1")
		if !ok || !bytes.Contains(got, []byte(`id="img"`)) {
			t.Errorf("ImageContents(uxmp1) = %q, %v; want the image's XMP packet", got, ok)
		}
		if _, ok := pkg.ImageContents(spreadPath, "uempty1"); ok {
			t.Error("ImageContents(uempty1) found a payload for an empty Contents")
		}
	})
}
