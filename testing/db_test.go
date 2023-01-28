package testing

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
This shall test the database operations
to test database we can run a sample database inside a container
In the same directory find a docker-compose file to start the mongo db
============================================== */
import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/eensymachines.in/useracc"
	"github.com/eensymachines.in/useracc/nosql"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	// for testing we use an independent session
	Conn *mgo.Session
)

const (
	DATABASE_NAME  = "useraccs"
	COLL_NAME      = "users"
	ARCHVCOLL_NAME = "archive_users"
	DBHOST         = "localhost:47017"
)

// SetupMongoConn :  dials a mongo connection to testing host and can send out to the testing functions
// forceSeed : flag will force seed the database all over again
// set this t false if you need only to get the database connection
// will flush only the default collection before seeding it again
func SetupMongoConn(forceSeed bool) (nosql.IDBConn, func(), error) {
	db, close, err := nosql.InitDialConn(&nosql.DBInitConfig{
		Host:      DBHOST,
		DB:        DATABASE_NAME,
		UName:     "",
		Passwd:    "",
		Coll:      COLL_NAME,
		ArchvColl: ARCHVCOLL_NAME,
		DBTyp:     reflect.TypeOf(&nosql.MongoDB{}),
	})
	if err != nil {
		return nil, nil, err
	}
	if forceSeed {
		byt, err := os.ReadFile("./seed.json")
		if err != nil {
			return nil, nil, fmt.Errorf("seedDB/ReadAll %s", err)
		}
		toInsert := []useracc.UserAccount{}
		if err := json.Unmarshal(byt, &toInsert); err != nil {
			return nil, nil, fmt.Errorf("seedDB/ReadAll %s", err)
		}
		// HACK: having to convert to a pointer to get the aggregated session
		// this cannot be but is ok only since its the testing environment
		coll := db.(*nosql.MongoDB).Session.DB(DATABASE_NAME).C(COLL_NAME)
		coll.RemoveAll(bson.M{}) // flushing the db
		for _, item := range toInsert {
			ua, err := useracc.NewUsrAccount(item.Eml, item.Ttle, item.Phn, item.Addr.Pincode)
			if err == nil { // no error creating new account
				err = coll.Insert(ua)
				if err != nil { // error inserting item into collection
					continue
				}
			}
		}
	}
	return db, close, nil
}

// TestRemoveDoc : aimed at testing RemoveFromColl(coll string, id string, softDel bool, affected *int) error
func TestRemoveDoc(t *testing.T) {
	t.Log("now testing for removing an account")
	db, close, err := SetupMongoConn(true)
	if err != nil { // if the mongo connection was not setup, aborting the entire test
		t.Error(err)
		return
	}
	defer close()
	// getting the sample data for test
	// getting ids of 10 documents that can be used for deletion
	var result interface{}
	err = db.(nosql.IQryable).GetSampleFromColl(COLL_NAME, 10, &result)
	if err != nil { // fails to get even the samples, abort test
		t.Error(err)
		return
	}
	ids := result.(map[string][]bson.ObjectId)
	// here we go ahead to test deletion of an item from database
	// Soft deletion
	for _, id := range ids["sample"] {
		var count int
		err := db.(nosql.IQryable).RemoveFromColl(COLL_NAME, id.Hex(), true, &count)
		assert.Nil(t, err, fmt.Sprintf("Unexpected error when deleting doc with id %s", err))
	}
}

func TestGetSampleFromColl(t *testing.T) {
	// ==============
	// dial connecting the database
	// ==============
	db, close, err := SetupMongoConn(true)
	defer close()
	assert.Nil(t, err, "failed to connect to db")
	assert.NotNil(t, db, "nil db pointer")
	// ================
	var result interface{}
	err = db.(nosql.IQryable).GetSampleFromColl(COLL_NAME, 10, &result)
	assert.Nil(t, err, "unexpected error when GetSampleFromColl")
	assert.NotNil(t, result, "nil result for GetSampleFromColl")
	byt, err := json.Marshal(result)
	assert.Nil(t, err, "unexpected error when json.Marshal")
	t.Log(string(byt))
	// ================
	// GetSampleFromColl with invalid size
	err = db.(nosql.IQryable).GetSampleFromColl(COLL_NAME, -10, &result)
	// with invalid size, you dont get any error
	// the sample set would be empty
	assert.Nil(t, err, "Unexpected error when getting sample with invalid size")
	ids := result.(map[string][]bson.ObjectId)
	sample := ids["sample"]
	assert.Equal(t, 0, len(sample), "Unexpected non-empty sample size")
}

func TestGetGetOneFromColl(t *testing.T) {
	// ==============
	// dial connecting the database
	// ==============
	db, close, err := SetupMongoConn(true)
	defer close()
	assert.Nil(t, err, "failed to connect to db")
	assert.NotNil(t, db, "nil db pointer")
	// Now getting one sample from database so as to test
	var result interface{}
	db.(nosql.IQryable).GetSampleFromColl(COLL_NAME, 1, &result)
	ids := result.(map[string][]bson.ObjectId)
	sample := ids["sample"]
	assert.Equal(t, 1, len(sample), "Unexpected number of items in the sample")
	// ===============
	ua := useracc.UserAccount{}
	var uaMap map[string]interface{}
	err = db.(nosql.IQryable).GetOneFromColl(COLL_NAME, func() bson.M { return bson.M{"_id": sample[0]} }, &uaMap)
	t.Log(uaMap)
	byt, _ := json.Marshal(uaMap)
	t.Log(string(byt))
	json.Unmarshal(byt, &ua)
	assert.Nil(t, err, "Unexpected error when GetOneFromColl")
	t.Log(ua)
}
