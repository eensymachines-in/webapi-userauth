package useracc

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFetchPOs	: every address object starts with a pin code and then gets its details from 3rd part api
// this test is to testing the same
func TestFetchPOs(t *testing.T) {
	pincode := "411038"
	addrOptions := []Address{}
	err := FetchPOs(pincode, &addrOptions)
	assert.Nil(t, err, "Unexpected error when getting the POs from the pincode")
	addr := FindWithName("bhusari", addrOptions)
	assert.NotNil(t, addr, "Unexpected nil address")
	t.Log(addr)
}

func TestAddressNameFind(t *testing.T) {
	// for finding the address with a relevant name phrase we test the regex functionality go here
	regx := regexp.MustCompile(fmt.Sprintf("(?i)%s", "Hello"))
	// then have this match the string incoming
	t.Log(regx.MatchString("_Hello_There"))

}
