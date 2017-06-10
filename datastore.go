// Some monstrosity to make sql support the ipfs-datastore interface.
// Work. In. Progress.
package sql_datastore

import (
	"database/sql"
	"fmt"
	"github.com/archivers-space/sqlutil"
	datastore "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

type Datastore struct {
	DB     *sql.DB
	models []sqlutil.Model
}

func NewDatastore(db *sql.DB) *Datastore {
	return &Datastore{DB: db}
}

func (ds *Datastore) Register(models ...sqlutil.Model) error {
	for _, model := range models {
		// TODO - sanity check to make sure the model behaves.
		// return error if not
		ds.models = append(ds.models, model)
	}
	return nil
}

func (ds Datastore) Put(key datastore.Key, value interface{}) error {
	sqlModelValue, ok := value.(sqlutil.Model)
	if !ok {
		return fmt.Errorf("value is not a valid sql model")
	}

	exists, err := ds.Has(key)
	if err != nil {
		return err
	}

	if exists {
		return ds.exec(sqlModelValue, sqlutil.CmdUpdateOne)
	} else {
		return ds.exec(sqlModelValue, sqlutil.CmdInsertOne)
	}
}

func (ds Datastore) Get(key datastore.Key) (value interface{}, err error) {
	m, err := ds.modelForKey(key)
	if err != nil {
		return nil, err
	}

	row, err := ds.queryRow(m, sqlutil.CmdSelectOne)
	if err != nil {
		return nil, err
	}

	v := m.NewSQLModel(key.Name())
	if err := v.UnmarshalSQL(row); err != nil {
		return nil, err
	}
	return v, nil
}

func (ds Datastore) Has(key datastore.Key) (exists bool, err error) {
	m, err := ds.modelForKey(key)
	if err != nil {
		return false, err
	}

	row, err := ds.queryRow(m, sqlutil.CmdExistsOne)
	if err != nil {
		return false, err
	}

	err = row.Scan(&exists)
	return
}

func (ds Datastore) Delete(key datastore.Key) error {
	m, err := ds.modelForKey(key)
	if err != nil {
		return err
	}

	return ds.exec(m, sqlutil.CmdDeleteOne)
}

func (ds Datastore) modelForKey(key datastore.Key) (sqlutil.Model, error) {
	for _, m := range ds.models {
		if m.DatastoreType() == key.Type() {
			// return a model with "ID" set to the key param
			return m.NewSQLModel(key.Name()), nil
		}
	}
	return nil, fmt.Errorf("no usable model found for key, did you call register on the model?: %s", key.String())
}

func (ds Datastore) exec(m sqlutil.Model, t sqlutil.CmdType) error {
	query, params, err := ds.prepQuery(m, t)
	if err != nil {
		return err
	}
	_, err = ds.DB.Exec(query, params...)
	return err
}

func (ds Datastore) queryRow(m sqlutil.Model, t sqlutil.CmdType) (*sql.Row, error) {
	query, params, err := ds.prepQuery(m, t)
	if err != nil {
		return nil, err
	}
	return ds.DB.QueryRow(query, params...), nil
}

func (ds Datastore) query(m sqlutil.Model, t sqlutil.CmdType) (*sql.Rows, error) {
	query, params, err := ds.prepQuery(m, t)
	if err != nil {
		return nil, err
	}
	return ds.DB.Query(query, params...)
}

func (ds Datastore) prepQuery(m sqlutil.Model, t sqlutil.CmdType) (string, []interface{}, error) {
	query := m.SQLQuery(t)
	if query == "" {
		// TODO - make sqlutil.CmdType satisfy stringer, provide better error
		return "", nil, fmt.Errorf("missing required command: %d", t)
	}
	params := m.SQLParams(t)
	return query, params, nil
}

// lolololol wut
func (ds Datastore) Query(q query.Query) (query.Results, error) {
	return nil, fmt.Errorf("querying SQL datastore not supported")
}

func (ds *Datastore) Batch() (datastore.Batch, error) {
	return nil, datastore.ErrBatchUnsupported
}
