package testing

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Simple test for testing the useraccount datamodel
============================================== */

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/eensymachines.in/useracc"
	"github.com/eensymachines.in/useracc/nosql"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

func TestEmailRegex(t *testing.T) {
	okEmails := []string{
		"smepsted1@printfriendly.com",
		"dmont4@purevolume.com",
		"cdufore0@7bing.com",
		"rboxe0@independent.co.uk",
	}
	patternEmail := `^[[:alnum:]]+[.\-_]{0,1}[[:alnum:]]*[@]{1}[[:alnum:]]+[.]{1}[[:alnum:]]{2,}[.]{0,1}[[:alnum:]]{0,}$`
	rgx, _ := regexp.Compile(patternEmail)
	for _, em := range okEmails {
		assert.True(t, rgx.MatchString(em), fmt.Sprintf("Unexpected fail to match pattern in %s", em))
	}
}

// TestJsonUserAcc : this is to get the user account tested for json marshalling and unmarshalling
// NOTE: round trip with json will determine if the json tags are all in place as expected
func TestJsonUserAcc(t *testing.T) {
	pncde := "411038"
	// this one goes right in
	newReg, err := useracc.NewUsrAccount("johndoe@somedomain.com", "John Doe", "8390306860", pncde)
	if newReg == nil || err != nil {
		t.Error(err)
		return
	}
	// if the object can make a round trip in json we know all the json tags are working
	byt, err := json.Marshal(newReg)
	assert.Nil(t, err, "Unexpected error when marshalling user object")
	var result useracc.UserAccount
	json.Unmarshal(byt, &result)
	t.Log(useracc.IString(&result).Stringify())
}

// TestNewRegisterAcc : test for account registration
// Will test registration of user accounts
// Will test for query failure though
// NOTE: Will NOT test for validation of accounts, duplicate accounts though - that would happen in separate tests
// instead this will test the function for errors in the db
func TestNewRegisterAcc(t *testing.T) {
	// ========
	// setting up the database in the database
	// ========
	connectDB()
	flushDB()
	seedDB() // seed db will open the connection to database
	defer Conn.Close()
	db, close, err := nosql.DialConnectDB(nosql.InitDB("localhost:47017", DATABASE_NAME, "", "", reflect.TypeOf(&nosql.MongoDB{})), COLL_NAME, ARCHVCOLL_NAME)
	defer close()
	assert.Nil(t, err, "failed to connect to db")
	assert.NotNil(t, db, "nil db pointer")
	// Define data in a map - typically this data will be used in constructor in creating accounts
	data := []map[string]string{
		{"email": "hslade0@shared.com", "title": "Hodge Slade", "phone": "6785263002", "pincode": "411038"},
		{"email": "ccarme1@harvard.edu", "title": "Curran Carme", "phone": "3796025636", "pincode": "411057"},
	}
	var result map[string]interface{}
	for _, d := range data {
		ac, err := useracc.NewUsrAccount(d["email"], d["title"], d["phone"], d["pincode"])
		if err != nil {
			t.Error(err)
			return
		}
		err = useracc.RegisterNewAccount(ac.(useracc.IUsrAcc), db.(nosql.IQryable), &result)
		assert.Nil(t, err, fmt.Sprintf("Unexpected error when RegisterNewAccount : %s", err))
	}
	// we then test for the account registration when the query fails
	// way to make the query fail is send in queryable nil
	for _, d := range data {
		ac, err := useracc.NewUsrAccount(d["email"], d["title"], d["phone"], d["pincode"])
		if err != nil {
			t.Error(err)
			return
		}
		// NOTE: the queryable is nil - this bound to fail after
		err = useracc.RegisterNewAccount(ac.(useracc.IUsrAcc), nil, &result)
		assert.NotNil(t, err, fmt.Sprintf("Unexpected error when RegisterNewAccount : %s", err))
	}
}

// TestDuplctAcc : negative test for trying to register an account with email that is already registered
func TestDuplctAcc(t *testing.T) {
	// ========
	// setting up the database in the database
	// ========
	connectDB()
	flushDB()
	seedDB() // seed db will open the connection to database
	defer Conn.Close()
	db, close, err := nosql.DialConnectDB(nosql.InitDB("localhost:47017", DATABASE_NAME, "", "", reflect.TypeOf(&nosql.MongoDB{})), COLL_NAME, ARCHVCOLL_NAME)
	defer close()
	assert.Nil(t, err, "failed to connect to db")
	assert.NotNil(t, db, "nil db pointer")
	data := []map[string]string{
		{"email": "cdobrowski0@pcworld.com", "title": "Chevy Dobrowski", "phone": "7202565595", "pincode": "411038"},
		{"email": "ygodlee1@unblog.fr", "title": "Yuri Godlee", "phone": "3342499299", "pincode": "411057"},
	}
	for _, d := range data {
		ac, err := useracc.NewUsrAccount(d["email"], d["title"], d["phone"], d["pincode"])
		if err != nil {
			t.Errorf("Unexpected error creating in new user account %s", err)
			return
		}
		yes, err := useracc.DuplicateAccount(ac.(useracc.IUsrAcc), db.(nosql.IQryable))
		// err = useracc.RegisterNewAccount(ac.(useracc.IUsrAcc), db.(nosql.IQryable), &result)
		assert.Nil(t, err, "Unexpected nil error when RegisterNewAccount")
		assert.True(t, yes, "Was expecting the account to be duplicate, reported not to be")
	}
	/*Now testing for accounts that arent registered
	 */
	// NOTE: +ve test and will test for err ==nil since we are trying to get the account not already registered
	// duplicate will not be found
	data = []map[string]string{
		{"email": "hhubbocks0@wp.com", "title": "Hadleigh Hubbocks", "phone": "8988567352", "pincode": "411038"},
		{"email": "ynewcomen1@nasa.gov", "title": "Yelena Newcomen", "phone": "3442881535", "pincode": "411057"},
	}
	for _, d := range data {
		ac, err := useracc.NewUsrAccount(d["email"], d["title"], d["phone"], d["pincode"])
		if err != nil {
			t.Errorf("Unexpected error creating in new user account %s", err)
			return
		}
		yes, err := useracc.DuplicateAccount(ac.(useracc.IUsrAcc), db.(nosql.IQryable))
		// err = useracc.RegisterNewAccount(ac.(useracc.IUsrAcc), db.(nosql.IQryable), &result)
		assert.Nil(t, err, "Unexpected nil error when RegisterNewAccount")
		assert.False(t, yes, "Wasnt expecting the account to be reported as duplicate")
	}
}

func TestDelAcc(t *testing.T) {
	// ========
	// setting up the database in the database
	// ========
	connectDB()
	flushDB()
	seedDB() // seed db will open the connection to database
	defer Conn.Close()
	db, close, err := nosql.DialConnectDB(nosql.InitDB("localhost:47017", DATABASE_NAME, "", "", reflect.TypeOf(&nosql.MongoDB{})), COLL_NAME, ARCHVCOLL_NAME)
	defer close()
	assert.Nil(t, err, "failed to connect to db")
	assert.NotNil(t, db, "nil db pointer")
	// Need to get random from the database
	var result interface{}
	db.(nosql.IQryable).GetSampleFromColl(COLL_NAME, 10, &result)
	assert.NotNil(t, result, "Unexpected nil result GetSampleFromColl")
	ids, ok := result.(map[string][]bson.ObjectId)
	assert.True(t, ok, "failed conversion for GetSampleFromColl result")
	assert.NotEqual(t, 0, len(ids), "Unexpected empty result from GetSampleFromColl")
	// once we have the samples, we then proceed for delete
	for _, id := range ids["sample"] {
		var count int
		err := db.(nosql.IQryable).RemoveFromColl(COLL_NAME, id.Hex(), &count)
		assert.Nil(t, err, "Unexpected error when RemoveFromColl")
		assert.Equal(t, 1, count)
	}
}
