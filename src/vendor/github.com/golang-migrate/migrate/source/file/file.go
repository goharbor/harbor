package file

import (
	"fmt"
	"io"
	"io/ioutil"
	nurl "net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/golang-migrate/migrate/source"
)

func init() {
	source.Register("file", &File{})
}

type File struct {
	url        string
	path       string
	migrations *source.Migrations
}

func (f *File) Open(url string) (source.Driver, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	// concat host and path to restore full path
	// host might be `.`
	p := u.Host + u.Path

	if len(p) == 0 {
		// default to current directory if no path
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		p = wd

	} else if p[0:1] == "." || p[0:1] != "/" {
		// make path absolute if relative
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		p = abs
	}

	// scan directory
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return nil, err
	}

	nf := &File{
		url:        url,
		path:       p,
		migrations: source.NewMigrations(),
	}

	for _, fi := range files {
		if !fi.IsDir() {
			m, err := source.DefaultParse(fi.Name())
			if err != nil {
				continue // ignore files that we can't parse
			}
			if !nf.migrations.Append(m) {
				return nil, fmt.Errorf("unable to parse file %v", fi.Name())
			}
		}
	}
	return nf, nil
}

func (f *File) Close() error {
	// nothing do to here
	return nil
}

func (f *File) First() (version uint, err error) {
	if v, ok := f.migrations.First(); !ok {
		return 0, &os.PathError{Op: "first", Path: f.path, Err: os.ErrNotExist}
	} else {
		return v, nil
	}
}

func (f *File) Prev(version uint) (prevVersion uint, err error) {
	if v, ok := f.migrations.Prev(version); !ok {
		return 0, &os.PathError{Op: fmt.Sprintf("prev for version %v", version), Path: f.path, Err: os.ErrNotExist}
	} else {
		return v, nil
	}
}

func (f *File) Next(version uint) (nextVersion uint, err error) {
	if v, ok := f.migrations.Next(version); !ok {
		return 0, &os.PathError{Op: fmt.Sprintf("next for version %v", version), Path: f.path, Err: os.ErrNotExist}
	} else {
		return v, nil
	}
}

func (f *File) ReadUp(version uint) (r io.ReadCloser, identifier string, err error) {
	if m, ok := f.migrations.Up(version); ok {
		r, err := os.Open(path.Join(f.path, m.Raw))
		if err != nil {
			return nil, "", err
		}
		return r, m.Identifier, nil
	}
	return nil, "", &os.PathError{Op: fmt.Sprintf("read version %v", version), Path: f.path, Err: os.ErrNotExist}
}

func (f *File) ReadDown(version uint) (r io.ReadCloser, identifier string, err error) {
	if m, ok := f.migrations.Down(version); ok {
		r, err := os.Open(path.Join(f.path, m.Raw))
		if err != nil {
			return nil, "", err
		}
		return r, m.Identifier, nil
	}
	return nil, "", &os.PathError{Op: fmt.Sprintf("read version %v", version), Path: f.path, Err: os.ErrNotExist}
}
