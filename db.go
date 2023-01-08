package useracc

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Queries on the database and the interfaces to configure and dial to connections
============================================== */
import (
	"fmt"
	"reflect"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Database implementations and interfaces
This needs to be moved to a separate package later
============================================== */

// IDBConn :
type IDBConn interface {
	Init(host, db, uname, pass string) IDBConn // initializes the db object with dialinfo
	Dial() error                               // dials the connection and establishes contact, use IAnyDB to setup
	Close()
}
type IMongoQry interface {
	AddOneToColl(obj interface{}, coll string) (int, error)
	AddBulkToColl(list interface{}, coll string) (int, error)
	GetOneFromColl(coll string, flt func() bson.M, result interface{}) error
	GetSampleFromColl(coll string, size int, result *interface{}) error
	EditOneFromColl(coll string, flt func() bson.M, result interface{}) error
	DeleteOneFromColl(coll, id string, flt func(id string) bson.M, result *interface{}) error
}
type MongoDB struct {
	dialInfo *mgo.DialInfo // for connecting to the databse
	session  *mgo.Session  // this reference can help us run queries
}

func (mgdb *MongoDB) Dial() error {
	if mgdb.dialInfo != nil {
		s, err := mgo.DialWithInfo(mgdb.dialInfo)
		if err != nil {
			return err
		}
		// Reads may not be entirely up-to-date, but they will always see the
		// history of changes moving forward, the data read will be consistent
		// across sequential queries in the same session, and modifications made
		// within the session will be observed in following queries (read-your-writes).
		// http://godoc.org/labix.org/v2/mgo#Session.SetMode
		s.SetMode(mgo.Monotonic, true)
		mgdb.session = s
		return nil
	}
	return fmt.Errorf("empty dialinfo, check and dial again")

}
func (mgdb *MongoDB) Close() {
	mgdb.session.Close()
}
func (mgdb *MongoDB) Init(host, db, uname, pass string) IDBConn {
	mgdb.dialInfo = &mgo.DialInfo{
		Addrs:    []string{host},
		Timeout:  3 * time.Second,
		Database: db,
		Username: uname,
		Password: pass,
	}
	mgdb.session = nil
	return mgdb
}

func (mgdb *MongoDB) GetOneFromColl(coll string, flt func() bson.M, result interface{}) error {
	return nil
}

// AddOneToColl : this will add one data to the collection
// DB name is not required since empty string will lead to calling the database thats used for dialling within the dialinfo
// error only when the query fails
func (mgdb *MongoDB) AddOneToColl(obj interface{}, coll string) (int, error) {
	if err := mgdb.session.DB("").C(coll).Insert(obj); err != nil {
		return 0, fmt.Errorf("%s: failed to add one to collection", err)
	}
	return 1, nil
}
func (mgdb *MongoDB) EditOneFromColl(coll string, flt func() bson.M, result interface{}) error {
	return nil
}
func (mgdb *MongoDB) DeleteOneFromColl(coll, id string, flt func(id string) bson.M, result *interface{}) error {
	if err := mgdb.session.DB("").C(coll).Remove(flt(id)); err != nil {
		return fmt.Errorf("%s: failed to remove from collection", err)
	}
	*result = map[string]interface{}{"ok": 1}
	return nil
}
func (mgdb *MongoDB) AddBulkToColl(list interface{}, coll string) (int, error) {
	if err := mgdb.session.DB("").C(coll).Insert(list); err != nil {
		return 0, fmt.Errorf("%s: failed to bulk add to the collection", err)
	}
	return 1, nil
}

// GetSampleFromColl : gets a sample of documents from a collection
// sends back the _id object ids as list of ids in a map
// result["sample"] gives the result of all the ids
func (mgdb *MongoDB) GetSampleFromColl(coll string, size int, result *interface{}) error {
	res := map[string]interface{}{}
	mgdb.session.DB("").C(coll).Pipe([]bson.M{
		{"$sample": bson.M{"size": size}},
		{"$project": bson.M{"_id": 1}},
		{"$group": bson.M{"_id": "", "sample": bson.M{"$push": "$_id"}}},
		{"$project": bson.M{"_id": 0, "sample": 1}},
	}).All(&res)
	*result = res
	return nil
}

// InitDB : creates an instance of the DB service
// sends the same back over the interface IDBConn for testing the connection
//
/*

 */
func InitDB(host, db, uname, pass string, ty reflect.Type) IDBConn {
	dbAsItf := reflect.New(ty.Elem()).Interface()
	dbAsItf.(IDBConn).Init(host, db, uname, pass)
	return dbAsItf.(IDBConn)
}

// DialConnectDB : Calls the extensible Dial method on the DB and checks for errors.
// Use this after you have InitDB. Incase the Dial is success it sends back the DB over IDBConn interface.
// Check for the error to know if the Dial has failed.
//
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
	if conn.Dial() != nil {
		return conn, nil, fmt.Errorf("failed to connect to DB")
	}
	return conn, func() {
		// to be called from client side
		conn.Close()
	}, nil
}
