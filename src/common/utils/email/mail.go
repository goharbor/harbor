// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package email

import (
	tlspkg "crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/hashicorp/go-multierror"
)

// Send ...
func Send(addr, identity, username, password string,
	timeout int, tls, insecure bool, from string,
	to []string, subject, message string) error {

	client, err := newClient(addr, identity, username,
		password, timeout, tls, insecure)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, t := range to {
		if err = client.Rcpt(t); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	template := "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\n%s\r\n"
	data := fmt.Sprintf(template, from,
		strings.Join(to, ","), subject, message)

	_, err = w.Write([]byte(data))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// Ping tests the connection and authentication with email server
// If tls is true, a secure connection is established, or Ping
// trys to upgrate the insecure connection to a secure one if
// email server supports it.
// Ping doesn't verify the server's certificate and hostname when
// needed if the parameter insecure is ture
func Ping(addr, identity, username, password string,
	timeout int, tls, insecure bool) error {
	client, err := newClient(addr, identity, username, password,
		timeout, tls, insecure)
	if err != nil {
		return err
	}
	defer client.Close()
	return nil
}

// caller needs to close the client
func newClient(addr, identity, username, password string,
	timeout int, tls, insecure bool) (*smtp.Client, error) {
	log.Debugf("establishing TCP connection with %s ...", addr)
	var conn net.Conn

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	tlsOK := true

	conn, err = tlspkg.Dial("tcp", addr, &tlspkg.Config{
		ServerName:         host,
		InsecureSkipVerify: insecure,
	})
	if err != nil {
		tlsOK = false

		log.Debugf("could not establish TLS connection to %s: %v", addr, err)
		dialer := net.Dialer{}
		conn, err = dialer.Dial("tcp", addr)
		if err != nil {
			return nil, errors.Wrap(err, "establish connection to server")
		}
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return nil, errors.Wrap(err, "create SMTP client")
	}

	ok, _ := client.Extension("STARTTLS")
	if ok {
		tlsConf := &tlspkg.Config{
			ServerName:         host,
			InsecureSkipVerify: insecure,
		}

		if err := client.StartTLS(tlsConf); err != nil {
			return nil, errors.Wrap(err, "send STARTTLS command")
		}
	} else if tls && !tlsOK {
		return nil, errors.Errorf("%q does not advertise the STARTTLS extension and establishing a TLS connection failed", host)
	}

	if ok, mech := client.Extension("AUTH"); ok {
		auth, err := findAuthMech(mech, identity, username, password, host)
		if err != nil {
			return nil, errors.Wrap(err, "find auth mechanism")
		}
		if auth != nil {
			if err := client.Auth(auth); err != nil {
				return nil, errors.Wrapf(err, "authentication with mechanism %T", auth)
			}
		}
	} else {
		log.Debugf("the email server %s does not support AUTH, skip",
			addr)
	}

	log.Debug("create smtp client successfully")

	return client, nil
}

func findAuthMech(mechs, identity, username, password, host string) (smtp.Auth, error) {
	var err error
	for _, mech := range strings.Split(mechs, " ") {
		switch mech {
		case "CRAM-MD5":
			if password == "" {
				multierror.Append(err, errors.New("missing secret for CRAM-MD5 auth mechanism"))
				continue
			}
			return smtp.CRAMMD5Auth(username, password), nil
		case "PLAIN":
			if password == "" {
				multierror.Append(err, errors.New("missing password for PLAIN auth mechanism"))
				continue
			}
			return smtp.PlainAuth(identity, username, password, host), nil
		case "LOGIN":
			if password == "" {
				multierror.Append(err, errors.New("missing password for LOGIN auth mechanism"))
				continue
			}
			return LoginAuth(username, password), nil
		}
	}
	if err == nil {
		multierror.Append(err, errors.New("unknown auth mechanism: "+mechs))
	}
	return nil, err
}

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

// Used for AUTH LOGIN. (Maybe password should be encrypted)
func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch strings.ToLower(string(fromServer)) {
		case "username:":
			return []byte(a.username), nil
		case "password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unexpected server challenge")
		}
	}
	return nil, nil
}
