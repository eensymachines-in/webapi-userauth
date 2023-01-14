package useracc

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Middleware queries for accounts in eensymachines
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

const (
	patternEmail = `^[[:alnum:]]+[.\-_]{0,1}[[:alnum:]]*[@]{1}[[:alpha:]]+[.]{1}[[:alnum:]]{2,}[.]{0,1}[[:alnum:]]{0,}$`
	patternPhone = `^[0-9]{10}$`
	patternTitle = `^[a-zA-Z0-9_\-.\s]{1,16}$`
)

type IUsrAcc interface {
	IdAsStr() string
	Title() string
	Contact() map[string]interface{}
	Address() string
	SetNewID() IUsrAcc
}

// http://www.postalpincode.in/Api-Details
// this can store the postal address for the pincode given location
type Address struct {
	PO       string `json:"Name"`
	State    string `json:"State"`
	District string `json:"District"`
	Division string `json:"Division"`
	Block    string `json:"Block"`
	Country  string `json:"Country"`
	Pincode  string `json:"Pincode"`
}

// Dialate : this for the given postal code will get the other details of the address
// this works only for the indian postal network
func (addr *Address) Dialate() error {
	url := fmt.Sprintf("https://api.postalpincode.in/pincode/%s", addr.Pincode)
	client := http.Client{
		Timeout: 4 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Dialate Address: %s", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Dialate Address: failed request %s", err)
	}
	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("Dialate Address: Pincode not found %s", addr.Pincode)
	case http.StatusInternalServerError:
		return fmt.Errorf("Dialate Address: temporary server downtime")
	case http.StatusOK:
		defer resp.Body.Close() // json body of the address details
		byt, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Dialate Address: failed to read response from postal server %s", err)
		}
		if err := json.Unmarshal(byt, addr); err != nil {
			return fmt.Errorf("Dialate Address: failed to unmarshal result from postal server %s", err)
		}
	default:
		return fmt.Errorf("Dialate Address: unknown error on postal pincode server %d", resp.StatusCode)
	}
	return nil
}

type UserAccount struct {
	// Id   bson.ObjectId `bson:"_id" json:",omitempty"`
	// https://stackoverflow.com/questions/20215510/cannot-retrieve-id-value-using-mgo-with-golang
	Id bson.ObjectId `bson:"_id,omitempty"`
	// When unmarshalling to form json byt AccId is used carry the object Id
	AccId string `json:"_id"`
	Ttle  string `json:"title" bson:"title"`
	Eml   string `json:"email" bson:"email"`
	Phn   string `json:"phone" bson:"phone"`
	Addr  string `json:"address,omitempty" bson:"address"`
}

func NewUsrAccnt(tt, em, ph string) IUsrAcc {
	return &UserAccount{
		Ttle: tt,
		Eml:  em,
		Phn:  ph,
	}
}
func (ua *UserAccount) IdAsStr() string {
	// return ua.Id.Hex()
	return ""
}
func (ua *UserAccount) Title() string {
	return ua.Ttle
}
func (ua *UserAccount) Contact() map[string]interface{} {
	return map[string]interface{}{
		"email": ua.Eml,
		"phone": ua.Phn,
	}
}
func (ua *UserAccount) Address() string {
	return ua.Addr
}

func (ua *UserAccount) SetNewID() IUsrAcc {
	// ua.Id = bson.NewObjectId()
	return ua
}

// MarshalJSON : override from the json package
// since the useracc has the mongo id, that needs conversion to hex before it can be fitted into json
func (ua *UserAccount) MarshalJSON() ([]byte, error) {
	// https://stackoverflow.com/questions/55554434/golang-runtime-goroutine-stack-exceeds-1000000000-byte-limit
	// basically means you cannot call json.Marshal from within UserAccount.MarshalJSON
	// that would amount to recursion and thus stack overflow
	jRes := map[string]interface{}{
		// "id":    ua.Id.Hex(),
		"title": ua.Ttle,
		"phone": ua.Contact()["phone"],
		"email": ua.Contact()["email"],
	}
	return json.Marshal(jRes)
}

// ValidUserAccount : checks the user account for title, email and phone
func ValidUserAccount(ua IUsrAcc) bool {
	if matched, _ := regexp.Match(patternTitle, []byte(ua.Title())); !matched {
		return false
	}
	contact := ua.Contact()
	email := contact["email"]
	phone := contact["phone"]
	if matched, _ := regexp.Match(patternEmail, []byte(email.(string))); !matched {
		return false
	}
	if matched, _ := regexp.Match(patternPhone, []byte(phone.(string))); !matched {
		return false
	}
	return true
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
	if !ValidUserAccount(ua) {
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

	var newAccMap map[string]interface{}
	err = db.GetOneFromColl("usraccs", func() bson.M {
		return bson.M{"email": ua.Contact()["email"]}
	}, &newAccMap)
	if newAccMap == nil {
		// added to the db but not found, this is weird should never happen
		return apierr.Throw(fmt.Errorf("one or more fields for the user account is invalid")).Code(apierr.ErrorCode(apierr.InvldParamErr)).Context("RegisterNewAccount").Message("Invalid user account")
	}
	if err != nil {
		return apierr.Throw(fmt.Errorf("one or more fields for the user account is invalid")).Code(apierr.ErrorCode(apierr.InvldParamErr)).Context("RegisterNewAccount").Message("Invalid user account")
	}
	// roundtripping from json to get the user account from it
	newAcc := UserAccount{}
	byt, _ := json.Marshal(newAccMap)
	json.Unmarshal(byt, &newAcc)

	return nil
}
