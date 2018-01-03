package rethinkdb

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-connections/tlsconfig"
	"gopkg.in/dancannon/gorethink.v3"
)

// Timing can be embedded into other gorethink models to
// add time tracking fields
type Timing struct {
	CreatedAt time.Time `gorethink:"created_at"`
	UpdatedAt time.Time `gorethink:"updated_at"`
	DeletedAt time.Time `gorethink:"deleted_at"`
}

// AdminConnection sets up an admin RethinkDB connection to the host (`host:port` format)
// using the CA .pem file provided at path `caFile`
func AdminConnection(tlsOpts tlsconfig.Options, host string) (*gorethink.Session, error) {
	logrus.Debugf("attempting to connect admin to host %s", host)
	t, err := tlsconfig.Client(tlsOpts)
	if err != nil {
		return nil, err
	}
	return gorethink.Connect(
		gorethink.ConnectOpts{
			Address:   host,
			TLSConfig: t,
		},
	)
}

// UserConnection sets up a user RethinkDB connection to the host (`host:port` format)
// using the CA .pem file provided at path `caFile`, using the provided username.
func UserConnection(tlsOpts tlsconfig.Options, host, username, password string) (*gorethink.Session, error) {
	logrus.Debugf("attempting to connect user %s to host %s", username, host)
	t, err := tlsconfig.Client(tlsOpts)
	if err != nil {
		return nil, err
	}
	return gorethink.Connect(
		gorethink.ConnectOpts{
			Address:   host,
			TLSConfig: t,
			Username:  username,
			Password:  password,
		},
	)
}
