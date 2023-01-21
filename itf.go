package useracc

/*
==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: JAN'23
Definition of interface on the packages for the UserAccount
==============================================
*/
const (
	patternEmail = `^[[:alnum:]]+[.\-_]{0,1}[[:alnum:]]*[@]{1}[[:alpha:]]+[.]{1}[[:alnum:]]{2,}[.]{0,1}[[:alnum:]]{0,}$`
	patternPhone = `^[0-9]{10}$`
	patternTitle = `^[a-zA-Z0-9_\-.\s]{1,16}$`
)

type IUsrAcc interface {
	IdAsStr() string
	Title() string
	Contact() map[string]interface{}
	Address() Address
	SetNewID() IUsrAcc
}

// IValidate : this helps any account to be validated
type IValidate interface {
	IsValid() bool
}

type IString interface {
	Stringify() string
}
