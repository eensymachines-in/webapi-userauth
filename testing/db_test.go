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
	"time"

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
	DATABASE_NAME = "useraccs"
	COLL_NAME     = "users"
)

func connectDB() error {
	var err error
	Conn, err = mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{"localhost:47017"},
		Timeout:  3 * time.Second,
		Database: DATABASE_NAME, //default fallback database, when collection is not specified this will be considered
		Username: "",
		Password: "",
	})
	if err != nil {
		return fmt.Errorf("connectDB/DialWithInfo: %s", err)
	}
	if err := Conn.Ping(); err != nil {
		return fmt.Errorf("connectDB/Ping: %s", err)
	}
	return nil
}

// flushDB : this will help to flush the database of the previous seed
// Caution do not call this function in production environment
func flushDB() {
	if Conn != nil {
		Conn.DB(DATABASE_NAME).C(COLL_NAME).RemoveAll(bson.M{})
	}
}
func seedDB() error {
	byt, err := os.ReadFile("./seed.json")
	if err != nil {
		return fmt.Errorf("seedDB/ReadAll %s", err)
	}
	// use the session to seed database
	toInsert := []useracc.UserAccount{}
	if err := json.Unmarshal(byt, &toInsert); err != nil {
		return fmt.Errorf("seedDB/json.Unmarshal %s", err)
	}
	coll := Conn.DB(DATABASE_NAME).C(COLL_NAME)
	// NOTE: The address needs to be expanded before it can pushed to the DB
	// NewUsrAccount : will expand the address
	for _, item := range toInsert {
		ua, err := useracc.NewUsrAccount(item.Eml, item.Ttle, item.Phn, item.Addr.Pincode)
		if err == nil { // no error creating new account
			err = coll.Insert(ua)
			if err != nil { // error inserting item into collection
				continue
			}
			// IMP: incase error inserting user account will skip
		}
		// IMP: incase of error getting the address, will skip inserting the account
	}
	return nil
}

// TestDBConnection : testing the general DB connection
// initializing and dialling the database connection
func TestDBConnection(t *testing.T) {
	/* ================================
	testing for happy path - the server and database does exists
	if the database does not exists it would be implicitly created
	================================*/
	t.Log("=========== testing DB connection and init ========")
	db := nosql.InitDB("localhost:37017", "useraccs", "", "", reflect.TypeOf(&nosql.MongoDB{}))
	persist, close, err := nosql.DialConnectDB(db)
	assert.Nil(t, err, "failed to dial connection on db")
	assert.NotNil(t, persist, "nil connection on dial")
	close()
	/* ================================
	testing for host sever that does not exists
	================================*/
	t.Log(">> Now testing with false host connection ")
	// this is when we have the server not correctl spelt and hence the connection should fail
	_, _, err = nosql.DialConnectDB(nosql.InitDB("somehost:37017", "useraccs", "", "", reflect.TypeOf(&nosql.MongoDB{})))

	assert.NotNil(t, err, "unexpected nil err when connecting to invalid host")
	t.Log(">> Now testing with false host connection ")
	t.Log("=============")
}

func TestInsertDeleteUserAccount(t *testing.T) {
	t.Log("now testing for removing an account")
	db, close, err := nosql.DialConnectDB(nosql.InitDB("localhost:37017", "useraccs", "", "", reflect.TypeOf(&nosql.MongoDB{})))
	assert.Nil(t, err, "unexpected error when Connecting to DB")
	defer close()
	var result interface{}
	err = db.(nosql.IQryable).DeleteOneFromColl("useraccs", "", func(id string) bson.M {
		return bson.M{"_id": bson.IsObjectIdHex(id)}
	}, &result)
	t.Logf("%v", result)
	t.Log(err)
}

func TestGetSampleFromColl(t *testing.T) {
	// ========
	// setting up the database in the database
	// ========
	connectDB()
	flushDB()
	seedDB() // seed db will open the connection to database
	defer Conn.Close()
	// ==============
	// dial connecting the database
	// ==============
	db, close, err := nosql.DialConnectDB(nosql.InitDB("localhost:47017", DATABASE_NAME, "", "", reflect.TypeOf(&nosql.MongoDB{})))
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
	// ========
	// setting up the database in the database
	// ========
	connectDB()
	flushDB()
	seedDB() // seed db will open the connection to database
	defer Conn.Close()
	// ==============
	// dial connecting the database
	// ==============
	db, close, err := nosql.DialConnectDB(nosql.InitDB("localhost:47017", DATABASE_NAME, "", "", reflect.TypeOf(&nosql.MongoDB{})))
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
