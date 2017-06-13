// Some monstrosity to make sql support the ipfs-datastore interface.
// Work. In. Progress.
package sql_datastore

import (
	"database/sql"
	"fmt"
	datastore "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

type Datastore struct {
	DB     *sql.DB
	models []Model
}

func NewDatastore(db *sql.DB) *Datastore {
	return &Datastore{DB: db}
}

func (ds *Datastore) Register(models ...Model) error {
	for _, model := range models {
		// TODO - sanity check to make sure the model behaves.
		// return error if not
		ds.models = append(ds.models, model)
	}
	return nil
}

func (ds Datastore) Put(key datastore.Key, value interface{}) error {
	sqlModelValue, ok := value.(Model)
	if !ok {
		return fmt.Errorf("value is not a valid sql model")
	}

	exists, err := ds.hasModel(sqlModelValue)
	if err != nil {
		return err
	}

	if exists {
		return ds.exec(sqlModelValue, CmdUpdateOne)
	} else {
		return ds.exec(sqlModelValue, CmdInsertOne)
	}
}

func (ds Datastore) Get(key datastore.Key) (value interface{}, err error) {
	m, err := ds.modelForKey(key)
	if err != nil {
		return nil, err
	}

	row, err := ds.queryRow(m, CmdSelectOne)
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

	row, err := ds.queryRow(m, CmdExistsOne)
	if err != nil {
		return false, err
	}

	err = row.Scan(&exists)
	return
}

func (ds Datastore) hasModel(m Model) (exists bool, err error) {
	row, err := ds.queryRow(m, CmdExistsOne)
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

	return ds.exec(m, CmdDeleteOne)
}

func (ds Datastore) modelForKey(key datastore.Key) (Model, error) {
	for _, m := range ds.models {
		if m.DatastoreType() == key.Type() {
			// return a model with "ID" set to the key param
			return m.NewSQLModel(key.Name()), nil
		}
	}
	return nil, fmt.Errorf("no usable model found for key, did you call register on the model?: %s", key.String())
}

func (ds Datastore) exec(m Model, t Cmd) error {
	query, params, err := ds.prepQuery(m, t)
	if err != nil {
		return err
	}
	_, err = ds.DB.Exec(query, params...)
	return err
}

func (ds Datastore) queryRow(m Model, t Cmd) (*sql.Row, error) {
	query, params, err := ds.prepQuery(m, t)
	if err != nil {
		return nil, err
	}
	return ds.DB.QueryRow(query, params...), nil
}

func (ds Datastore) query(m Model, t Cmd, prebind ...interface{}) (*sql.Rows, error) {
	query, params, err := ds.prepQuery(m, t)
	if err != nil {
		return nil, err
	}
	return ds.DB.Query(query, append(prebind, params...)...)
}

func (ds Datastore) prepQuery(m Model, t Cmd) (string, []interface{}, error) {
	query := m.SQLQuery(t)
	if query == "" {
		// TODO - make Cmd satisfy stringer, provide better error
		return "", nil, fmt.Errorf("missing required command: %d", t)
	}
	params := m.SQLParams(t)
	return query, params, nil
}

// Ok, this is nothing more than a first step. In the future
// it seems datastore will need to construct these queries, which
// will require more info (tablename, expected response schema) from
// the model.
// Currently it's required that the passed-in prefix be equal to DatastoreType()
// which query will use to determine what model to ask for a ListCmd
func (ds Datastore) Query(q query.Query) (query.Results, error) {
	// TODO - support query Filters
	if len(q.Filters) > 0 {
		return nil, fmt.Errorf("sql datastore queries do not support filters")
	}
	// TODO - support query Orders
	if len(q.Orders) > 0 {
		return nil, fmt.Errorf("sql datastore queries do not support ordering")
	}
	// TODO - support KeysOnly
	if q.KeysOnly {
		return nil, fmt.Errorf("sql datastore doesn't support keysonly ordering")
	}

	// TODO - ugh this so bad
	m, err := ds.modelForKey(datastore.NewKey(fmt.Sprintf("/%s:", q.Prefix)))
	if err != nil {
		return nil, err
	}

	// This is totally janky, but will work for now. It's expected that
	// the returned CmdList will have at least 2 bindvars:
	// $1 : LIMIT
	// $2 : OFFSET
	// From there it can provide zero or more additional bindvars to
	// organize the query, which should be returned by the SQLParams method
	// TODO - this seems to hint at a need for some sort of Controller-like
	// pattern in userland. Have a think.
	rows, err := ds.query(m, CmdList, q.Limit, q.Offset)
	if err != nil {
		return nil, err
	}

	// TODO - should this be q.Limit or query.NormalBufferSize
	reschan := make(chan query.Result, q.Limit)
	go func() {
		defer close(reschan)

		for rows.Next() {

			model := m.NewSQLModel("")
			if err := model.UnmarshalSQL(rows); err != nil {
				reschan <- query.Result{
					Error: err,
				}
			}

			reschan <- query.Result{
				Entry: query.Entry{
					Key:   m.GetId(),
					Value: model,
				},
			}

		}
	}()

	return query.ResultsWithChan(q, reschan), nil
}

func (ds *Datastore) Batch() (datastore.Batch, error) {
	return nil, datastore.ErrBatchUnsupported
}
