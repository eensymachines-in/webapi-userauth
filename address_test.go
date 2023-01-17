package useracc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAddrDialate	: every address object starts with a pin code and then gets its details from 3rd part api
// this test is to testing the same
func TestAddrDialate(t *testing.T) {
	pincode := "411038"
	addrOptions := Addresses{}
	err := FetchPOs(pincode, &addrOptions)
	assert.Nil(t, err, "Unexpected error when getting the POs from the pincode")
	addr := addrOptions.FindNearbyLikely("Kothrud")
}
