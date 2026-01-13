package idml

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"

	"github.com/dimelords/idmllib/pkg/common"
)

// Default limits for ZIP bomb protection
const (
	// DefaultMaxTotalSize is the maximum total uncompressed size (500 MB)
	DefaultMaxTotalSize int64 = 500 * 1024 * 1024

	// DefaultMaxFileSize is the maximum size of a single file (100 MB)
	DefaultMaxFileSize int64 = 100 * 1024 * 1024

	// DefaultMaxFileCount is the maximum number of files in the archive
	DefaultMaxFileCount int = 10000

	// DefaultMaxCompressionRatio is the maximum compression ratio allowed
	// A ratio of 100 means uncompressed size can be at most 100x compressed size
	DefaultMaxCompressionRatio int64 = 100
)

// ReadOptions configures the Read operation with optional limits.
type ReadOptions struct {
	// MaxTotalSize limits the total uncompressed size of all files.
	// Set to 0 to use DefaultMaxTotalSize, -1 for no limit.
	MaxTotalSize int64

	// MaxFileSize limits the size of individual files.
	// Set to 0 to use DefaultMaxFileSize, -1 for no limit.
	MaxFileSize int64

	// MaxFileCount limits the number of files in the archive.
	// Set to 0 to use DefaultMaxFileCount, -1 for no limit.
	MaxFileCount int

	// MaxCompressionRatio limits the compression ratio (uncompressed/compressed).
	// Set to 0 to use DefaultMaxCompressionRatio, -1 for no limit.
	MaxCompressionRatio int64
}

// applyDefaults fills in default values for zero-value options.
func (opts *ReadOptions) applyDefaults() {
	if opts.MaxTotalSize == 0 {
		opts.MaxTotalSize = DefaultMaxTotalSize
	}
	if opts.MaxFileSize == 0 {
		opts.MaxFileSize = DefaultMaxFileSize
	}
	if opts.MaxFileCount == 0 {
		opts.MaxFileCount = DefaultMaxFileCount
	}
	if opts.MaxCompressionRatio == 0 {
		opts.MaxCompressionRatio = DefaultMaxCompressionRatio
	}
}

// isValidZipPath checks if a ZIP entry path is safe to extract.
// It rejects absolute paths and paths that attempt directory traversal.
func isValidZipPath(name string) bool {
	// Reject empty paths
	if name == "" {
		return false
	}

	// Reject absolute paths
	if filepath.IsAbs(name) {
		return false
	}

	// Clean the path and check for traversal attempts
	cleaned := filepath.Clean(name)

	// Reject if cleaned path starts with ".."
	if strings.HasPrefix(cleaned, "..") {
		return false
	}

	// Reject if path contains ".." components (even in middle)
	for _, part := range strings.Split(cleaned, string(filepath.Separator)) {
		if part == ".." {
			return false
		}
	}

	return true
}

// extractPackage is the shared logic for all Read* functions.
// It processes a slice of ZIP files and creates a Package with safety checks.
func extractPackage(files []*zip.File, opts *ReadOptions, source string) (*Package, error) {
	// Check file count limit
	if err := validateFileCount(files, opts, source); err != nil {
		return nil, err
	}

	pkg := New()
	var totalSize int64

	// Read each file in the archive
	for _, f := range files {
		if err := validateZipFile(f, opts, &totalSize, source); err != nil {
			return nil, err
		}

		data, err := extractZipFileData(f, opts, source)
		if err != nil {
			return nil, err
		}

		storeFileInPackage(pkg, f, data)
	}

	// Automatically parse designmap.xml for structured access
	if err := validateDesignMap(pkg); err != nil {
		return nil, err
	}

	return pkg, nil
}

// validateFileCount checks if the number of files exceeds the limit.
func validateFileCount(files []*zip.File, opts *ReadOptions, source string) error {
	if opts.MaxFileCount > 0 && len(files) > opts.MaxFileCount {
		return common.WrapErrorWithPath("idml", "read", source, fmt.Errorf("archive contains %d files, exceeds limit of %d", len(files), opts.MaxFileCount))
	}
	return nil
}

// validateZipFile performs security validation on a ZIP file entry.
func validateZipFile(f *zip.File, opts *ReadOptions, totalSize *int64, source string) error {
	// Validate path to prevent directory traversal attacks
	if !isValidZipPath(f.Name) {
		return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("invalid path: potential directory traversal"))
	}

	// Check individual file size limit
	if err := validateFileSize(f, opts); err != nil {
		return err
	}

	// Check compression ratio to detect ZIP bombs
	if err := validateCompressionRatio(f, opts); err != nil {
		return err
	}

	// Track total size with overflow protection
	if f.UncompressedSize64 > math.MaxInt64 {
		return common.WrapErrorWithPath("idml", "read", source, fmt.Errorf("file size %d bytes exceeds maximum supported size", f.UncompressedSize64))
	}
	// #nosec G115 - Conversion is safe: overflow check performed on line above
	uncompressedSize := int64(f.UncompressedSize64)
	*totalSize += uncompressedSize
	if opts.MaxTotalSize > 0 && *totalSize > opts.MaxTotalSize {
		return common.WrapErrorWithPath("idml", "read", source, fmt.Errorf("total uncompressed size exceeds limit of %d bytes", opts.MaxTotalSize))
	}

	return nil
}

// validateFileSize checks if individual file size exceeds limits.
func validateFileSize(f *zip.File, opts *ReadOptions) error {
	if f.UncompressedSize64 > math.MaxInt64 {
		return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("file size %d bytes exceeds maximum supported size", f.UncompressedSize64))
	}
	// #nosec G115 - Conversion is safe: overflow check performed on line above
	uncompressedSize := int64(f.UncompressedSize64)
	if opts.MaxFileSize > 0 && uncompressedSize > opts.MaxFileSize {
		return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("file size %d bytes exceeds limit of %d bytes", f.UncompressedSize64, opts.MaxFileSize))
	}
	return nil
}

// validateCompressionRatio checks for potential ZIP bombs based on compression ratio.
func validateCompressionRatio(f *zip.File, opts *ReadOptions) error {
	if opts.MaxCompressionRatio > 0 && f.CompressedSize64 > 0 && f.Method != zip.Store {
		if f.UncompressedSize64 > math.MaxInt64 || f.CompressedSize64 > math.MaxInt64 {
			return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("file sizes exceed maximum supported size"))
		}
		// #nosec G115 - Conversions are safe: overflow checks performed on line above
		uncompressedSize := int64(f.UncompressedSize64)
		compressedSize := int64(f.CompressedSize64) // #nosec G115 - Conversion is safe: overflow check performed above
		ratio := uncompressedSize / compressedSize
		if ratio > opts.MaxCompressionRatio {
			return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("compression ratio %d exceeds limit of %d (potential ZIP bomb)", ratio, opts.MaxCompressionRatio))
		}
	}
	return nil
}

// extractZipFileData safely extracts data from a ZIP file entry.
func extractZipFileData(f *zip.File, opts *ReadOptions, source string) ([]byte, error) {
	// Add validation for parameters
	if f == nil {
		return nil, common.Errorf("idml", "extract zip file data", "", "zip file entry is nil")
	}

	if opts == nil {
		return nil, common.Errorf("idml", "extract zip file data", "", "read options is nil")
	}

	if source == "" {
		source = "<unknown>"
	}

	// Open the file within the ZIP
	rc, err := f.Open()
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "read", source+"/"+f.Name, err)
	}
	defer func() {
		if closeErr := rc.Close(); closeErr != nil {
			// Log close error but don't override the main error
		}
	}()

	// Use LimitReader to prevent reading more than declared size + small buffer
	maxRead := calculateMaxReadSize(f, opts)
	limitedReader := io.LimitReader(rc, maxRead)

	// Read the entire file content
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "read", source+"/"+f.Name, err)
	}

	// Verify actual size doesn't significantly exceed declared size
	if err := validateActualFileSize(f, data); err != nil {
		return nil, err
	}

	return data, nil
}

// calculateMaxReadSize determines the maximum safe read size for a file.
func calculateMaxReadSize(f *zip.File, opts *ReadOptions) int64 {
	// Add validation for parameters
	if f == nil {
		return 1024 // Default small size for safety
	}

	if opts == nil {
		if f.UncompressedSize64 > math.MaxInt64-1024 {
			return math.MaxInt64
		}
		// #nosec G115 - Conversion is safe: overflow check performed on line above
		uncompressedSize := int64(f.UncompressedSize64)
		return uncompressedSize + 1024
	}

	if f.UncompressedSize64 > math.MaxInt64-1024 {
		maxRead := int64(math.MaxInt64)
		if opts.MaxFileSize > 0 && maxRead > opts.MaxFileSize {
			maxRead = opts.MaxFileSize + 1
		}
		return maxRead
	}
	// #nosec G115 - Conversion is safe: overflow check performed above (line 258)
	uncompressedSize := int64(f.UncompressedSize64)
	maxRead := uncompressedSize + 1024 // Allow 1KB slack for edge cases
	if opts.MaxFileSize > 0 && maxRead > opts.MaxFileSize {
		maxRead = opts.MaxFileSize + 1
	}
	return maxRead
}

// validateActualFileSize verifies the actual read size matches expectations.
func validateActualFileSize(f *zip.File, data []byte) error {
	// Add validation for parameters
	if f == nil {
		return common.Errorf("idml", "validate file size", "", "zip file entry is nil")
	}

	if data == nil {
		return common.Errorf("idml", "validate file size", f.Name, "file data is nil")
	}

	// Check for potential overflow before conversion
	if f.UncompressedSize64 > math.MaxInt64-1024 {
		return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("file size %d bytes exceeds maximum supported size", f.UncompressedSize64))
	}

	// #nosec G115 - Conversion is safe: overflow check performed on line above
	uncompressedSize := int64(f.UncompressedSize64)
	maxExpectedSize := uncompressedSize + 1024
	if int64(len(data)) > maxExpectedSize {
		return common.WrapErrorWithPath("idml", "read", f.Name, fmt.Errorf("actual size %d exceeds declared size %d (potential ZIP bomb)", len(data), f.UncompressedSize64))
	}
	return nil
}

// storeFileInPackage stores the extracted file data in the package.
func storeFileInPackage(pkg *Package, f *zip.File, data []byte) {
	// Add validation for parameters
	if pkg == nil || f == nil {
		return // Silently ignore invalid parameters to avoid panics
	}

	// Ensure data is not nil
	if data == nil {
		data = []byte{}
	}

	pkg.files[f.Name] = &fileEntry{
		data:   data,
		header: &f.FileHeader,
	}
	pkg.fileOrder = append(pkg.fileOrder, f.Name)
}

// validateDesignMap ensures the designmap.xml file exists and can be parsed.
func validateDesignMap(pkg *Package) error {
	if _, err := pkg.Document(); err != nil {
		// If designmap.xml doesn't exist or can't be parsed, that's an error
		// since it's a required file in IDML packages
		return err
	}
	return nil
}

// ReadBytes parses IDML from an in-memory byte slice with default safety limits.
//
// This is useful when you already have the IDML data in memory (e.g., from
// S3, HTTP response, or other sources) and want to avoid writing to a temp file.
//
// The function applies the same ZIP bomb protection as Read().
func ReadBytes(data []byte) (*Package, error) {
	return ReadBytesWithOptions(data, nil)
}

// ReadBytesWithOptions parses IDML from a byte slice with configurable safety limits.
//
// If opts is nil, default limits are applied. To disable a specific limit,
// set it to -1.
//
// Example:
//
//	data, _ := os.ReadFile("document.idml")
//	pkg, err := idml.ReadBytes(data)
func ReadBytesWithOptions(data []byte, opts *ReadOptions) (*Package, error) {
	// Apply defaults
	if opts == nil {
		opts = &ReadOptions{}
	}
	opts.applyDefaults()

	// Create a bytes.Reader which implements io.ReaderAt
	bytesReader := bytes.NewReader(data)

	// Create a zip.Reader from the bytes
	zipReader, err := zip.NewReader(bytesReader, int64(len(data)))
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "read bytes", "<memory>", err)
	}

	// Use shared extraction logic
	return extractPackage(zipReader.File, opts, "<memory>")
}

// ReadFrom parses IDML from an io.ReaderAt with default safety limits.
//
// This is useful for streaming scenarios where you have an io.ReaderAt
// (e.g., bytes.Reader, os.File, or custom implementations) and want to
// parse IDML without loading into a byte slice first.
//
// The size parameter must be the total size of the IDML data.
func ReadFrom(r io.ReaderAt, size int64) (*Package, error) {
	return ReadFromWithOptions(r, size, nil)
}

// ReadFromWithOptions parses IDML from an io.ReaderAt with configurable safety limits.
//
// If opts is nil, default limits are applied. To disable a specific limit,
// set it to -1.
//
// Example:
//
//	file, _ := os.Open("document.idml")
//	stat, _ := file.Stat()
//	pkg, err := idml.ReadFrom(file, stat.Size())
func ReadFromWithOptions(r io.ReaderAt, size int64, opts *ReadOptions) (*Package, error) {
	// Apply defaults
	if opts == nil {
		opts = &ReadOptions{}
	}
	opts.applyDefaults()

	// Create a zip.Reader from the io.ReaderAt
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "read from reader", "<stream>", err)
	}

	// Use shared extraction logic
	return extractPackage(zipReader.File, opts, "<stream>")
}

// Read opens an IDML file and loads it into memory with default safety limits.
//
// The function reads the entire ZIP archive and:
// 1. Stores each file's content as raw bytes
// 2. Automatically parses designmap.xml into a Document struct
// 3. Preserves ZIP metadata for perfect roundtrip
// 4. Applies ZIP bomb protection with default limits
//
// This allows for both byte-perfect preservation of unknown files
// and structured manipulation of the document manifest.
//
// Use ReadWithOptions for custom limits or to disable protection.
func Read(path string) (*Package, error) {
	return ReadWithOptions(path, nil)
}

// ReadWithOptions opens an IDML file with configurable safety limits.
//
// If opts is nil, default limits are applied. To disable a specific limit,
// set it to -1. For example:
//
//	// Allow unlimited total size but keep other limits
//	pkg, err := idml.ReadWithOptions(path, &idml.ReadOptions{
//	    MaxTotalSize: -1,
//	})
func ReadWithOptions(path string, opts *ReadOptions) (*Package, error) {
	// Apply defaults
	if opts == nil {
		opts = &ReadOptions{}
	}
	opts.applyDefaults()

	// Open the ZIP file
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "read", path, err)
	}
	defer r.Close()

	// Use shared extraction logic
	return extractPackage(r.File, opts, path)
}
