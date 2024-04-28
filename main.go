package main

/* Microservice for user registration, authorization, authentication.
This also has auxilliary function that lead to deletion and updation of user
uses http rest over json to define endpoints that can be called to modiufy users
author		:kneerunjun@gmail.com
*/
import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/eensymachines-in/utilities"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// AppEnviron : Object defined for containing all the environment variables.
type AppEnviron struct {
	MongoSrvr string `json:"MONGO_SRVR"`
	MongoUsr  string `json:"MONGO_USER"`
	MongoPass string `json:"MONGO_PASS"`
}

var (
	environ = AppEnviron{} // instance of the app environment, gets  populated in the init functio
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
		PadLevelText:  true,
	})
	log.SetReportCaller(false)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel) // default is info level, if verbose then trace

	/* -------------------  Reading in the environment */
	// and then give it a spin to test
	tempEnviron := map[string]string{}
	for _, v := range []string{"MONGO_SRVR", "MONGO_USER", "MONGO_PASS"} {
		// checking for all the environment variables, incase missing will panic
		if os.Getenv(v) == "" {
			log.Fatalf("One or more of the environment variables is missing %s", v)
		} else {
			tempEnviron[v] = os.Getenv(v)
		}
	}
	byt, _ := json.Marshal(tempEnviron)
	if err := json.Unmarshal(byt, &environ); err != nil {
		log.Fatalf("failed to read in the environment variables %s", err)
	}
	log.Info("All environment vars as expected...")

	/* ----------------- Ping test for the database or go burst */
	if err := utilities.MongoPingTest(environ.MongoSrvr, environ.MongoUsr, environ.MongoPass); err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Info("Starting the userauth service")
	defer log.Warn("Closing the userauth service")
	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	api := r.Group("/api").Use(utilities.CORS)
	api.GET("/ping", func(ctx *gin.Context) {
		ctx.AbortWithStatusJSON(http.StatusOK, gin.H{
			"data": "If you can see this the webapi-userauth service is running",
		})
	})
	/* Login authentication for user, sends back a jwt token  */
	// ?action=login
	// ?action=create
	users := api.Use(utilities.MongoConnect(environ.MongoSrvr, environ.MongoUsr, environ.MongoPass, "users"))
	users.POST("/users", HndlLstUsers)
	users.GET("/users", HndlLstUsers)
	users.GET("/users/:id", HndlAUser)
	log.Fatal(r.Run(":8080"))
}
