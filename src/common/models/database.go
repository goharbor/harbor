//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package models

// Database ...
type Database struct {
	Type       string      `json:"type"`
	PostGreSQL *PostGreSQL `json:"postgresql,omitempty"`
}

// MySQL ...
type MySQL struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Database string `json:"database"`
}

// SQLite ...
type SQLite struct {
	File string `json:"file"`
}

// PostGreSQL ...
type PostGreSQL struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password,omitempty"`
	Database     string `json:"database"`
	SSLMode      string `json:"sslmode"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxOpenConns int    `json:"max_open_conns"`
}
