package endpoints

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/johannaojeling/go-rest-api/pkg/repositories"
)

var (
	columns = []string{"id", "first_name", "last_name", "email"}
)

type Suite struct {
	suite.Suite
	mock   sqlmock.Sqlmock
	router *gin.Engine
}

func (s *Suite) SetupSuite() {
	conn, mock, err := sqlmock.New()
	if err != nil {
		s.T().Fatalf("error creating sql mock: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		Conn: conn,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		s.T().Fatalf("error opening db connection: %v", err)
	}

	router := gin.Default()
	userRepository := repositories.NewSQLUserRepository(db)
	handler := NewUsersHandler(userRepository)
	handler.Register(router.Group(""))

	s.mock = mock
	s.router = router
}

func (s *Suite) AfterTest(_, _ string) {
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func TestUsersHandler(t *testing.T) {
	suite.Run(t, &Suite{})
}

func (s *Suite) TestUsersHandler_CreateUser() {
	testCases := []struct {
		requestBody  map[string]interface{}
		returnRow    []driver.Value
		expectedCode int
		expectedBody map[string]interface{}
		reason       string
	}{
		{
			requestBody: map[string]interface{}{
				"first_name": "Jane",
				"last_name":  "Doe",
				"email":      "jane.doe@mail.com",
			},
			returnRow:    []driver.Value{"abc123", "Jane", "Doe", "jane.doe@mail.com"},
			expectedCode: 201,
			expectedBody: map[string]interface{}{
				"id":         "abc123",
				"first_name": "Jane",
				"last_name":  "Doe",
				"email":      "jane.doe@mail.com",
			},
			reason: "Should return status 201 and newly created user",
		},
		{
			requestBody: map[string]interface{}{
				"first_name": "Jane",
				"last_name":  "Doe",
			},
			returnRow:    nil,
			expectedCode: 400,
			expectedBody: map[string]interface{}{
				"details": "invalid request body",
			},
			reason: "Should return status 400 and details when request body is invalid",
		},
	}

	for i, tc := range testCases {
		s.T().Run(fmt.Sprintf("Test %d: %s", i, tc.reason), func(t *testing.T) {
			if tc.returnRow != nil {
				query := `INSERT INTO "users" ("first_name","last_name","email","created_at","updated_at") ` +
					`VALUES ($1,$2,$3,$4,$5) RETURNING "id"`
				rows := sqlmock.NewRows(columns).AddRow(tc.returnRow...)

				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(tc.requestBody["first_name"], tc.requestBody["last_name"], tc.requestBody["email"], sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(rows)
				s.mock.ExpectCommit()
			}

			jsonBody, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("error marshaling reqeust body to json %v", err)
			}

			request, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatalf("error creating request %v", err)
			}

			recorder := httptest.NewRecorder()
			s.router.ServeHTTP(recorder, request)

			assert.Equal(t, tc.expectedCode, recorder.Code, "Should match response code")

			var actualBody map[string]interface{}
			err = json.Unmarshal(recorder.Body.Bytes(), &actualBody)
			if err != nil {
				t.Fatalf("error unmarshaling response: %v", err)
			}

			assert.Equal(
				t,
				tc.expectedBody,
				actualBody,
				"Should match response body",
			)
		})
	}
}

func (s *Suite) TestUsersHandler_GetUser() {
	testCases := []struct {
		id           string
		returnRow    []driver.Value
		expectedCode int
		expectedBody map[string]interface{}
		reason       string
	}{
		{
			id:           "abc123",
			returnRow:    []driver.Value{"abc123", "Jane", "Doe", "jane.doe@mail.com"},
			expectedCode: 200,
			expectedBody: map[string]interface{}{
				"id":         "abc123",
				"first_name": "Jane",
				"last_name":  "Doe",
				"email":      "jane.doe@mail.com",
			},
			reason: "Should return status 200 and user with id 'abc123'",
		},
		{
			id:           "abc123",
			returnRow:    nil,
			expectedCode: 404,
			expectedBody: map[string]interface{}{
				"details": "no user with id \"abc123\" exists",
			},
			reason: "Should return status 404 and details when no user with id exists",
		},
	}

	for i, tc := range testCases {
		s.T().Run(fmt.Sprintf("Test %d: %s", i, tc.reason), func(t *testing.T) {
			query := `SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT 1`
			rows := sqlmock.NewRows(columns)

			if tc.returnRow != nil {
				rows = rows.AddRow(tc.returnRow...)
			}

			s.mock.ExpectQuery(regexp.QuoteMeta(query)).
				WithArgs(tc.id).
				WillReturnRows(rows)

			request, err := http.NewRequest("GET", "/"+tc.id, nil)
			if err != nil {
				t.Fatalf("error creating request %v", err)
			}

			recorder := httptest.NewRecorder()
			s.router.ServeHTTP(recorder, request)

			assert.Equal(t, tc.expectedCode, recorder.Code, "Should match response code")

			var actualBody map[string]interface{}
			err = json.Unmarshal(recorder.Body.Bytes(), &actualBody)
			if err != nil {
				t.Fatalf("error unmarshaling response: %v", err)
			}

			assert.Equal(
				t,
				tc.expectedBody,
				actualBody,
				"Should match response body",
			)
		})
	}
}

func (s *Suite) TestUsersHandler_GetUsers() {
	testCases := []struct {
		returnRows   [][]driver.Value
		expectedCode int
		expectedBody []map[string]interface{}
		reason       string
	}{
		{
			returnRows: [][]driver.Value{
				{"abc123", "Jane", "Doe", "jane.doe@mail.com"},
				{"abc124", "John", "Doe", "john.doe@mail.com"},
			},
			expectedCode: 200,
			expectedBody: []map[string]interface{}{
				{
					"id":         "abc123",
					"first_name": "Jane",
					"last_name":  "Doe",
					"email":      "jane.doe@mail.com",
				},
				{
					"id":         "abc124",
					"first_name": "John",
					"last_name":  "Doe",
					"email":      "john.doe@mail.com",
				},
			},
			reason: "Should return status 200 and list of users",
		},
		{
			returnRows:   [][]driver.Value{},
			expectedCode: 200,
			expectedBody: []map[string]interface{}{},
			reason:       "Should return status 200 and empty list",
		},
	}

	for i, tc := range testCases {
		s.T().Run(fmt.Sprintf("Test %d: %s", i, tc.reason), func(t *testing.T) {
			query := `SELECT * FROM "users"`

			rows := sqlmock.NewRows(columns)
			for _, row := range tc.returnRows {
				rows = rows.AddRow(row...)
			}

			s.mock.ExpectQuery(regexp.QuoteMeta(query)).
				WillReturnRows(rows)

			request, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("error creating request %v", err)
			}

			recorder := httptest.NewRecorder()
			s.router.ServeHTTP(recorder, request)

			assert.Equal(t, tc.expectedCode, recorder.Code, "Should match response code")

			var actualBody []map[string]interface{}
			err = json.Unmarshal(recorder.Body.Bytes(), &actualBody)
			if err != nil {
				t.Fatalf("error unmarshaling response: %v", err)
			}

			assert.Equal(
				t,
				tc.expectedBody,
				actualBody,
				"Should match response body",
			)
		})
	}
}

func (s *Suite) TestUsersHandler_UpdateUser() {
	testCases := []struct {
		id              string
		requestBody     map[string]interface{}
		updateReturnRow []driver.Value
		createReturnRow []driver.Value
		expectedCode    int
		expectedBody    map[string]interface{}
		reason          string
	}{
		{
			id: "abc123",
			requestBody: map[string]interface{}{
				"first_name": "Jane",
				"last_name":  "Doe",
				"email":      "jane@mail.com",
			},
			updateReturnRow: []driver.Value{"abc123", "Jane", "Doe", "jane.doe@mail.com"},
			expectedCode:    200,
			expectedBody: map[string]interface{}{
				"id":         "abc123",
				"first_name": "Jane",
				"last_name":  "Doe",
				"email":      "jane@mail.com",
			},
			reason: "Should return status 200 and newly updated user when user with id exists",
		},
		{
			id: "abc123",
			requestBody: map[string]interface{}{
				"first_name": "Jane",
				"last_name":  "Doe",
				"email":      "jane@mail.com",
			},
			updateReturnRow: nil,
			createReturnRow: []driver.Value{"abc123", "Jane", "Doe", "jane@mail.com"},
			expectedCode:    201,
			expectedBody: map[string]interface{}{
				"id":         "abc123",
				"first_name": "Jane",
				"last_name":  "Doe",
				"email":      "jane@mail.com",
			},
			reason: "Should return status 201 and newly created user when no user with id exists",
		},
	}

	for i, tc := range testCases {
		s.T().Run(fmt.Sprintf("Test %d: %s", i, tc.reason), func(t *testing.T) {
			selectQuery := `SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT 1`
			updateRows := sqlmock.NewRows(columns)

			if tc.updateReturnRow != nil {
				updateRows = updateRows.AddRow(tc.updateReturnRow...)
			}

			s.mock.ExpectQuery(regexp.QuoteMeta(selectQuery)).
				WithArgs(tc.id).
				WillReturnRows(updateRows)

			if tc.updateReturnRow != nil {
				updateQuery := `UPDATE "users" SET "first_name"=$1,"last_name"=$2,"email"=$3,"updated_at"=$4 WHERE "id" = $5`
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(updateQuery)).
					WithArgs(tc.requestBody["first_name"], tc.requestBody["last_name"], tc.requestBody["email"], sqlmock.AnyArg(), tc.id).
					WillReturnResult(sqlmock.NewResult(1, 1))
				s.mock.ExpectCommit()
			} else {
				createQuery := `INSERT INTO "users" ("first_name","last_name","email","created_at","updated_at","id") ` +
					`VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`
				createRows := sqlmock.NewRows(columns).AddRow(tc.createReturnRow...)

				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(createQuery)).
					WithArgs(tc.requestBody["first_name"], tc.requestBody["last_name"], tc.requestBody["email"], sqlmock.AnyArg(), sqlmock.AnyArg(), tc.id).
					WillReturnRows(createRows)
				s.mock.ExpectCommit()
			}

			jsonBody, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("error marshaling reqeust body to json %v", err)
			}

			request, err := http.NewRequest("PUT", "/"+tc.id, bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatalf("error creating request %v", err)
			}

			recorder := httptest.NewRecorder()
			s.router.ServeHTTP(recorder, request)

			assert.Equal(t, tc.expectedCode, recorder.Code, "Should match response code")

			var actualBody map[string]interface{}
			err = json.Unmarshal(recorder.Body.Bytes(), &actualBody)
			if err != nil {
				t.Fatalf("error unmarshaling response: %v", err)
			}

			assert.Equal(
				t,
				tc.expectedBody,
				actualBody,
				"Should match response body",
			)
		})
	}
}

func (s *Suite) TestUsersHandler_DeleteUser() {
	testCases := []struct {
		id           string
		returnRow    []driver.Value
		expectedCode int
		expectedBody map[string]interface{}
		reason       string
	}{
		{
			id:           "abc123",
			returnRow:    []driver.Value{"abc123", "Jane", "Doe", "jane.doe@mail.com"},
			expectedCode: 204,
			expectedBody: nil,
			reason:       "Should return status 204 and delete user when user with id exists",
		},
		{
			id:           "abc123",
			returnRow:    nil,
			expectedCode: 404,
			expectedBody: map[string]interface{}{
				"details": "no user with id \"abc123\" exists",
			},
			reason: "Should return status 404 and details when no user with id exists",
		},
	}

	for i, tc := range testCases {
		s.T().Run(fmt.Sprintf("Test %d: %s", i, tc.reason), func(t *testing.T) {
			selectQuery := `SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT 1`
			rows := sqlmock.NewRows(columns)

			if tc.returnRow != nil {
				rows = rows.AddRow(tc.returnRow...)
			}

			s.mock.ExpectQuery(regexp.QuoteMeta(selectQuery)).
				WithArgs(tc.id).
				WillReturnRows(rows)

			if tc.returnRow != nil {
				deleteQuery := `DELETE FROM "users" WHERE "users"."id" = $1`
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(deleteQuery)).
					WithArgs(tc.id).
					WillReturnResult(sqlmock.NewResult(1, 1))
				s.mock.ExpectCommit()
			}

			request, err := http.NewRequest("DELETE", "/"+tc.id, nil)
			if err != nil {
				t.Fatalf("error creating request %v", err)
			}

			recorder := httptest.NewRecorder()
			s.router.ServeHTTP(recorder, request)

			assert.Equal(t, tc.expectedCode, recorder.Code, "Should match response code")

			if tc.returnRow == nil {
				var actualBody map[string]interface{}
				err = json.Unmarshal(recorder.Body.Bytes(), &actualBody)
				if err != nil {
					t.Fatalf("error unmarshaling response: %v", err)
				}

				assert.Equal(
					t,
					tc.expectedBody,
					actualBody,
					"Should match response body",
				)
			}
		})
	}
}
