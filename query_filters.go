package sql_datastore

import (
	"fmt"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

// Filter
type FilterKeyTypeEq string

func (f FilterKeyTypeEq) Key() datastore.Key {
	return datastore.NewKey(fmt.Sprintf("/%s:", f.String()))
}

func (f FilterKeyTypeEq) String() string {
	return string(f)
}

// TODO - make this work properly for the sake of other datastores
func (f FilterKeyTypeEq) Filter(e query.Entry) bool {
	return true
}
