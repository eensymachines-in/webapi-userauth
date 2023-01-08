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

func seedDB() error {
	if err := connectDB(); err != nil {
		return err
	}
	// file, err := os.Open("./seed.json")
	// if err != nil {
	// 	return fmt.Errorf("seedDB/os.Open %s", err)
	// }
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
	docs := make([]interface{}, len(toInsert))
	for i, ins := range toInsert {
		docs[i] = ins
	}
	err = coll.Insert(docs...)
	if err != nil {
		return fmt.Errorf("seedDB/Insert %s", err)
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
	db := useracc.InitDB("localhost:37017", "useraccs", "", "", reflect.TypeOf(&useracc.MongoDB{}))
	persist, close, err := useracc.DialConnectDB(db)
	assert.Nil(t, err, "failed to dial connection on db")
	assert.NotNil(t, persist, "nil connection on dial")
	close()
	/* ================================
	testing for host sever that does not exists
	================================*/
	t.Log(">> Now testing with false host connection ")
	// this is when we have the server not correctl spelt and hence the connection should fail
	_, _, err = useracc.DialConnectDB(useracc.InitDB("somehost:37017", "useraccs", "", "", reflect.TypeOf(&useracc.MongoDB{})))

	assert.NotNil(t, err, "unexpected nil err when connecting to invalid host")
	t.Log(">> Now testing with false host connection ")
	t.Log("=============")
}

func TestInsertDeleteUserAccount(t *testing.T) {
	t.Log("now testing for removing an account")
	db, close, err := useracc.DialConnectDB(useracc.InitDB("localhost:37017", "useraccs", "", "", reflect.TypeOf(&useracc.MongoDB{})))
	assert.Nil(t, err, "unexpected error when Connecting to DB")
	defer close()
	var result interface{}
	err = db.(useracc.IMongoQry).DeleteOneFromColl("useraccs", "", func(id string) bson.M {
		return bson.M{"_id": bson.IsObjectIdHex(id)}
	}, &result)
	t.Logf("%v", result)
	t.Log(err)
}

func TestGetSampleFromColl(t *testing.T) {
	seedDB() // seed db will open the connection to database
	defer Conn.Close()
	result := map[string]interface{}{}
	Conn.DB("").C(COLL_NAME).Pipe([]bson.M{
		{"$sample": bson.M{"size": 10}},
		{"$project": bson.M{"_id": 1}},
		{"$group": bson.M{"_id": "", "sample": bson.M{"$push": "$_id"}}},
		{"$project": bson.M{"_id": 0, "sample": 1}},
	}).One(&result)
	// TODO: test this getting sample of useraccounts from the database
	// once you have this query then finalize (mgdb *MongoDB) GetSampleFromColl
	// which then can help you testing further
	t.Log(result["sample"])
}
