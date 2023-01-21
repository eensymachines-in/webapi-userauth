package useracc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAccRegistration(t *testing.T) {
	newReg, err := NewUsrAccount("johndoe@somedomain.com", "John Doe", "8390306860", "411038")
	assert.Nil(t, err, "Unexpected error when creating a new account, check validation")
	assert.NotNil(t, newReg, "Unpexpected nil account")
	assert.True(t, newReg.IsValid(), "Unexpected false on validation")
}

// TestJsonUserAcc : this is to get the user account tested for json marshalling and unmarshalling
func TestJsonUserAcc(t *testing.T) {
	pncde := "411038"
	// this one goes right in
	newReg, err := NewUsrAccount("johndoe@somedomain.com", "John Doe", "8390306860", pncde)
	if newReg == nil || err != nil {
		t.Error(err)
		return
	}
	// if the object can make a round trip in json we know all the json tags are working
	byt, err := json.Marshal(newReg)
	assert.Nil(t, err, "Unexpected error when marshalling user object")
	var result UserAccount
	json.Unmarshal(byt, &result)
	t.Log(IString(&result).Stringify())
}