package dao

import (
	"fmt"
	"strings"
)

// JoinNumberConditions - To join number condition into string,used in sql query
func JoinNumberConditions(ids []int) string {
	return strings.Trim(strings.Replace(fmt.Sprint(ids), " ", ",", -1), "[]")
}
