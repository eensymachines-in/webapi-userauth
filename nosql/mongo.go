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

// derives from session object
type MongoDB struct {
	*mgo.Session                 // this reference can help us run queries
	DefaultColl  *mgo.Collection // default collection from which data is sought
	ArchiveColl  *mgo.Collection // soft delete operations will shift the data to archive coll
}

func (mgdb *MongoDB) SetDefaultColl(c *mgo.Collection) IDBConfig {
	mgdb.DefaultColl = c
	return mgdb
}
func (mgdb *MongoDB) SetArchiveColl(c *mgo.Collection) IDBConfig {
	mgdb.ArchiveColl = c
	return mgdb
}

// DialConn		: will use the DB's configuration to dial the connection
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
		return fmt.Errorf("DialConn: failed to connect to mongo instance: %s", err)
	}
	// Reads may not be entirely up-to-date, but they will always see the
	// history of changes moving forward, the data read will be consistent
	// across sequential queries in the same session, and modifications made
	// within the session will be observed in following queries (read-your-writes).
	// http://godoc.org/labix.org/v2/mgo#Session.SetMode
	mgdb.Session.SetMode(mgo.Monotonic, true)
	// and also assign the collections
	mgdb.SetDefaultColl(mgdb.Session.DB("").C(c.Coll)).SetArchiveColl(mgdb.Session.DB("").C(c.ArchvColl))
	return nil

}
func (mgdb *MongoDB) CloseConn() {
	mgdb.Close()
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
			return fmt.Errorf("RemoveFromColl: failed to get item to archive %s", err)
		}
		// Now push the resul/t to archival collection
		if err := mgdb.ArchiveColl.Insert(result); err != nil {
			return fmt.Errorf("RemoveFromColl: failed to insert in archival collection %s", err)
		}
	}

	if err := mgdb.DB("").C(coll).Remove(bson.M{"_id": bson.ObjectIdHex(id)}); err != nil {
		return fmt.Errorf("RemoveFromColl: failed to remove account from database %s", err)
	}
	*affected = 1
	return nil
}
