package nosql

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Generic mongo queries. Will implement callbacks on most of the funcs, so as to let the client set the filtering client
============================================== */
import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoConfig struct {
	*mgo.DialInfo                 // we just extend the dial info so as to add more meat
	DefaultColl   *mgo.Collection // default collection from which data is sought
	ArchiveColl   *mgo.Collection // soft delete operations will shift the data to archive coll
}

func (mcfg *MongoConfig) SetDefaultColl(c *mgo.Collection) IDBConfig {
	mcfg.DefaultColl = c
	return mcfg
}
func (mcfg *MongoConfig) SetArchiveColl(c *mgo.Collection) IDBConfig {
	mcfg.ArchiveColl = c
	return mcfg
}

// derives from session object
type MongoDB struct {
	*mgo.Session           // this reference can help us run queries
	Config       IDBConfig // aggregates the db configuration
	// dialInfo     *mgo.DialInfo // for connecting to the databse
}

// DialConn		: will use the DB's configuration to dial the connection
// will also assign the session and the collections
// coll 		: name of the default collection
// archv 		: name of the archival collection collection
func (mgdb *MongoDB) DialConn(coll, archv string) error {
	if mgdb.Config != nil {
		cfg, _ := mgdb.Config.(*MongoConfig)
		s, err := mgo.DialWithInfo(cfg.DialInfo)
		if err != nil {
			return err
		}
		// Reads may not be entirely up-to-date, but they will always see the
		// history of changes moving forward, the data read will be consistent
		// across sequential queries in the same session, and modifications made
		// within the session will be observed in following queries (read-your-writes).
		// http://godoc.org/labix.org/v2/mgo#Session.SetMode
		s.SetMode(mgo.Monotonic, true)
		mgdb.Session = s
		// and also assign the collections
		mgdb.Config.SetDefaultColl(mgdb.Session.DB("").C(coll)).SetArchiveColl(mgdb.Session.DB("").C(archv))
		return nil
	}
	return fmt.Errorf("empty dialinfo, check and dial again")

}
func (mgdb *MongoDB) CloseConn() {
	mgdb.Close()
}
func (mgdb *MongoDB) InitConn(host, db, uname, pass string) IDBConn {
	// Tries to instantiate the config object with dial info
	// will not instantiate the session and the collection pointers
	// That happens only when we DialConn()
	mgdb.Config = &MongoConfig{DialInfo: &mgo.DialInfo{
		Addrs:    []string{host},
		Timeout:  3 * time.Second,
		Database: db,
		Username: uname,
		Password: pass,
	}, DefaultColl: nil, ArchiveColl: nil}
	mgdb.Session = nil // not assigned until dialled
	return mgdb
}

// GetOneFromColl : irrespective of the collection this one calls to get on specific from the collection
// Provide the name of the collection and the flt as callback
// errors if the query fails.
// need not instantiate result before calling
// round trip to json byt then to object you seek
func (mgdb *MongoDB) GetOneFromColl(coll string, flt func() bson.M, result *map[string]interface{}) error {
	qr := map[string]interface{}{} // result map onto which the object is imprinted
	err := mgdb.DB("").C(coll).Find(flt()).One(&qr)
	if err != nil {
		return fmt.Errorf("GetOneFromColl: %s", err)
	}
	*result = qr
	return nil
}

// AddToColl : this will add one data to the collection
// DB name is not required since empty string will lead to calling the database thats used for dialling within the dialinfo
// error only when the query fails
func (mgdb *MongoDB) AddToColl(obj interface{}, coll string) (int, error) {
	if err := mgdb.DB("").C(coll).Insert(obj); err != nil {
		return 0, fmt.Errorf("%s: failed to add one to collection", err)
	}
	return 1, nil
}
func (mgdb *MongoDB) EditOneFromColl(coll string, flt func() bson.M, result interface{}) error {
	return nil
}
func (mgdb *MongoDB) DeleteOneFromColl(coll, id string, flt func(id string) bson.M, result *interface{}) error {
	if err := mgdb.DB("").C(coll).Remove(flt(id)); err != nil {
		return fmt.Errorf("%s: failed to remove from collection", err)
	}
	*result = map[string]interface{}{"ok": 1}
	return nil
}

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
func (mgdb *MongoDB) GetSampleFromColl(coll string, size int, result *interface{}) error {
	res := map[string][]bson.ObjectId{}
	mgdb.DB("").C(coll).Pipe([]bson.M{
		{"$sample": bson.M{"size": size}},
		{"$project": bson.M{"_id": 1}},
		{"$group": bson.M{"_id": "", "sample": bson.M{"$push": "$_id"}}},
		{"$project": bson.M{"_id": 0, "sample": 1}},
	}).One(&res)
	*result = res
	return nil
}

// CountFromColl : for the given filter on the collection this can get count of documents
// coll		: ithe collection on which the filter applies
// flt		: filter the documents on the collection using this
func (mgdb *MongoDB) CountFromColl(coll string, flt func() bson.M) (int, error) {
	n, err := mgdb.DB("").C(coll).Find(flt()).Count()
	if err != nil {
		return -1, err
	}
	return n, nil
}

func (mgdb *MongoDB) RemoveFromColl(coll string, id string, affected *int) error {
	// remove from the main database , and add into the archived data base
	// need to see what details are pushed onto the archived database
	// ASK: where can we get the name of the archival database ?
	// ASK: it has to be percolated into all functions form a central configuration

	if err := mgdb.DB("").C(coll).Remove(bson.M{"_id": bson.ObjectIdHex(id)}); err != nil {
		return fmt.Errorf("failed to remove account from database")
	}
	*affected = 1
	return nil
}
