package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	//ErrNoRefs not found refs
	ErrNoRefs = errors.New("no refs")
	//ErrBadURI bad git uri
	ErrBadURI = errors.New("bad uri")
	//ErrBadName bad repo name
	ErrBadName = errors.New("bad repo name")
	//ErrBadWorkspace invalid workspace dir
	ErrBadWorkspace = errors.New("bad workspace")
	//ErrBadCmd bad git subcommand
	ErrBadCmd = errors.New("bad git subcommand")
	//ErrRemoteAdd git remote add failed
	ErrRemoteAdd = errors.New("git remote add failed")
	//ErrBadRef invalid ref
	ErrBadRef = errors.New("bad ref")
	//ErrBadFile invalid filepath
	ErrBadFile = errors.New("bad file")
	//ErrBadCommit invalid commit
	ErrBadCommit = errors.New("bad commit")
)

//Client git client for executing git commands
type Client struct {
	URI    string
	Branch string
	Path   string
	Dir    string
}

//NewClient create new client
func NewClient(workspace, project, name, uri, branch string) (*Client, error) {
	if len(uri) == 0 {
		return nil, ErrBadURI
	}

	client := &Client{
		URI:    uri,
		Branch: branch,
		Dir:    name,
	}

	if err := client.initRepo(fmt.Sprintf("%s\\%s", workspace, project)); err != nil {
		return nil, err
	}
	return client, nil
}

//String impl
func (client *Client) String() string {
	return fmt.Sprintf("uri: %s, branch: %s", client.URI, client.Branch)
}

//initRepo init empty repo
func (client *Client) initRepo(workspace string) error {
	if len(workspace) == 0 {
		workspace = "/var/lib/drone/workspace/"
	}
	if !filepath.IsAbs(workspace) {
		log.Errorln("bad workspace", workspace)
		return ErrBadWorkspace
	}
	path := workspace
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	client.Path = path
	return nil
}

func gitCmd(path string, args ...string) (*exec.Cmd, error) {
	if len(path) == 0 {
		log.Errorln("bad workspace", path)
		return nil, ErrBadWorkspace
	}
	if len(args) < 1 {
		return nil, ErrBadCmd
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	return cmd, nil
}

//clone the repository
func (client *Client) Clone() error {
	cmd, err := gitCmd(client.Path, "clone", client.URI)
	if err != nil {
		return err
	}
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

//pull the update info from remote branch
func (client *Client) Pull() error {
	cmd, err := gitCmd(fmt.Sprintf("%s\\%s", client.Path, client.Dir), "pull", "origin", client.Branch)
	if err != nil {
		return err
	}
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

//update Client Catalog
func (client *Client) updateCatalog() {
	gitTimer := time.NewTicker(time.Second * 2)
	for {
		<-gitTimer.C
		client.Pull()
	}
}

// Trace writes each command to standard error (preceded by a ‘$ ’) before it
// is executed. Used for debugging your build.
func trace(cmd *exec.Cmd) {
	log.Infoln("$", cmd.Dir, strings.Join(cmd.Args, " "))
}
