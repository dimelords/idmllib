package idml

import (
	"archive/zip"
	"io"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Lazy reading.
//
// By default Read loads every entry of the archive into memory. Most work only
// touches a few of them: reading the document structure needs designmap.xml,
// extracting text needs the stories, and an image-heavy spread may never be
// looked at. With ReadOptions.Lazy an entry is read the first time it is
// actually used, and an entry that is never used is never decompressed.
//
// Writing benefits too. An entry that was neither loaded nor modified is copied
// straight from the source archive in its compressed form, so it is not
// decompressed and not recompressed.
//
// The source must stay readable and unchanged for the lifetime of the package.
// A package read lazily from a file holds that file open; call Close to release
// it. Close is safe to call on any package.

// load materializes an entry's bytes from the source archive if they have not
// been read yet.
func (p *Package) load(filename string, entry *fileEntry) error {
	if entry.data != nil || entry.src == nil {
		return nil
	}
	data, err := extractZipFileData(entry.src, p.readOpts, p.source)
	if err != nil {
		return err
	}
	entry.data = data
	return nil
}

// isUnloaded reports whether the entry still lives only in the source archive,
// so it can be copied across without being decompressed.
func (e *fileEntry) isUnloaded() bool {
	return e.data == nil && e.src != nil
}

// Close releases the source archive held by a lazily read package. Entries that
// were already loaded stay usable; reading an entry that was never loaded fails
// afterwards. Close is a no-op for packages read eagerly, and safe to call more
// than once.
func (p *Package) Close() error {
	closer := p.closer
	p.closer = nil
	for _, entry := range p.files {
		entry.src = nil
	}
	if closer == nil {
		return nil
	}
	if err := closer.Close(); err != nil {
		return common.WrapErrorWithPath("idml", "close", p.source, err)
	}
	return nil
}

// Lazy reports whether the package was read with ReadOptions.Lazy.
func (p *Package) Lazy() bool {
	return p.lazy
}

// LoadedFiles returns the number of entries whose bytes are currently in
// memory. For a lazily read package this grows as files are used.
func (p *Package) LoadedFiles() int {
	n := 0
	for _, entry := range p.files {
		if entry.data != nil {
			n++
		}
	}
	return n
}

// LoadAll materializes every entry, after which the package no longer depends
// on the source archive. It is the way to turn a lazily read package into a
// self-contained one before closing the source.
func (p *Package) LoadAll() error {
	for _, name := range p.fileOrder {
		entry, ok := p.files[name]
		if !ok {
			continue
		}
		if err := p.load(name, entry); err != nil {
			return err
		}
	}
	return nil
}

// writeEntry writes one entry to the output archive, copying it in compressed
// form when it was never loaded or modified.
func writeEntry(w *zip.Writer, name string, entry *fileEntry) error {
	if entry.isUnloaded() {
		if err := w.Copy(entry.src); err != nil {
			return common.WrapErrorWithPath("idml", "write", name, err)
		}
		return nil
	}
	if entry.header == nil {
		entry.header = defaultHeader(name)
	}
	fileWriter, err := w.CreateHeader(entry.header)
	if err != nil {
		return common.WrapErrorWithPath("idml", "write", name, err)
	}
	if _, err := fileWriter.Write(entry.data); err != nil {
		return common.WrapErrorWithPath("idml", "write", name, err)
	}
	return nil
}

// closerFunc adapts a function to io.Closer.
type closerFunc func() error

func (f closerFunc) Close() error { return f() }

var _ io.Closer = closerFunc(nil)
