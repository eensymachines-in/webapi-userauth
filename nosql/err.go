package nosql

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

/*
==============================================
Copyright (c) Eensymachines
Developed by 		: kneerunjun@gmail.com
Developed on 		: FEB'23
Error definition, custom error for no sql package
==============================================
*/
type ErrCodeDBNoSQL uint8 // custom error for the nosql package

const (
	ErrEmptyResult = iota + uint8(0) // no result from the query
	ErrQryFail                       // query fails
	ErrNoConn                        // connection to the database failed
	ErrInvldFlt                      // query filter is nil or invalid
	ErrInvldColl                     // invalid colleciton name
	ErrEmptyInsert                   // attempt to insert an empty item in the datbaase
)

// ErrNoSQL : custom error for no sql db
type ErrNoSQL struct {
	Code      ErrCodeDBNoSQL // from one of the above error codes
	Internal  error          // errors from lower stacks
	Context   string         // location from where the error has originated
	Diagnosis string         // remedy for the error
	UsrMsg    string         // message of the error shown on the client
	LogEntry  *log.Entry     // this helps in debugging the error on the server
}

func ThrowErrNoSQL(code uint8) *ErrNoSQL {
	return &ErrNoSQL{
		Code: ErrCodeDBNoSQL(code),
	}
}

// chained set properties
func (ens *ErrNoSQL) SetInternalErr(e error) *ErrNoSQL {
	ens.Internal = e
	return ens
}
func (ens *ErrNoSQL) SetContext(ctx string) *ErrNoSQL {
	ens.Context = ctx
	return ens
}
func (ens *ErrNoSQL) SetDiagnosis(diag string) *ErrNoSQL {
	ens.Diagnosis = diag
	return ens
}
func (ens *ErrNoSQL) SetUsrMsg(msg string) *ErrNoSQL {
	ens.UsrMsg = msg
	return ens
}

func (ens *ErrNoSQL) SetLogEntry(le *log.Entry) *ErrNoSQL {
	ens.LogEntry = le
	return ens
}

// Error : so that we have conformation on the error interface
// use this to print error messages fit for user consumption
func (ens *ErrNoSQL) Error() string {
	return fmt.Sprintf("%s: %s", ens.UsrMsg, ens.Diagnosis)
}

// Logs : logs the error ith code and internals
// does not set the log preferences
func (ens *ErrNoSQL) Log() {
	ens.LogEntry.Errorf("%d: %s: %s", ens.Code, ens.Internal, ens.Context)
}
