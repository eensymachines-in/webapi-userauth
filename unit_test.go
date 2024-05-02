package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/eensymachines-in/errx/httperr"
	"github.com/eensymachines-in/webapi-userauth/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gopkg.in/mgo.v2/bson"
)

const (
	TESTDB_NAME   = "testaquaponics"
	TESTCOLL_NAME = "users"
)

// TestCaseArgs : arguments for making the user
type TestCaseArgs struct {
	Uname  models.UserName
	Uemail models.UserEmail
	Urole  models.UserRole
	Uteleg int64
	Upass  string
}

type TestCase struct {
	Name string
	Args *TestCaseArgs
	Want httperr.HttpErr
}

func testConnectDatabase() (*models.UsersCollection, error) {
	listEnviron := os.Environ()
	var server, usr, pass string
	for _, env := range listEnviron {
		entry := strings.Split(env, "=")
		if entry[0] == "MONGO_SRVR" {
			server = entry[1]
		} else if entry[0] == "MONGO_USER" {
			usr = entry[1]
		} else if entry[0] == "MONGO_PASS" {
			pass = entry[1]
		}
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s", usr, pass, server)))
	if err != nil {
		return nil, err
	}
	/* here we get references to the databases and the collection onto which we do all the operattions  */
	db := client.Database(TESTDB_NAME)
	coll := db.Collection(TESTCOLL_NAME)
	uc := models.UsersCollection{DbColl: coll}
	/* from dummy json we will insert all the data for the teest database
	incase there is an error we report that back when before running the test */

	f, err := os.Open("./dummy.json")
	if err != nil {
		logrus.Error("failed to open dummy data file")
		return &uc, nil
	}
	byt, err := io.ReadAll(f)
	if err != nil {
		logrus.Error("failed to read dummy data file")
		return &uc, nil
	}
	dummyUsers := []models.User{}
	if err := json.Unmarshal(byt, &dummyUsers); err != nil {
		logrus.Error("failed unmrshall dummy data ")
		return &uc, nil
	}
	for _, u := range dummyUsers {
		u.Auth, err = models.UserPassword(u.Auth).StringHash()
		if err != nil {
			logrus.Error("failed to hash password")
			continue
		}
		_, err := coll.InsertOne(context.Background(), u)
		if err != nil {
			logrus.Error("failed to insert data.")
			continue
		}
	}
	return &uc, nil // using the test database
}

func TestAuthUser(t *testing.T) {
	uc, err := testConnectDatabase()
	if err != nil {
		t.Error(err)
		return
	}
	testCases := []TestCase{
		// only name later
		{Name: "Authenticate-user", Args: &TestCaseArgs{Uemail: "bsmewings1@storify.com", Upass: "oikTAF118*2No3K"}, Want: nil},
		{Name: "Authenticate-user", Args: &TestCaseArgs{Uemail: "pmosconi2@tiny.cc", Upass: "bnpOYT803XhLvBaZW"}, Want: nil},
	}
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			temp := &models.User{Email: tt.Args.Uemail, Auth: tt.Args.Upass}
			got := uc.Authenticate(temp)
			assert.Nil(t, got, "Unexpected error when authenticating user")
			t.Log(got)
			t.Log(temp.Auth) // spits out the authentication token
		})
	}
	t.Cleanup(func() {
		ctx := context.Background()
		uc.DbColl.DeleteMany(ctx, bson.M{})
		uc.DbColl.Database().Client().Disconnect(ctx)
	})
}

func TestUserEdit(t *testing.T) {
	uc, err := testConnectDatabase()
	if err != nil {
		t.Error(err)
		return
	}
	testCases := []TestCase{
		// only name later
		{Name: "Alter-Name", Args: &TestCaseArgs{Uname: "Felipe Janny", Uemail: "struce0@bloomberg.com"}, Want: nil},
		// now altering the password as well
		{Name: "Alter-Pass", Args: &TestCaseArgs{Uname: "Felipe Janny", Uemail: "struce0@bloomberg.com", Upass: "lrpKGV515"}, Want: nil},
		{Name: "Alter-Pass", Args: &TestCaseArgs{Uname: "Felipe Janny", Uemail: "struce0@bloomberg.com", Upass: "lrpKGV515", Uteleg: 7657657566}, Want: nil},
	}
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			got := uc.EditUser(string(tt.Args.Uemail), string(tt.Args.Uname), tt.Args.Upass, tt.Args.Uteleg)
			assert.Equal(t, got, tt.Want, "unexpected error response when altering user details")
		})
	}
	/* here we perform some bad test cases*/
	badTestCases := []TestCase{
		// User that wasnt found registered cannot be altered
		{Name: "Alter-NotFound", Args: &TestCaseArgs{Uname: "Felipe Janny", Uemail: "struce0@gandberg.com"}, Want: nil},
		{Name: "Alter-BadUName", Args: &TestCaseArgs{Uname: "??^%^$^Mr", Uemail: "struce0@bloomberg.com"}, Want: nil},
		// case of a bad password
		{Name: "Alter-BadUName", Args: &TestCaseArgs{Uname: "Felipe Janny", Uemail: "struce0@bloomberg.com", Upass: "54655"}, Want: nil},
	}
	for _, tt := range badTestCases {
		t.Run(tt.Name, func(t *testing.T) {
			got := uc.EditUser(string(tt.Args.Uemail), string(tt.Args.Uname), tt.Args.Upass, tt.Args.Uteleg)
			assert.NotNil(t, got, "Unexpected nil error when in bad test case")
		})
	}
	t.Cleanup(func() {
		ctx := context.Background()
		uc.DbColl.DeleteMany(ctx, bson.M{})
		uc.DbColl.Database().Client().Disconnect(ctx)
	})
}

// TestUserCRUD : will test the complete crud operation fo the users
func TestUserInsert(t *testing.T) {
	uc, err := testConnectDatabase()
	if err != nil {
		t.Error(err)
		return
	}
	// 200 ok test for creating new user
	tests := []TestCase{
		{Name: "Simple user insert-I", Want: nil, Args: &TestCaseArgs{Uname: "Belva Cutchie", Uemail: "bcutchie0@live.com", Urole: models.Guest, Uteleg: 4564765786, Upass: "xrqYOB165f8V"}},
		{Name: "Simple user insert-II", Want: nil, Args: &TestCaseArgs{Uname: "Laurie Kmietsc", Uemail: "lkmietsch0@mac.com", Urole: models.Guest, Uteleg: 4564765776, Upass: "lcxWLM5753Xo"}},
		{Name: "Simple user insert-III", Want: nil, Args: &TestCaseArgs{Uname: "Wake Aaron", Uemail: "waaron1@merriamwebster.com", Urole: models.Guest, Uteleg: 5564765786, Upass: "feuTUC462GH"}},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := uc.NewUser(&models.User{Name: tt.Args.Uname, Email: tt.Args.Uemail, Role: tt.Args.Urole, TelegID: tt.Args.Uteleg, Auth: tt.Args.Upass})
			assert.Equal(t, got, tt.Want, "unexpected response when creating new user")
		})
	}
	// Now fori bad test cases
	badtests := []TestCase{
		{Name: "Bad user insert-I", Want: nil, Args: &TestCaseArgs{Uname: "", Uemail: "bcutchie0@live.com", Urole: models.Guest, Uteleg: 4564765786, Upass: "xrqYOB165f8V"}},                                                // user name isnt valid
		{Name: "Bad user insert-II", Want: nil, Args: &TestCaseArgs{Uname: "Laurie Kmietsc", Uemail: "lkmietsch0@mac.com", Urole: models.Guest, Uteleg: 4564765776, Upass: "lcxWLM5753Xo~$@#$@#$$%#$%E^$%^%$^%$^$%^$%^$%"}}, // password aint valid
		{Name: "Bad user insert-III", Want: nil, Args: &TestCaseArgs{Uname: "Wake Aaron", Uemail: "waaron1@@@@", Urole: models.Guest, Uteleg: 5564765786, Upass: "feuTUC462GH"}},                                            // email aint valid
		{Name: "Bad user insert-IV", Want: nil, Args: &TestCaseArgs{Uname: "Belva Cutchie", Uemail: "bcutchie0@live.com", Urole: models.Guest, Uteleg: 4564765786, Upass: "xrqYOB165f8V"}},
		{Name: "Bad user insert-V", Want: nil, Args: &TestCaseArgs{Uname: "??&%^&^%&", Uemail: "bcutchie0@live.com", Urole: models.Guest, Uteleg: 4564765786, Upass: "xrqYOB165f8V"}}, // bad user name
		{Name: "Bad user insert-VI", Want: nil, Args: &TestCaseArgs{Uname: "??&%^&^%&", Uemail: "", Urole: models.Guest, Uteleg: 4564765786, Upass: "xrqYOB165f8V"}},                  // bad user email again
	}

	for _, tt := range badtests {
		t.Run(tt.Name, func(t *testing.T) {
			got := uc.NewUser(&models.User{Name: tt.Args.Uname, Email: tt.Args.Uemail, Role: tt.Args.Urole, TelegID: tt.Args.Uteleg, Auth: tt.Args.Upass})
			assert.NotNil(t, got, "Unexpected nil error when creating a new user")
		})
	}

	t.Cleanup(func() {
		ctx := context.Background()
		uc.DbColl.DeleteMany(ctx, bson.M{})
		uc.DbColl.Database().Client().Disconnect(ctx)
	})
}

// func TestFindUser(t *testing.T) {
// 	type arguments struct {
// 		oid string
// 		usr *models.User
// 	}
// 	tests := []struct {
// 		name string
// 		args arguments
// 		want error
// 	}{
// 		{
// 			name: "Simple finduser test",
// 			args: arguments{oid: "662f968131842af60afd8995", usr: &models.User{}},
// 			want: nil,
// 		},
// 		{
// 			name: "Another simple finduser test",
// 			args: arguments{oid: "662f968131842af60afd8995", usr: &models.User{}},
// 			want: nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got := uc.FindUser(tt.args.oid, tt.args.usr)
// 			assert.Equal(t, got, tt.want, "Unexpected error when getting single user ")
// 		})
// 	}
// }
