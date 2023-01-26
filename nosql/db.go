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

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type IDBConfig interface {
	SetDefaultColl(c *mgo.Collection) IDBConfig
	SetArchiveColl(c *mgo.Collection) IDBConfig
}

// IDBConn : basic nosql database behaviours
// can init, close and dial connections to databases
type IDBConn interface {
	InitConn(host, db, uname, pass string) IDBConn // initializes the db object with dialinfo
	DialConn(coll, archv string) error             // dials the connection and establishes contact, use IAnyDB to setup
	CloseConn()
}

// IQryable : basic mongoquery beahviour facilitating crud operations
type IQryable interface {
	AddToColl(obj interface{}, coll string) (int, error) // capable of adding one or more items to the same collection
	// RemoveFromColl : will remove item from the collection
	// Whether or not the item will be maintained in the archive collection is contextual to the business logic
	// coll : name of the collection
	// id 	: id as string, cannot be bson id since the interface applies to various nosql databases
	// affected	: number of documents affected
	RemoveFromColl(coll string, id string, affected *int) error
	GetOneFromColl(coll string, flt func() bson.M, result *map[string]interface{}) error
	// GetSampleFromColl : gets a sample of documents from a collection
	// sends back the _id object ids as list of ids in a map
	// result["sample"] gives the result of all the ids
	// result is of the type map[string][]bson.ObjectId{}
	//
	/*
		var result map[string][]bson.ObjectId
		if db.GetSampleFromColl("collname", 10, &result)!=nil{
			fmt.Errorf("failed to get sample from database")
		}
		return nil
	*/
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

//DialConnectDB		: given the instantiated connection will dial the connection
// conn				: instantiated connection from InitDB
//
/*
persist, close, err := nosql.DialConnectDB(db, COLL_NAME, ARCHVCOLL_NAME)
*/
func DialConnectDB(conn IDBConn, coll, archv string) (IDBConn, func(), error) {
	// with the given fields all what you need is to see if ping works
	// if ping works, then shift the handle to IMongoQry or send back error
	if conn.DialConn(coll, archv) != nil {
		return conn, nil, fmt.Errorf("failed to connect to DB")
	}
	return conn, func() {
		// to be called from client side
		conn.CloseConn()
	}, nil
}
