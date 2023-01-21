package useracc

import "fmt"

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Defines in detail the address object for the useracc
Every useracc has an address.
============================================== */

// http://www.postalpincode.in/Api-Details
// this can store the postal address for the pincode given location
// gets aggregated within the useracc
// PinCode is vital in getting the address. - fields are hydrated from 3rdparty api
type Address struct {
	PO       string `json:"Name" bson:"po"`
	State    string `json:"State" bson:"state"`
	District string `json:"District" bson:"district"`
	Division string `json:"Division" bson:"div"`
	Block    string `json:"Block" bson:"block"`
	Country  string `json:"Country" bson:"country"`
	Pincode  string `json:"Pincode" bson:"pincode"`
}

func (addr *Address) Stringify() string {
	return fmt.Sprintf("%s-%s-%s-%s", addr.PO, addr.State, addr.District, addr.Pincode)
}
