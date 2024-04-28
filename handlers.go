package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eensymachines-in/errx/httperr"
	"github.com/eensymachines-in/webapi-userauth/models"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

func HndlAUser(c *gin.Context) {
	// --------- mongo connections
	val, _ := c.Get("mongo-client")
	mongoClient := val.(*mongo.Client)
	val, _ = c.Get("mongo-database")
	db := val.(*mongo.Database)
	uc := models.UsersCollection{DbColl: db.Collection("users")}
	defer mongoClient.Disconnect(context.Background())

	usrId := c.Param("id")
	if usrId == "" {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrContxParamMissing(fmt.Errorf("missing user id for which request")), log.WithFields(log.Fields{
			"stack": "HndlAUser",
		}))
		return
	}
	usr := models.User{} // onto which we take the payload on
	if err := c.ShouldBind(&usr); err != nil {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrContxParamMissing(fmt.Errorf("missing user id for which request")), log.WithFields(log.Fields{
			"stack": "HndlAUser",
		}))
		return
	}
	if err := uc.FindUser(usrId, &usr); err != nil {
		httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
			"stack": "HndlAUser",
		}))
		return
	}
	if c.Request.Method == "GET" {
		// trying to get the single user i
		c.AbortWithStatusJSON(http.StatusOK, usr)
	}
}

// HndlLstUsers : handles list of users, can post a new user
// Can login when POST, action=auth
// Can authorize when GET action=auth
// for all other purposes it will ne method not allowed
func HndlLstUsers(c *gin.Context) {
	// --------- request binding

	// --------- mongo connections
	val, _ := c.Get("mongo-client")
	mongoClient := val.(*mongo.Client)
	val, _ = c.Get("mongo-database")
	db := val.(*mongo.Database)
	uc := models.UsersCollection{DbColl: db.Collection("users")}
	defer mongoClient.Disconnect(context.Background())

	action := c.Query("action")

	if c.Request.Method == "POST" {
		usr := models.User{}
		err := httperr.ErrBinding(c.ShouldBind(&usr))
		if err != nil {
			httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
				"stack": "HndlUserAuth",
			}))
			return
		}
		if action == "auth" {
			err := uc.Authenticate(&usr)
			if err != nil {
				httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
					"stack": "HndlUserAuth",
				}))
				return
			}
			// time to send back the token
			c.AbortWithStatusJSON(http.StatusOK, usr)
		} else if action == "create" {
			usr.Role = models.EndUser // when creating new user the role will always be EndUser
			err = uc.NewUser(&usr)
			if err != nil {
				httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
					"stack": "HndlUsers",
				}))
				return
			}
			c.AbortWithStatusJSON(http.StatusOK, &usr)
		} else {
			c.AbortWithStatus(http.StatusMethodNotAllowed) // on all other cases method is not allowed.
		}
	} else if c.Request.Method == "GET" {
		if action == "auth" {
			tok := c.Request.Header.Get("Authorization")
			if tok == "" {
				httperr.HttpErrOrOkDispatch(c, httperr.ErrForbidden(fmt.Errorf("empty token cannot request authorization")), log.WithFields(log.Fields{
					"stack": "HndlUserAuth",
				}))
				return
			} else {
				err := uc.Authorize(tok) // user fields would be empty per say since its only the token you are authorizing
				if err != nil {
					httperr.HttpErrOrOkDispatch(c, err, log.WithFields(log.Fields{
						"stack": "HndlUserAuth",
					}))
					return
				}
				c.AbortWithStatus(http.StatusOK)
			}
		} else {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
		}
	} else {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
	}
}
