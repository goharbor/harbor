/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package email

import (
	"bytes"
	"crypto/tls"
	"strconv"
	//"strings"

	"net/smtp"
	"text/template"

	//"github.com/astaxie/beego"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/ui/config"
)

// Mail holds information about content of Email
type Mail struct {
	From    string
	To      []string
	Subject string
	Message string
}

var mc models.Email

// SendMail sends Email according to the configurations
func (m Mail) SendMail() error {
	mc, err := config.Email()
	if err != nil {
		return err
	}

	mailTemplate, err := template.ParseFiles("views/mail.tpl")
	if err != nil {
		return err
	}
	mailContent := new(bytes.Buffer)
	err = mailTemplate.Execute(mailContent, m)
	if err != nil {
		return err
	}
	content := mailContent.Bytes()

	auth := smtp.PlainAuth(mc.Identity, mc.Username, mc.Password, mc.Host)
	if mc.SSL {
		err = sendMailWithTLS(m, auth, content)
	} else {
		err = sendMail(m, auth, content)
	}

	return err
}

func sendMail(m Mail, auth smtp.Auth, content []byte) error {
	return smtp.SendMail(mc.Host+":"+strconv.Itoa(mc.Port), auth, m.From, m.To, content)
}

func sendMailWithTLS(m Mail, auth smtp.Auth, content []byte) error {
	conn, err := tls.Dial("tcp", mc.Host+":"+strconv.Itoa(mc.Port), nil)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, mc.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("AUTH"); ok {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(m.From); err != nil {
		return err
	}

	for _, to := range m.To {
		if err = client.Rcpt(to); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(content)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

/*
func loadConfig() {
	config, err := beego.AppConfig.GetSection("mail")
	if err != nil {
		panic(err)
	}

	var useTLS = false
	if config["ssl"] != "" && strings.ToLower(config["ssl"]) == "true" {
		useTLS = true
	}
	mc = MailConfig{
		Identity: config["identity"],
		Host:     config["host"],
		Port:     config["port"],
		Username: config["username"],
		Password: config["password"],
		TLS:      useTLS,
	}
}
*/
