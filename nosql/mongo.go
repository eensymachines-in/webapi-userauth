package nosql

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Generic mongo queries. Will implement callbacks on most of the funcs, so as to let the client set the filtering client
============================================== */
import (
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	MSG_SERVR_BROKEN = "Something on server is broken right now, try after some time"
	MSG_DATA_FAIL    = "One or more failed operation(s) on data, try after some time"
	MSG_DATA_INVLD   = "Invalid data, aborting operations on server"
)

// derives from session object
// stores references to 2 active collections
type MongoDB struct {
	*mgo.Session                 // this reference can help us run queries
	DefaultColl  *mgo.Collection // default collection from which data is sought
	ArchiveColl  *mgo.Collection // soft delete operations will shift the data to archive coll
}

// DialConn		: will use the DB's configuration to dial the connection.
// Error if DialWithInfo fails internally
//
/*
config := &nosql.DBInitConfig{
	Host:      "localhost:37017",
	DB:        "db_name",
	UName:     "user_name",
	Passwd:    "secret_pass",
	Coll:      "coll_name",
	ArchvColl: "archival_coll",
	DBTyp:     reflect.TypeOf(&nosql.MongoDB{}),
}
if err := conn.DialConn(cfg); err != nil {
	return err
}
*/
func (mgdb *MongoDB) DialConn(c *DBInitConfig) error {
	// Dials an mgo connection with DBInitConfig
	// error when session is unreachable
	var err error
	mgdb.Session, err = mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{c.Host},
		Timeout:  3 * time.Second,
		Database: c.DB,
		Username: c.UName,
		Password: c.Passwd,
	})
	if err != nil {
		return ThrowErrNoSQL(ErrNoConn).SetContext("DialConn").SetInternalErr(err).SetDiagnosis("check for the connection params").SetUsrMsg(MSG_SERVR_BROKEN).SetLogEntry(log.WithFields(log.Fields{
			"host":  c.Host,
			"db":    c.DB,
			"uname": c.UName,
			"pass":  c.Passwd,
		}))
	}
	// Reads may not be entirely up-to-date, but they will always see the
	// history of changes moving forward, the data read will be consistent
	// across sequential queries in the same session, and modifications made
	// within the session will be observed in following queries (read-your-writes).
	// http://godoc.org/labix.org/v2/mgo#Session.SetMode
	mgdb.Session.SetMode(mgo.Monotonic, true)
	// and also assign the collections
	mgdb.DefaultColl = mgdb.Session.DB(c.DB).C(c.Coll)
	mgdb.ArchiveColl = mgdb.Session.DB(c.DB).C(c.ArchvColl)
	return nil

}
func (mgdb *MongoDB) CloseConn() {
	mgdb.Close()
}

// GetOneFromColl 	: irrespective of the collection this one calls to get on specific from the collection
// result			:need not instantiate  before calling
// Provide the name of the collection and the flt as callback
// errors if the query fails, if the collection name is invalid, and if the filter is nil
func (mgdb *MongoDB) GetOneFromColl(coll string, flt func() bson.M, result *map[string]interface{}) error {
	if coll == "" {
		// return fmt.Errorf("GetOneFromColl: Invalid collection name or filter function")
		return ThrowErrNoSQL(ErrInvldColl).SetContext("GetOneFromColl").SetInternalErr(nil).SetDiagnosis("check for name of collection").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"coll": coll,
		}))
	}
	if flt == nil {
		return ThrowErrNoSQL(ErrInvldFlt).SetContext("GetOneFromColl").SetInternalErr(nil).SetDiagnosis("check for name of collection").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"filter": flt,
		}))
	}
	qr := map[string]interface{}{} // result map onto which the object is imprinted
	err := mgdb.DB("").C(coll).Find(flt()).One(&qr)
	if err != nil {
		return ThrowErrNoSQL(ErrNoConn).SetContext("GetOneFromColl").SetInternalErr(nil).SetDiagnosis("check database connection").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"err": err,
		}))
	}
	*result = qr
	return nil
}

// AddToColl : this will add one data to the collection
// DB name is not required since empty string will lead to calling the database thats used for dialling within the dialinfo
// error only when the query fails
func (mgdb *MongoDB) AddToColl(obj interface{}, coll string) (int, error) {
	if coll == "" {
		return -1, ThrowErrNoSQL(ErrInvldColl).SetContext("AddToColl").SetInternalErr(nil).SetDiagnosis("check for name of collection").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"coll": coll,
		}))
	}
	if obj == nil {
		return -1, ThrowErrNoSQL(ErrEmptyInsert).SetContext("AddToColl").SetInternalErr(nil).SetDiagnosis("object to insert cannot be nil").SetUsrMsg(MSG_DATA_INVLD)
	}
	if err := mgdb.DB("").C(coll).Insert(obj); err != nil {
		return -1, ThrowErrNoSQL(ErrQryFail).SetContext("AddToColl").SetInternalErr(err).SetDiagnosis("check inputs to query and try again").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"err": err,
		}))
	}
	return 1, nil
}

func (mgdb *MongoDB) EditOneFromColl(coll string, flt, patch func() bson.M, countUpdated *int) error {
	if coll == "" {
		return ThrowErrNoSQL(ErrInvldColl).SetContext("EditOneFromColl").SetInternalErr(nil).SetDiagnosis("check for name of collection").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"coll": coll,
		}))
	}
	if flt == nil || patch == nil {
		return ThrowErrNoSQL(ErrInvldFlt).SetContext("EditOneFromColl").SetInternalErr(nil).SetDiagnosis("check for query input params").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"flt":   flt,
			"patch": patch,
		}))
	}
	change, err := mgdb.DB("").C(coll).UpdateAll(flt(), patch())
	if err != nil {
		return ThrowErrNoSQL(ErrQryFail).SetContext("EditOneFromColl").SetInternalErr(err).SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"err": err,
		}))
	}
	*countUpdated = change.Updated
	return nil
}

// GetSampleFromColl : gets a sample of documents from a collection
// sends back the _id object ids as list of ids in a map
// result["sample"] gives the result of all the ids
// result is of the type map[string][]bson.ObjectId{}
//
/*
	var result interface{}
	err = db.(nosql.IQryable).GetSampleFromColl(COLL_NAME, 10, &result)
	if err != nil {
		return
	}
	ids := result.(map[string][]bson.ObjectId)
	for _, id := range ids["sample"]{
		fmt.Sprintf("%s", id.Hex())
	}
*/
func (mgdb *MongoDB) GetSampleFromColl(coll string, size uint32, result *interface{}) error {
	if coll == "" {
		return ThrowErrNoSQL(ErrInvldColl).SetContext("AddToColl").SetInternalErr(nil).SetDiagnosis("check for name of collection").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"coll": coll,
		}))
	}
	res := map[string][]bson.ObjectId{}
	err := mgdb.DB("").C(coll).Pipe([]bson.M{
		{"$sample": bson.M{"size": size}},
		{"$project": bson.M{"_id": 1}},
		{"$group": bson.M{"_id": "", "sample": bson.M{"$push": "$_id"}}},
		{"$project": bson.M{"_id": 0, "sample": 1}},
	}).One(&res)
	if err != nil {
		return ThrowErrNoSQL(ErrQryFail).SetContext("GetSampleFromColl").SetInternalErr(err).SetDiagnosis("").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"err": err,
		}))
	}
	*result = res
	return nil
}

// CountFromColl : for the given filter on the collection this can get count of documents
// coll		: ithe collection on which the filter applies
// flt		: filter the documents on the collection using this
func (mgdb *MongoDB) CountFromColl(coll string, flt func() bson.M) (int, error) {
	if coll == "" {
		return -1, ThrowErrNoSQL(ErrInvldColl).SetContext("CountFromColl").SetInternalErr(nil).SetDiagnosis("check for name of collection").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"coll": coll,
		}))
	}
	n, err := mgdb.DB("").C(coll).Find(flt()).Count()
	if err != nil {
		return -1, ThrowErrNoSQL(ErrQryFail).SetContext("CountFromColl").SetInternalErr(err).SetDiagnosis("query failed").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"coll": coll,
		}))
	}
	return n, nil
}

func (mgdb *MongoDB) RemoveFromColl(coll string, id string, softDel bool, affected *int) error {
	// remove from the main database , and add into the archived data base
	// need to see what details are pushed onto the archived database
	if softDel {
		// NOTE: this will backup the document in archival collection and only then delete
		// NOTE: leaves us with the room to restore documents when needed
		result := map[string]interface{}{}
		err := mgdb.GetOneFromColl(coll, func() bson.M {
			return bson.M{"_id": bson.ObjectIdHex(id)}
		}, &result)
		if err != nil {
			return ThrowErrNoSQL(ErrQryFail).SetContext("MongoDB.RemoveFromColl").SetInternalErr(err).SetDiagnosis("could not get document to delete").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
				"err": err,
			}))
		}
		// Now push the resul/t to archival collection
		if err := mgdb.ArchiveColl.Insert(result); err != nil {
			return ThrowErrNoSQL(ErrQryFail).SetContext("MongoDB.RemoveFromColl").SetInternalErr(err).SetDiagnosis("failed query on inserting into archive coll").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
				"coll": coll,
			}))
		}
	}
	if err := mgdb.DB("").C(coll).Remove(bson.M{"_id": bson.ObjectIdHex(id)}); err != nil {
		return ThrowErrNoSQL(ErrQryFail).SetContext("MongoDB.RemoveFromColl").SetInternalErr(err).SetDiagnosis("query failed").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"err": err,
		}))
	}
	*affected = 1
	return nil
}

// FilterFromColl : While GetOneFromColl is implemented to get one item this applies filter to get multiple items
// filter here can be customized from the client call
// TODO: come back here to review this and test it
func (mgdb *MongoDB) FilterFromColl(coll string, flt func() bson.M, result *map[string][]bson.ObjectId) error {
	if coll == "" {
		return ThrowErrNoSQL(ErrInvldColl).SetContext("FilterFromColl").SetInternalErr(nil).SetDiagnosis("check for name of collection").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"coll": coll,
		}))
	}
	if flt == nil {
		return ThrowErrNoSQL(ErrInvldFlt).SetContext("GetOneFromColl").SetInternalErr(nil).SetDiagnosis("filter for the query cannot be nil").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"filter": flt,
		}))
	}
	res := map[string][]bson.ObjectId{}
	err := mgdb.DB("").C(coll).Pipe([]bson.M{
		{"$match": flt()},
		// matches the documents and then pushes them in a slice
		// this can help get the ids all clubbed in one
		{"$group": bson.M{"_id": "", "all": bson.M{"$push": "$_id"}}},
		{"$project": bson.M{"_id": 0, "all": 1}},
	}).One(&res) //will err when no docs found
	if err != nil {
		if err == mgo.ErrNotFound {
			return ThrowErrNoSQL(ErrEmptyResult).SetContext("MongoDB.FilterFromColl").SetInternalErr(err).SetDiagnosis("empty result from filter").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
				"err": err,
			}))
		}
		return ThrowErrNoSQL(ErrQryFail).SetContext("MongoDB.FilterFromColl").SetInternalErr(err).SetDiagnosis("query failed").SetUsrMsg(MSG_DATA_FAIL).SetLogEntry(log.WithFields(log.Fields{
			"err": err,
		}))
	}
	*result = res
	return nil
}
