package dao

import "github.com/astaxie/beego/orm"

func paginateForQuerySetter(qs orm.QuerySeter, page, size int64) orm.QuerySeter {
	if size > 0 {
		qs = qs.Limit(size)
		if page > 0 {
			qs = qs.Offset((page - 1) * size)
		}
	}
	return qs
}
