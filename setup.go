package geaves

import (
	_ "embed"
)

//go:embed sql/tables.sql
var tableDefs string

//go:embed sql/reset.sql
var resetStmts string

func SetupSQL() string {
	return tableDefs
}

func ResetSQL() string {
	return resetStmts
}
