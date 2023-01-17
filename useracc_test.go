package useracc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressDialation(t *testing.T) {
	addr := &Address{Pincode: "411038"}
	assert.Nil(t, addr.Dialate(), "unexpected error when dialating address object")
	t.Log(addr)
}
