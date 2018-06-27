package github

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	nurl "net/url"
	"os"
	"path"
	"strings"

	"github.com/golang-migrate/migrate/source"
	"github.com/google/go-github/github"
)

func init() {
	source.Register("github", &Github{})
}

var (
	ErrNoUserInfo          = fmt.Errorf("no username:token provided")
	ErrNoAccessToken       = fmt.Errorf("no access token")
	ErrInvalidRepo         = fmt.Errorf("invalid repo")
	ErrInvalidGithubClient = fmt.Errorf("expected *github.Client")
	ErrNoDir               = fmt.Errorf("no directory")
)

type Github struct {
	client *github.Client
	url    string

	pathOwner  string
	pathRepo   string
	path       string
	options    *github.RepositoryContentGetOptions
	migrations *source.Migrations
}

type Config struct {
}

func (g *Github) Open(url string) (source.Driver, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	if u.User == nil {
		return nil, ErrNoUserInfo
	}

	password, ok := u.User.Password()
	if !ok {
		return nil, ErrNoUserInfo
	}

	tr := &github.BasicAuthTransport{
		Username: u.User.Username(),
		Password: password,
	}

	gn := &Github{
		client:     github.NewClient(tr.Client()),
		url:        url,
		migrations: source.NewMigrations(),
		options:    &github.RepositoryContentGetOptions{Ref: u.Fragment},
	}

	// set owner, repo and path in repo
	gn.pathOwner = u.Host
	pe := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pe) < 1 {
		return nil, ErrInvalidRepo
	}
	gn.pathRepo = pe[0]
	if len(pe) > 1 {
		gn.path = strings.Join(pe[1:], "/")
	}

	if err := gn.readDirectory(); err != nil {
		return nil, err
	}

	return gn, nil
}

func WithInstance(client *github.Client, config *Config) (source.Driver, error) {
	gn := &Github{
		client:     client,
		migrations: source.NewMigrations(),
	}
	if err := gn.readDirectory(); err != nil {
		return nil, err
	}
	return gn, nil
}

func (g *Github) readDirectory() error {
	fileContent, dirContents, _, err := g.client.Repositories.GetContents(context.Background(), g.pathOwner, g.pathRepo, g.path, g.options)
	if err != nil {
		return err
	}
	if fileContent != nil {
		return ErrNoDir
	}

	for _, fi := range dirContents {
		m, err := source.DefaultParse(*fi.Name)
		if err != nil {
			continue // ignore files that we can't parse
		}
		if !g.migrations.Append(m) {
			return fmt.Errorf("unable to parse file %v", *fi.Name)
		}
	}

	return nil
}

func (g *Github) Close() error {
	return nil
}

func (g *Github) First() (version uint, er error) {
	if v, ok := g.migrations.First(); !ok {
		return 0, &os.PathError{"first", g.path, os.ErrNotExist}
	} else {
		return v, nil
	}
}

func (g *Github) Prev(version uint) (prevVersion uint, err error) {
	if v, ok := g.migrations.Prev(version); !ok {
		return 0, &os.PathError{fmt.Sprintf("prev for version %v", version), g.path, os.ErrNotExist}
	} else {
		return v, nil
	}
}

func (g *Github) Next(version uint) (nextVersion uint, err error) {
	if v, ok := g.migrations.Next(version); !ok {
		return 0, &os.PathError{fmt.Sprintf("next for version %v", version), g.path, os.ErrNotExist}
	} else {
		return v, nil
	}
}

func (g *Github) ReadUp(version uint) (r io.ReadCloser, identifier string, err error) {
	if m, ok := g.migrations.Up(version); ok {
		file, _, _, err := g.client.Repositories.GetContents(context.Background(), g.pathOwner, g.pathRepo, path.Join(g.path, m.Raw), g.options)
		if err != nil {
			return nil, "", err
		}
		if file != nil {
			r, err := file.GetContent()
			if err != nil {
				return nil, "", err
			}
			return ioutil.NopCloser(bytes.NewReader([]byte(r))), m.Identifier, nil
		}
	}
	return nil, "", &os.PathError{fmt.Sprintf("read version %v", version), g.path, os.ErrNotExist}
}

func (g *Github) ReadDown(version uint) (r io.ReadCloser, identifier string, err error) {
	if m, ok := g.migrations.Down(version); ok {
		file, _, _, err := g.client.Repositories.GetContents(context.Background(), g.pathOwner, g.pathRepo, path.Join(g.path, m.Raw), g.options)
		if err != nil {
			return nil, "", err
		}
		if file != nil {
			r, err := file.GetContent()
			if err != nil {
				return nil, "", err
			}
			return ioutil.NopCloser(bytes.NewReader([]byte(r))), m.Identifier, nil
		}
	}
	return nil, "", &os.PathError{fmt.Sprintf("read version %v", version), g.path, os.ErrNotExist}
}
