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

	// try to switch to SSL/TLS
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

	if ok, mechs := client.Extension("AUTH"); ok {
		log.Debug("authenticating the client...")
		auth, err := handleAuth(host, identity, username, password, mechs)
		if err != nil {
			return nil, err
		}

		if auth != nil {
			if err = client.Auth(auth); err != nil {
				return nil, err
			}
		}
	} else {
		log.Debugf("the email server %s does not support AUTH, skip",
			addr)
	}

	log.Debug("create smtp client successfully")

	return client, nil
}

func handleAuth(host, identity, username, password, mechs string) (smtp.Auth, error) {
	if username == "" {
		log.Debug("username is not configured, attempting to send email without authenticating")
		return nil, nil
	}

	for _, mech := range strings.Split(mechs, " ") {
		switch mech {
		case "CRAM-MD5":
			return smtp.CRAMMD5Auth(username, password), nil
		case "PLAIN":
			return smtp.PlainAuth(identity, username, password, host), nil
		}
	}

	return nil, errors.New("unknown auth mechanism: " + mechs)
}
