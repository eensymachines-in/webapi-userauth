package nosql

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Queries on the database and the interfaces to configure and dial to connections
This needs to be moved to a separate package later
============================================== */
import (
	"fmt"
	"reflect"

	"gopkg.in/mgo.v2/bson"
)

// IDBConn : basic nosql database behaviours
// can init, close and dial connections to databases
type IDBConn interface {
	InitConn(host, db, uname, pass string) IDBConn // initializes the db object with dialinfo
	DialConn() error                               // dials the connection and establishes contact, use IAnyDB to setup
	CloseConn()
}

// IQryable : basic mongoquery beahviour facilitating crud operations
type IQryable interface {
	AddToColl(obj interface{}, coll string) (int, error) // capable of adding one or more items to the same collection
	GetOneFromColl(coll string, flt func() bson.M, result *map[string]interface{}) error
	GetSampleFromColl(coll string, size int, result *interface{}) error
	EditOneFromColl(coll string, flt func() bson.M, result interface{}) error
	DeleteOneFromColl(coll, id string, flt func(id string) bson.M, result *interface{}) error
	CountFromColl(coll string, flt func() bson.M) (int, error)
}

// InitDB : creates an instance of the DB service
// sends the same back over the interface IDBConn for testing the connection
//
/*

 */
func InitDB(host, db, uname, pass string, ty reflect.Type) IDBConn {
	dbAsItf := reflect.New(ty.Elem()).Interface()
	dbAsItf.(IDBConn).InitConn(host, db, uname, pass)
	return dbAsItf.(IDBConn)
}

// DialConnectDB : Calls the extensible Dial method on the DB and checks for errors.
// Use this after you have InitDB. Incase the Dial is success it sends back the DB over IDBConn interface.
// Check for the error to know if the Dial has failed.
//
// NOTE: unless initiated this function would not work
/*
	db, close ,err := DialConnectDB(InitDB(host, db, coll, reflect.TypeOf(&MongoDB)))
	if err !=nil{
		return err
	}
	defer close()
	// call the queries on persistence interfaces

*/

func DialConnectDB(conn IDBConn) (IDBConn, func(), error) {
	// with the given fields all what you need is to see if ping works
	// if ping works, then shift the handle to IMongoQry or send back error
	if conn.DialConn() != nil {
		return conn, nil, fmt.Errorf("failed to connect to DB")
	}
	return conn, func() {
		// to be called from client side
		conn.CloseConn()
	}, nil
}
