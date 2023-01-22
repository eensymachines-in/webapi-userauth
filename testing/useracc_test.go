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
	"testing"

	"github.com/eensymachines.in/useracc"
	"github.com/eensymachines.in/useracc/nosql"
	"github.com/stretchr/testify/assert"
)

func TestUserAccRegistration(t *testing.T) {
	newReg, err := useracc.NewUsrAccount("johndoe@somedomain.com", "John Doe", "8390306860", "411038")
	assert.Nil(t, err, "Unexpected error when creating a new account, check validation")
	assert.NotNil(t, newReg, "Unpexpected nil account")
	assert.True(t, newReg.IsValid(), "Unexpected false on validation")
}

// TestJsonUserAcc : this is to get the user account tested for json marshalling and unmarshalling
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

func TestNewRegisterAcc(t *testing.T) {
	// ========
	// setting up the database in the database
	// ========
	connectDB()
	flushDB()
	seedDB() // seed db will open the connection to database
	defer Conn.Close()
	db, close, err := nosql.DialConnectDB(nosql.InitDB("localhost:47017", DATABASE_NAME, "", "", reflect.TypeOf(&nosql.MongoDB{})))
	defer close()
	assert.Nil(t, err, "failed to connect to db")
	assert.NotNil(t, db, "nil db pointer")

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

}
