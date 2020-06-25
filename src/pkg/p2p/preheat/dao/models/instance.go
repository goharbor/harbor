package models

// Instance defines distribution instance metadata.
type Instance struct {
	ID             int64  `orm:"pk;auto;column(id)" json:"id"`
	Name           string `orm:"column(name)" json:"name"`
	Description    string `orm:"column(description)" json:"description"`
	Provider       string `orm:"column(provider)" json:"provider"`
	Endpoint       string `orm:"column(endpoint)" json:"endpoint"`
	AuthMode       string `orm:"column(auth_mode)" json:"auth_mode"`
	AuthData       string `orm:"column(auth_data)" json:"auth_data"`
	Status         string `orm:"column(status)" json:"status"`
	Enabled        bool   `orm:"column(enabled)" json:"enabled"`
	SetupTimestamp int64  `orm:"column(setup_timestamp)" json:"setup_timestamp"`
	Extensions     string `orm:"column(extensions)" json:"extensions"`
}

// TableName set table name for ORM.
func (i *Instance) TableName() string {
	return "p2p_preheat_instance"
}
