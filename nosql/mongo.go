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
	*mgo.Session               // this reference can help us run queries
	dialInfo     *mgo.DialInfo // for connecting to the databse

}

func (mgdb *MongoDB) DialConn() error {
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
		mgdb.Session = s
		return nil
	}
	return fmt.Errorf("empty dialinfo, check and dial again")

}
func (mgdb *MongoDB) CloseConn() {
	mgdb.Close()
}
func (mgdb *MongoDB) InitConn(host, db, uname, pass string) IDBConn {
	mgdb.dialInfo = &mgo.DialInfo{
		Addrs:    []string{host},
		Timeout:  3 * time.Second,
		Database: db,
		Username: uname,
		Password: pass,
	}
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
