package useracc

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Defines in detail the address object for the useracc
Every useracc has an address.
============================================== */
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// http://www.postalpincode.in/Api-Details
// this can store the postal address for the pincode given location
// gets aggregated within the useracc
// PinCode is vital in getting the address. - fields are hydrated from 3rdparty api
type Address struct {
	PO       string `json:"Name"`
	State    string `json:"State"`
	District string `json:"District"`
	Division string `json:"Division"`
	Block    string `json:"Block"`
	Country  string `json:"Country"`
	Pincode  string `json:"Pincode"`
}
type Addresses []Address

// FindNearbyLikely : for a list of addresses this can compare the PO and pick the correct one most nearby your areaa
func (addrs Addresses) FindNearbyLikely(name string) *Address {
	for _, item := range addrs {
		if strings.Contains(item.PO, name) {
			return &item
		}
	}
	return nil
}

// FetchPOs : this for the given postal code will get the other details of the address
// this works only for the indian postal network
// findSimilar : string of the name of the postoffice nearby
// for a given postal code there can be multiple zones that post offices cover,
// api when queried gets back with multiple post offices, to choose from among them use findSimilar
func FetchPOs(pncde string, result *Addresses) error {
	url := fmt.Sprintf("https://api.postalpincode.in/pincode/%s", pncde)
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("FetchPOs: %s", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("FetchPOs: failed request %s", err)
	}
	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("FetchPOs: Pincode not found %s", pncde)
	case http.StatusInternalServerError:
		return fmt.Errorf("FetchPOs: temporary server downtime")
	case http.StatusOK:
		defer resp.Body.Close() // json body of the address details
		byt, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("FetchPOs: failed to read response from postal server %s", err)
		}
		// the object we receive in response
		type Payload struct {
			PostOffices []Address `PostOffice`
		}
		respPayload := []Payload{} // somehow the response is a slice, at the top
		if err := json.Unmarshal(byt, &respPayload); err != nil {
			return fmt.Errorf("FetchPOs: failed to unmarshal result from postal server %s", err)
		}
		// TODO: check for the length of the response on the top
		*result = respPayload[0].PostOffices
	default:
		return fmt.Errorf("FetchPOs: unknown error on postal pincode server %d", resp.StatusCode)
	}
	return nil
}
