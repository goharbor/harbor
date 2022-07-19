package model

import (
	"encoding/json"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"

	"github.com/beego/beego/orm"
)

func init() {
	orm.RegisterModel(&Robot{})
}

// Robot holds the details of a robot.
type Robot struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Name         string    `orm:"column(name)" json:"name" sort:"default"`
	Description  string    `orm:"column(description)" json:"description"`
	Secret       string    `orm:"column(secret)" json:"secret"`
	Salt         string    `orm:"column(salt)" json:"-"`
	Duration     int64     `orm:"column(duration)" json:"duration"`
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	ExpiresAt    int64     `orm:"column(expiresat)" json:"expires_at"`
	Disabled     bool      `orm:"column(disabled)" json:"disabled"`
	Visible      bool      `orm:"column(visible)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName ...
func (r *Robot) TableName() string {
	return "robot"
}

// FromJSON parses robot from json data
func (r *Robot) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to parse")
	}

	return json.Unmarshal([]byte(jsonData), r)
}

// ToJSON marshals Robot to JSON data
func (r *Robot) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
