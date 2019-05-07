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
	"bytes"
	tlspkg "crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/utils/log"
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
	conn, err := net.DialTimeout("tcp", addr,
		time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	if tls {
		log.Debugf("establishing SSL/TLS connection with %s ...", addr)
		tlsConn := tlspkg.Client(conn, &tlspkg.Config{
			ServerName:         host,
			InsecureSkipVerify: insecure,
		})
		if err = tlsConn.Handshake(); err != nil {
			return nil, err
		}

		conn = tlsConn
	}

	log.Debugf("creating SMTP client for %s ...", host)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return nil, err
	}

	// try to swith to SSL/TLS
	if !tls {
		if ok, _ := client.Extension("STARTTLS"); ok {
			log.Debugf("switching the connection with %s to SSL/TLS ...", addr)
			if err = client.StartTLS(&tlspkg.Config{
				ServerName:         host,
				InsecureSkipVerify: insecure,
			}); err != nil {
				return nil, err
			}
		} else {
			log.Debugf("the email server %s does not support STARTTLS", addr)
		}
	}

	// refer to https://github.com/go-gomail/gomail/blob/master/smtp.go
	if ok, auths := client.Extension("AUTH"); ok {
		log.Debug("authenticating the client...")
		var auth smtp.Auth
		if strings.Contains(auths, "CRAM-MD5") {
			auth = smtp.CRAMMD5Auth(username, password)
		} else if strings.Contains(auths, "LOGIN") &&
			!strings.Contains(auths, "PLAIN") {
			auth = &loginAuth{
				username: username,
				password: password,
				host:     host,
			}
		} else {
			auth = smtp.PlainAuth("", username, password, host)
		}
		if err = client.Auth(auth); err != nil {
			return nil, err
		}
	} else {
		log.Debugf("the email server %s does not support AUTH, skip",
			addr)
	}

	log.Debug("create smtp client successfully")

	return client, nil
}

// refer to https://github.com/go-gomail/gomail/blob/master/smtp.go
type loginAuth struct {
	username string
	password string
	host     string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		advertised := false
		for _, mechanism := range server.Auth {
			if mechanism == "LOGIN" {
				advertised = true
				break
			}
		}
		if !advertised {
			return "", nil, errors.New("unencrypted connection")
		}
	}
	if server.Name != a.host {
		return "", nil, errors.New("wrong host name")
	}
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}

	switch {
	case bytes.Equal(fromServer, []byte("Username:")):
		return []byte(a.username), nil
	case bytes.Equal(fromServer, []byte("Password:")):
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected server challenge: %s", fromServer)
	}
}
