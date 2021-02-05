package dao

import "testing"

func Test_Mysql_UpgradeSchema(t *testing.T) {
	mysql := NewMySQL("localhost","3306","root","123456","registry",2,0)
	mysql.Register()
	mysql.UpgradeSchema()
}