package idml

import (
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// File I/O helper methods for Package struct.
// These methods provide internal file access and manipulation utilities.

// hasFile checks if a file exists in the package.
func (p *Package) hasFile(filename string) bool {
	_, exists := p.files[filename]
	return exists
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
		entry.src = nil // the bytes now live here, not in the source archive
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

	// In lazy mode the bytes are read the first time the entry is used.
	if err := p.load(filename, entry); err != nil {
		return nil, err
	}
	if entry.data == nil {
		return nil, common.WrapErrorWithPath("idml", "get file entry", filename,
			common.Errorf("idml", "get file entry", filename, "package was closed before this file was read"))
	}

	return entry, nil
}
