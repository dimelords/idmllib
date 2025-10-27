package idml

import (
	"archive/zip"
	"fmt"
	"io"
)

// Open opens an IDML file and reads its structure
func Open(path string) (*Package, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open IDML: %w", err)
	}

	pkg := &Package{
		path:   path,
		reader: r,
	}

	if err := pkg.load(); err != nil {
		_ = r.Close()
		return nil, err
	}

	return pkg, nil
}

func (p *Package) load() error {
	for _, f := range p.reader.File {
		data, err := getZipFileData(f)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", f.Name, err)
		}
		switch name := f.Name; {
		case isStory(name):
			story, err := p.readStory(data)
			if err != nil {
				return fmt.Errorf("failed to read story %s: %w", f.Name, err)
			}
			p.Stories = append(p.Stories, story)
		case isSpread(name):
			spread, err := p.readSpread(data)
			if err != nil {
				return fmt.Errorf("failed to read spread %s: %w", f.Name, err)
			}
			p.Spreads = append(p.Spreads, spread)
		}
	}

	return nil
}

// Close closes the IDML package
func (p *Package) Close() error {
	if p.reader != nil {
		return p.reader.Close()
	}
	return nil
}

// readFileFromIDML reads a file from the IDML zip archive
func (p *Package) readFileFromIDML(name string) ([]byte, error) {
	for _, f := range p.reader.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer func(rc io.ReadCloser) {
				_ = rc.Close()
			}(rc)
			return io.ReadAll(rc)
		}
	}
	return nil, &FileNotFoundError{FileName: name}
}

// readFileFromZipFile reads data from a zip.File
func (p *Package) readFileFromZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer func(rc io.ReadCloser) {
		_ = rc.Close()
	}(rc)
	return io.ReadAll(rc)
}

func getZipFileData(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer func(rc io.ReadCloser) {
		_ = rc.Close()
	}(rc)

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return data, nil
}
