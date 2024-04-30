package models

import (
	"encoding/json"
	"regexp"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type UserRole uint8

const (
	SuperUser UserRole = iota
	Admin
	EndUser
	Guest
)

type UserPassword string

func (up UserPassword) IsValid() bool {
	passRegex := regexp.MustCompile(`^[0-9a-zA-Z_!@#$%^&*-]{9,12}$`) // password is 9-12 characters
	return passRegex.MatchString(string(up))
}

func (up UserPassword) StringHash() (string, error) {
	h, e := bcrypt.GenerateFromPassword([]byte(string(up)), 14)
	if e != nil {
		return "", e
	}
	return string(h), nil
}

type UserName string

func (un UserName) IsValid() bool {
	passRegex := regexp.MustCompile(`^[a-zA-Z\s]+$`) 
	return passRegex.MatchString(string(un))
}

type UserEmail string

func (ue UserEmail) IsValid() bool {
	passRegex := regexp.MustCompile(`^[a-zA-Z0-9]+[_.-]*[a-zA-Z0-9]*@[a-zA-Z0-9]+[.]{1}[a-zA-Z0-9]{2,}$`) // password is 9-12 characters
	return passRegex.MatchString(string(ue))
}

// User : any user in the system, can be authenticated against database
type User struct {
	Id      primitive.ObjectID `bson:"_id,omitempty"` // omit empty to indicate empty when marshalling and inserting
	Name    UserName           `bson:"name" `
	Email   UserEmail          `bson:"email" `
	Role    UserRole           `bson:"role"`
	TelegID int64              `bson:"telegid"`
	Auth    string             `bson:"auth"`
	AuthTok string             `bson:"-"` // has no significance in bson
}

// MarshalJSON : Since we want to trim out certain fields before json is sent back over http
// Since its implemented on USer and not *USer it suffices for both *User and USer
func (u User) MarshalJSON() ([]byte, error) {
	profile := struct {
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		Email   string   `json:"email"`
		Role    UserRole `json:"role"`
		TelegID int64    `json:"telegid"`
		AuthTok string   `json:"authtok"`
	}{
		ID:      u.Id.Hex(),
		Name:    string(u.Name),
		Email:   string(u.Email),
		Role:    u.Role,
		TelegID: u.TelegID,
		AuthTok: u.AuthTok,
	}
	return json.Marshal(&profile)
}

type CustomClaims struct {
	jwt.StandardClaims
	User     string   `json:"user"`
	UserRole UserRole `json:"user-role"`
}
