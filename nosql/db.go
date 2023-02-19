package nosql

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Queries on the database and the interfaces to configure and dial to connections
This needs to be moved to a separate package later
============================================== */
import (
	"reflect"

	"gopkg.in/mgo.v2/bson"
)

// IDBConn : basic nosql database behaviours
// can init, close and dial connections to databases
type IDBConn interface {
	// InitConn(host, db, uname, pass string) IDBConn // initializes the db object with dialinfo
	DialConn(*DBInitConfig) error // dials the connection and establishes contact, use IAnyDB to setup
	CloseConn()
}

// IQryable : basic mongoquery beahviour facilitating crud operations
type IQryable interface {
	// AddToColl : this will add one data to the collection
	// DB name is not required since empty string will lead to calling the database thats used for dialling within the dialinfo
	// error only when the query fails
	// no check for duplicacy - or validation on document fields.
	/*
		toAdd := bson.M{}
		added , err = db.(nosql.IQryable).AddToColl(toAdd,COLL_NAME)
		if err != nil {
			return
		}
		fmt.Printf("added %d items to the collection", added)
	*/
	AddToColl(obj interface{}, coll string) (int, error) // capable of adding one or more items to the same collection
	// RemoveFromColl : will remove item from the collection
	// Whether or not the item will be maintained in the archive collection is contextual to the business logic
	// coll : name of the collection
	// id 	: id as string, cannot be bson id since the interface applies to various nosql databases
	// affected	: number of documents affected
	// softDel	: set this flag to only archive the document and not delete it completely
	/*
		var count int
		removed, err := db.(nosql.IQryable).RemoveFromColl(COLL_NAME, id.Hex(), true, &count)
		if err != nil {
			return
		}
		fmt.Printf("deleted %d items to the collection", removed)
	*/
	RemoveFromColl(coll string, id string, softDel bool, affected *int) error
	// GetOneFromColl 	: irrespective of the collection this one calls to get on specific from the collection
	// result			:need not instantiate  before calling
	// Provide the name of the collection and the flt as callback
	// errors if the query fails, if the collection name is invalid, and if the filter is nil
	/*
		var result interface{}
		db.(nosql.IQryable).GetSampleFromColl(COLL_NAME, 1, &result)
		ids := result.(map[string][]bson.ObjectId)
		sample := ids["sample"]
		assert.Equal(t, 1, len(sample), "Unexpected number of items in the sample")
	*/
	GetOneFromColl(coll string, flt func() bson.M, result *map[string]interface{}) error
	// GetSampleFromColl : gets a sample of documents from a collection
	// sends back the _id object ids as list of ids in a map
	// result["sample"] gives the result of all the ids
	// result is of the type map[string][]bson.ObjectId{}
	//
	/*
		ua := useracc.UserAccount{}
		var uaMap map[string]interface{}
		err = db.(nosql.IQryable).GetOneFromColl(COLL_NAME, func() bson.M { return bson.M{"_id": sample[0]} }, &uaMap)
	*/
	GetSampleFromColl(coll string, size uint32, result *interface{}) error
	// EditOneFromColl : patching documents on selection
	// updates many documents at once - depends on the flt()
	// patch :  callback that sends out bson.M{} only to $set clauses
	//
	/*
		err = db.(nosql.IQryable).EditOneFromColl(COLL_NAME, func() bson.M {
			return bson.M{
				"email": "cdobrowski0@pcworld.com",
			} // selection filter
		}, func() bson.M {
			return bson.M{
				"$set": bson.M{"title": newTitle},
			} // setting action
		}, &count)
	*/
	EditOneFromColl(coll string, flt, patch func() bson.M, countUpdated *int) error
	// CountFromColl : for the given filter on the collection this can get count of documents
	// coll		: ithe collection on which the filter applies
	// flt		: filter the documents on the collection using this
	CountFromColl(coll string, flt func() bson.M) (int, error)
	// FilterFromColl : filters documents on custom filter , returns a slice of ids of such documents
	// Use GetOneFromColl to get detailed document object
	/*
		var result *map[string][]bson.ObjectId
		if db.FilterFromColl("collname", 10, &result)!=nil{
			fmt.Errorf("failed to get filtered documents from database")
		}
		return nil
	*/
	FilterFromColl(coll string, flt func() bson.M, result *map[string][]bson.ObjectId) error
}

// DBInitConfig : flywheel object that gets passed to InitDB for making a new DB instance
// extend this object if required to send in more params
type DBInitConfig struct {
	// server ip where mongo instance is running with port
	// localhost:47017
	Host string
	// Name of the DB that client connects to
	DB string
	// Username and Password for the authenticating on the database
	UName  string
	Passwd string
	// Set  of 2 collection name, one for default, other for archival
	Coll      string
	ArchvColl string
	// type of the DB that will be instantiated under IDBConn interface
	DBTyp reflect.Type
}

// InitDialConn : will make instance and dial the connection to database
// will send the connection over IDBConn with errors if any
func InitDialConn(cfg *DBInitConfig) (IDBConn, func(), error) {
	// IMP: since reflection, no chance of catching errors related to interface non-conformance at build time
	// mongodb database is not a type IQry - usually a build time error
	dbAsItf := reflect.New(cfg.DBTyp.Elem()).Interface()
	conn := dbAsItf.(IDBConn)
	if err := conn.DialConn(cfg); err != nil {
		return nil, nil, err
	}
	return conn, func() {
		conn.CloseConn()
	}, nil
}
