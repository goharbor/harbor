package dao

import "testing"

func Test_Mysql_UpgradeSchema(t *testing.T) {
	mysql := NewMySQL("10.0.2.5","3306","root","root123","registry",2,0)
	mysql.Register()
	mysql.UpgradeSchema()
}