package sql_datastore

import (
	"github.com/archivers-space/sqlutil"
)

type Cmd int

const (
	CmdUnknown Cmd = iota
	CmdCreateTable
	CmdAlterTable
	CmdDropTable
	CmdSelectOne
	CmdInsertOne
	CmdUpdateOne
	CmdDeleteOne
	CmdExistsOne
	CmdList
)

type Model interface {
	Unmarshaller
	Commandable

	NewSQLModel(id string) Model

	DatastoreType() string
	GetId() string
}

type Unmarshaller interface {
	UnmarshalSQL(sqlutil.Scannable) error
}

type Commandable interface {
	SQLQuery(Cmd) string
	SQLParams(Cmd) []interface{}
}
