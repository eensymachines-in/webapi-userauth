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
	AddToColl(obj interface{}, coll string) (int, error) // capable of adding one or more items to the same collection
	// RemoveFromColl : will remove item from the collection
	// Whether or not the item will be maintained in the archive collection is contextual to the business logic
	// coll : name of the collection
	// id 	: id as string, cannot be bson id since the interface applies to various nosql databases
	// affected	: number of documents affected
	// softDel	: set this flag to only archive the document and not delete it completely
	RemoveFromColl(coll string, id string, softDel bool, affected *int) error
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
	GetSampleFromColl(coll string, size uint32, result *interface{}) error
	EditOneFromColl(coll string, flt func() bson.M, result interface{}) error
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
	dbAsItf := reflect.New(cfg.DBTyp.Elem()).Interface()
	conn := dbAsItf.(IDBConn)
	if err := conn.DialConn(cfg); err != nil {
		return nil, nil, err
	}
	return conn, func() {
		conn.CloseConn()
	}, nil
}
