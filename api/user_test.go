package api

import (
	mockdb "SimpleBank/db/mock"
	db "SimpleBank/db/sqlc"
	"SimpleBank/db/util"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type EqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e EqCreateUserParamsMatcher) Matches(x interface{}) bool {
	fmt.Println("log2")

	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}
	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}
	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e EqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("match arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParam(arg db.CreateUserParams, password string) gomock.Matcher {
	fmt.Println("log1")
	return EqCreateUserParamsMatcher{arg, password}
}

// type H map[string]any
func TestCreateUser(t *testing.T) {
	user, password := RandomUser(t)
	hashedPassword, err := util.HashedPassword(password)
	require.NoError(t, err)
	testCases := []struct {
		name          string
		body          gin.H
		stubs         func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"UserName": user.Username,
				"Password": password,
				"FullName": user.FullName,
				"Email":    user.Email,
			},
			stubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username:       user.Username,
					FullName:       user.FullName,
					HashedPassword: hashedPassword,
					Email:          user.Email,
				}

				store.
					EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParam(arg, password)).
					Times(1).
					Return(user, nil)

				//store.EXPECT().CreateUser(gomock.Any(), gomock.Eq(arg)).Times(1).Return(user, nil)
				//Got: {wrwees $2a$10$CNAaRjqC/XV.5P5ryqwhb.E51ucRqW2TU8U8gdvyJyQ5WYLDO8me2 qqwgnw svjvoi@email.com} (db.CreateUserParams)
				//Want: is equal to {wrwees $2a$10$2KXptilx5oiCg12mECeGSuy6mn2rvK1z2hgSJxY2Mn1WEczormxJa qqwgnw svjvoi@email.com}

				//the above usage will give error of failing as body password being converted to hash after api is is hit will be different ,
				//than whats being direcctly stored in db thriugh querier interface's create user function as bcrypt produces different hashes
				//each time for plain text

				//store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(1).Return(user, nil)
				//if we use gomock.Any(), it will pass the testcases irrespctive of user object validation
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		// {
		// 	name: "InternalError",
		// 	body: gin.H{
		// 		"Username": user.Username,
		// 		"Password": password,
		// 		"FullName": user.FullName,
		// 		"Email":    user.Email,
		// 	},
		// 	stubs: func(store *mockdb.MockStore) {
		// 		store.
		// 			EXPECT().
		// 			CreateUser(gomock.Any(), gomock.Any()).
		// 			Times(1).
		// 			Return(db.User{}, sql.ErrConnDone)
		// 	},
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusInternalServerError, recorder.Code)
		// 	},
		// },
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			testCase := testCases[i]

			store := mockdb.NewMockStore(ctrl)
			testCase.stubs(store)
			server := newTestServer(t, store)
			data, err := json.Marshal(&testCase.body)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()
			url := "/users"
			httpRequest, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			server.router.ServeHTTP(recorder, httpRequest)
			testCase.checkResponse(t, recorder)

		})
	}

}

func TestGetUser(t *testing.T) {

}

func RandomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)

	hashedPassword, err := util.HashedPassword(password)
	if err != nil {
		return
	}

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	return

}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data := body.Bytes()
	require.NotEmpty(t, data)

	var gotUser db.User
	err := json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.NotEmpty(t, gotUser.HashedPassword)
}
