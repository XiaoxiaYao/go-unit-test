package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"webapp/pkg/data"
	mock_repository "webapp/pkg/repository/mock"

	"github.com/golang/mock/gomock"
)

func Test_app_authenticate(t *testing.T) {
	user := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
		Password:  "$2a$14$ajq8Q7fbtFRQvXpdCq7Jcuy.Rx1h/L4J60Otx.gyNLbAYctGMJ9tK",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var theTests = []struct {
		name               string
		requestBody        string
		buildStubs         func(mockDB *mock_repository.MockDatabaseRepo)
		expectedStatusCode int
	}{
		{"valid user", `{"email":"admin@example.com","password":"secret"}`, func(mockDB *mock_repository.MockDatabaseRepo) {
			mockDB.EXPECT().GetUserByEmail(gomock.Any()).Times(1).Return(&user, nil)
		}, http.StatusOK},
		{"not json", `I'm not JSON`, func(mockDB *mock_repository.MockDatabaseRepo) {
			mockDB.EXPECT().GetUserByEmail(gomock.Any()).Times(0)
		}, http.StatusUnauthorized},
		{"empty json", `{}`, func(mockDB *mock_repository.MockDatabaseRepo) {
			mockDB.EXPECT().GetUserByEmail(gomock.Any()).Times(1).Return(nil, errors.New("not found"))
		}, http.StatusUnauthorized},
		{"empty email", `{"email":""}`, func(mockDB *mock_repository.MockDatabaseRepo) {
			mockDB.EXPECT().GetUserByEmail(gomock.Any()).Times(1).Return(nil, errors.New("not found"))
		}, http.StatusUnauthorized},
		{"empty password", `{"email":"admin@example.com"}`, func(mockDB *mock_repository.MockDatabaseRepo) {
			mockDB.EXPECT().GetUserByEmail(gomock.Any()).Times(1).Return(&user, nil)
		}, http.StatusUnauthorized},
		{"invalid user", `{"email":"admin@someotherdomain.com","password":"secret"}`, func(mockDB *mock_repository.MockDatabaseRepo) {
			mockDB.EXPECT().GetUserByEmail(gomock.Any()).Times(1).Return(nil, errors.New("not found"))
		}, http.StatusUnauthorized},
	}

	for _, e := range theTests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mock_repository.NewMockDatabaseRepo(ctrl)
		e.buildStubs(mockDB)
		app.DB = mockDB

		var reader io.Reader
		reader = strings.NewReader(e.requestBody)
		req, _ := http.NewRequest("POST", "/auth", reader)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.authenticate)

		handler.ServeHTTP(rr, req)

		if e.expectedStatusCode != rr.Code {
			t.Errorf("%s: returned wrong status code; expected %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
	}
}
