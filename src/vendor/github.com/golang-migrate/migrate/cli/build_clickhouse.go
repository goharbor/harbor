// +build clickhouse

package main

import (
	_ "github.com/kshvakov/clickhouse"
	_ "github.com/golang-migrate/migrate/database/clickhouse"
)
