package models

/* Specific error definitions for user collection queries
we make them implement httperr.HttpErr */
import (
	"fmt"
	"net/http"

	"github.com/eensymachines-in/errx/httperr"
	log "github.com/sirupsen/logrus"
)

var (
	InvalidTokenErr = func(e error) httperr.HttpErr {
		return (&eInvalidToken{}).SetInternal(e)
	}
	MismatchPasswdErr = func(e error) httperr.HttpErr {
		return (&eMismatchPasswd{}).SetInternal(e)
	}
	AuthTokenErr = func(e error) httperr.HttpErr {
		return (&eGenTokenFail{}).SetInternal(e)
	}
)

type eInvalidToken struct {
	Internal error
}

type eMismatchPasswd struct {
	Internal error
}

type eGenTokenFail struct {
	Internal error
}

func (it *eInvalidToken) Error() string {
	return fmt.Sprintf("Failed to generate token: %s", it.Internal)
}
func (it *eInvalidToken) SetInternal(ie error) httperr.HttpErr {
	if ie == nil {
		return nil
	}
	it.Internal = ie
	return it
}
func (it *eInvalidToken) Log(le *log.Entry) httperr.HttpErr {
	le.WithFields(log.Fields{
		"internal_err": it.Internal,
	}).Error("invalid or expired token")
	return it
}
func (it *eInvalidToken) ClientErrData() string {
	return "Your authorization has expired/invalidated, kindly login again"
}
func (it *eInvalidToken) HttpStatusCode() int {
	return http.StatusForbidden
}

func (mp *eMismatchPasswd) Error() string {
	return fmt.Sprintf("Duplicate user: %s", mp.Internal)
}

func (mp *eMismatchPasswd) SetInternal(ie error) httperr.HttpErr {
	if ie == nil {
		return nil
	}
	mp.Internal = ie
	return mp
}
func (mp *eMismatchPasswd) Log(le *log.Entry) httperr.HttpErr {
	le.WithFields(log.Fields{
		"internal_err": mp.Internal,
	}).Error("incorrect password, unauthorized")
	return mp
}
func (mp *eMismatchPasswd) ClientErrData() string {
	return "Password did not match our records, authentication failed"
}
func (mp *eMismatchPasswd) HttpStatusCode() int {
	return http.StatusUnauthorized
}

func (gt *eGenTokenFail) Error() string {
	return fmt.Sprintf("Failed to generate token: %s", gt.Internal)
}
func (gt *eGenTokenFail) SetInternal(ie error) httperr.HttpErr {
	if ie == nil {
		return nil
	}
	gt.Internal = ie
	return gt
}
func (gt *eGenTokenFail) Log(le *log.Entry) httperr.HttpErr {
	le.WithFields(log.Fields{
		"internal_err": gt.Internal,
	}).Error("failed to generate jwt token")
	return gt
}
func (gt *eGenTokenFail) ClientErrData() string {
	return "One or more operations on server side has failed, contact an admin to fix this"
}
func (gt *eGenTokenFail) HttpStatusCode() int {
	return http.StatusInternalServerError
}
