package useracc

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Service functions that deal with useraccount over defined interfaces
These are often used to perform operations on the database using the useracc
Strictly speaking these are functions that would access useracc over an interface
============================================== */
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/eensymachines-in/apierr/v2"
	"github.com/eensymachines.in/useracc/nosql"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

// FindWithName : for a list of addresses this can compare the PO and pick the correct one most nearby your areaa
func FindWithName(name string, from []Address) *Address {
	// case insensitive match on the post office name
	// we are trying to find the phrase in the name of the address post office
	// if found return the address that contains the phrase
	if name != "" {
		regx := regexp.MustCompile(fmt.Sprintf("(?i)%s", name))
		for _, item := range from {
			if regx.MatchString(item.PO) {
				return &item
			}
		}
	} else {
		// incase there isnt a likely name for the po , this can pick the first one
		return &from[0]
	}
	return nil
}

// FetchPOs : this for the given postal code will get the other details of the address
// this works only for the indian postal network
// findSimilar : string of the name of the postoffice nearby
// for a given postal code there can be multiple zones that post offices cover,
// api when queried gets back with multiple post offices, to choose from among them use findSimilar
// request timeout is a huge problem with the postalpincode server , we need to see if we can build a local database of addresses
func FetchPOs(pncde string, result *[]Address) error {
	url := fmt.Sprintf("https://api.postalpincode.in/pincode/%s", pncde)
	/* =================
	this api server seems to be sluggish
	often have I see this time out
	either for now we need to bump up the timeout on the client
	or for this we need to build our own database
	=============*/
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
	case http.StatusRequestTimeout:
		return fmt.Errorf("FetchPOs: postal server unresponsive")
	case http.StatusOK:
		defer resp.Body.Close() // json body of the address details
		byt, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("FetchPOs: failed to read response from postal server %s", err)
		}
		// the object we receive in response
		type Payload struct {
			PostOffices []Address `json:"PostOffice"`
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

// NewAccount : this will create a new account object, get the address from the pincode and return over IValidate
// for the given pincode this can also make a new address by fetching the details from indian postal api
// multiple POs in the same pincode - this will assign the first one
// account can later be modified for the address
func NewUsrAccount(email, title, phone, pincode string) (IValidate, error) {
	var addr Address
	addrOpts := []Address{}
	if err := FetchPOs(pincode, &addrOpts); err == nil {
		// when we could get the address options
		// from the address options we select the address for the pincode
		addr = *(FindWithName("", addrOpts))
	}
	return &UserAccount{
		Eml:  email,
		Ttle: title,
		Phn:  phone,
		Addr: addr,
	}, nil
}

// DuplicateAccount : will check to see if the account with unique fields already exists
// Incase duplicate is found sends back a bool and error incase the operation on the db failed
func DuplicateAccount(ua IUsrAcc, db nosql.IQryable) (bool, error) {
	return false, nil
}

// RegisterNewAccount : registers a new account in the database and sends back the result
func RegisterNewAccount(ua IUsrAcc, db nosql.IQryable, result *map[string]interface{}) error {
	if ua == nil {
		return apierr.Throw(fmt.Errorf("account to register cannot be nil")).Code(apierr.ErrorCode(apierr.InvldParamErr)).Context("RegisterNewAccount").Message("Invalid user account")
	}
	if !ua.(IValidate).IsValid() {
		return apierr.Throw(fmt.Errorf("one or more fields for the user account is invalid")).Code(apierr.ErrorCode(apierr.InvldParamErr)).Context("RegisterNewAccount").Message("Invalid user account")
	}
	yes, err := DuplicateAccount(ua, db)
	if err != nil {
		return err
	}
	if yes {
		return apierr.Throw(fmt.Errorf("cannot register duplicate account")).Code(apierr.ErrorCode(apierr.InvldParamErr)).Context("RegisterNewAccount").Message("Duplicate user account").LogInfo(log.WithFields(log.Fields{
			"title": ua.Title(),
			"email": ua.Contact()["email"],
			"phone": ua.Contact()["phone"],
		}))
	}
	// and then finally we are ready to add a new account
	count, err := db.AddToColl(ua, "usraccs")
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"added_count": count,
		"email":       ua.Contact()["email"],
	}).Info("New account registered")
	// All then that remains is sending back the id of the newly created account
	// for that we need to get the account details that we just inserted
	err = db.GetOneFromColl("usraccs", func() bson.M {
		return bson.M{"email": ua.Contact()["email"]}
	}, result)
	if result == nil {
		// added to the db but not found, this is weird should never happen
		return apierr.Throw(fmt.Errorf("one or more fields for the user account is invalid")).Code(apierr.ErrorCode(apierr.InvldParamErr)).Context("RegisterNewAccount").Message("Invalid user account")
	}
	if err != nil {
		return apierr.Throw(fmt.Errorf("one or more fields for the user account is invalid")).Code(apierr.ErrorCode(apierr.InvldParamErr)).Context("RegisterNewAccount").Message("Invalid user account")
	}
	return nil
}
