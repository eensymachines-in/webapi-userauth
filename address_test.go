package useracc

import (
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
	addr := FindNearbyLikely("Kothrud", addrOptions)
	assert.NotNil(t, addr, "Unexpected nil address")
	t.Log(addr)
}
