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
package utils

import (
	"bytes"

	"net/smtp"
	"text/template"

	"github.com/astaxie/beego"
)

type Mail struct {
	From    string
	To      []string
	Subject string
	Message string
}
type MailConfig struct {
	Identity string
	Host     string
	Port     string
	Username string
	Password string
}

var mc MailConfig

func (m Mail) SendMail() error {

	if mc.Host == "" {
		loadConfig()
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
	return smtp.
		SendMail(mc.Host+":"+mc.Port,
		smtp.PlainAuth(mc.Identity, mc.Username, mc.Password, mc.Host),
		m.From, m.To, mailContent.Bytes())
}

func loadConfig() {
	config, err := beego.AppConfig.GetSection("mail")
	if err != nil {
		panic(err)
	}
	mc = MailConfig{
		Identity: "Mail Config",
		Host:     config["host"],
		Port:     config["port"],
		Username: config["username"],
		Password: config["password"],
	}
}
