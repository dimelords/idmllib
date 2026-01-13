package idml

import (
	"github.com/dimelords/idmllib/pkg/common"
)

// File I/O helper methods for Package struct.
// These methods provide internal file access and manipulation utilities.

// hasFile checks if a file exists in the package.
func (p *Package) hasFile(filename string) bool {
	_, exists := p.files[filename]
	return exists
}

// getFileData returns the raw data for a file.
// Returns ErrNotFound if the file doesn't exist.
func (p *Package) getFileData(filename string) ([]byte, error) {
	// Add validation for filename
	if filename == "" {
		return nil, common.Errorf("idml", "get file data", "", "filename is empty")
	}

	entry, exists := p.files[filename]
	if !exists {
		return nil, common.WrapErrorWithPath("idml", "get file data", filename, common.ErrNotFound)
	}

	// Add validation for file entry
	if entry == nil {
		return nil, common.WrapErrorWithPath("idml", "get file data", filename, common.Errorf("idml", "get file data", filename, "file entry is nil"))
	}

	return entry.data, nil
}

// setFileData sets the raw data for a file.
// Creates a new fileEntry if the file doesn't exist.
// Preserves the existing ZIP header if the file already exists.
func (p *Package) setFileData(filename string, data []byte) {
	// Add validation for filename
	if filename == "" {
		// Log error but don't fail - this is a void function
		return
	}

	// Add validation for data (nil is allowed for empty files)
	if data == nil {
		data = []byte{}
	}

	if entry, exists := p.files[filename]; exists {
		// Preserve existing header, update data
		entry.data = data
	} else {
		// Create new entry
		p.files[filename] = &fileEntry{
			data: data,
			// header will be created during Write() if needed
		}
		// Add to file order if it's a new file
		p.fileOrder = append(p.fileOrder, filename)
	}
}

// removeFile removes a file from the package.
// Returns true if the file was removed, false if it didn't exist.
func (p *Package) removeFile(filename string) bool {
	// Add validation for filename
	if filename == "" {
		return false
	}

	if _, exists := p.files[filename]; !exists {
		return false
	}

	// Remove from files map
	delete(p.files, filename)

	// Remove from fileOrder
	for i, name := range p.fileOrder {
		if name == filename {
			p.fileOrder = append(p.fileOrder[:i], p.fileOrder[i+1:]...)
			break
		}
	}

	return true
}

// getFileEntry returns the complete fileEntry for a file.
// This provides access to both data and ZIP metadata.
// Returns ErrNotFound if the file doesn't exist.
func (p *Package) getFileEntry(filename string) (*fileEntry, error) {
	// Add validation for filename
	if filename == "" {
		return nil, common.Errorf("idml", "get file entry", "", "filename is empty")
	}

	entry, exists := p.files[filename]
	if !exists {
		return nil, common.WrapErrorWithPath("idml", "get file entry", filename, common.ErrNotFound)
	}

	// Add validation for file entry
	if entry == nil {
		return nil, common.WrapErrorWithPath("idml", "get file entry", filename, common.Errorf("idml", "get file entry", filename, "file entry is nil"))
	}

	return entry, nil
}

// copyFileData creates a copy of file data to prevent accidental modification.
// Returns ErrNotFound if the file doesn't exist.
func (p *Package) copyFileData(filename string) ([]byte, error) {
	// Add validation for filename
	if filename == "" {
		return nil, common.Errorf("idml", "copy file data", "", "filename is empty")
	}

	entry, exists := p.files[filename]
	if !exists {
		return nil, common.WrapErrorWithPath("idml", "copy file data", filename, common.ErrNotFound)
	}

	// Add validation for file entry
	if entry == nil {
		return nil, common.WrapErrorWithPath("idml", "copy file data", filename, common.Errorf("idml", "copy file data", filename, "file entry is nil"))
	}

	// Handle nil data gracefully
	if entry.data == nil {
		return []byte{}, nil
	}

	// Create a copy to prevent modification of original data
	dataCopy := make([]byte, len(entry.data))
	copy(dataCopy, entry.data)
	return dataCopy, nil
}

// getFileSize returns the size of a file in bytes.
// Returns 0 if the file doesn't exist.
func (p *Package) getFileSize(filename string) int {
	entry, exists := p.files[filename]
	if !exists {
		return 0
	}
	return len(entry.data)
}

// listFilesByPattern returns all filenames that match a pattern.
// This is useful for finding all files in a directory (e.g., "Stories/", "Spreads/").
func (p *Package) listFilesByPattern(pattern string) []string {
	var matches []string
	for filename := range p.files {
		// Simple prefix matching - could be enhanced with regex if needed
		if len(filename) >= len(pattern) && filename[:len(pattern)] == pattern {
			matches = append(matches, filename)
		}
	}
	return matches
}
