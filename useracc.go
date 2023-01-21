package useracc

/* ==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
definition of basic model and its properties for the User Account
Also defines the methods that are directly available on it
============================================== */
import (
	"fmt"
	"regexp"

	"gopkg.in/mgo.v2/bson"
)

type UserAccount struct {
	// Id   bson.ObjectId `bson:"_id" json:",omitempty"`
	// https://stackoverflow.com/questions/20215510/cannot-retrieve-id-value-using-mgo-with-golang
	Id bson.ObjectId `bson:"_id,omitempty"`
	// When unmarshalling to form json byt AccId is used carry the object Id
	AccId string  `json:"id"`
	Ttle  string  `json:"title" bson:"title"`
	Eml   string  `json:"email" bson:"email"`
	Phn   string  `json:"phone" bson:"phone"`
	Addr  Address `json:"address,omitempty" bson:"address"`
}

func (ua *UserAccount) IsValid() bool {
	if ok, _ := regexp.MatchString(patternTitle, ua.Ttle); !ok {
		return false
	}
	if ok, _ := regexp.MatchString(patternEmail, ua.Eml); !ok {
		return false
	}
	if ok, _ := regexp.MatchString(patternPhone, ua.Phn); !ok {
		return false
	}
	return true
}
func (ua *UserAccount) IdAsStr() string {
	return ua.Id.Hex()
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
func (ua *UserAccount) Address() Address {
	return ua.Addr
}

func (ua *UserAccount) SetNewID() IUsrAcc {
	// ua.Id = bson.NewObjectId()
	return ua
}

// Strigify :typically used to convert the object to string in debugging purposes
func (ua *UserAccount) Stringify() string {
	return fmt.Sprintf("%s-%s-%s@%s", ua.Ttle, ua.Eml, ua.Phn, IString(&ua.Addr).Stringify())
}
