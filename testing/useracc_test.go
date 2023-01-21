package testing

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Simple test for testing the useraccount datamodel
============================================== */

import (
	"encoding/json"
	"testing"

	"github.com/eensymachines.in/useracc"
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
