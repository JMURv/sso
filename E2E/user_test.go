package tests

import (
	"bytes"
	"fmt"
	"github.com/JMURv/sso/E2E/helpers"
	controller "github.com/JMURv/sso/internal/controller"
	"github.com/JMURv/sso/internal/validation"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestUserSearch(t *testing.T) {
	router, _ := helpers.SetupRouter()
	tests := []struct {
		q            string
		expectedCode int
	}{
		{
			q:            "jmurv",
			expectedCode: http.StatusOK,
		},
		{
			q:            "jmu",
			expectedCode: http.StatusOK,
		},
		{
			q:            "test",
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.q, func(t *testing.T) {
			r, rr := helpers.SendHttpRequest(
				t, router, "", http.MethodGet, fmt.Sprintf("/api/users/search?q=%s", tc.q), nil,
			)
			assert.Equal(t, http.StatusOK, rr.Code)

			if tc.q == "jmurv" {
				resp := r["data"].([]any)
				assert.Equal(t, 1, len(resp))
				user := resp[0].(map[string]any)
				assert.Equal(t, user["email"], "jmurvz@gmail.com")
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	router, _ := helpers.SetupRouter()
	r, rr := helpers.SendHttpRequest(
		t, router, "", http.MethodGet, "/api/users", nil,
	)

	assert.Equal(t, 2, len(r["data"].([]any)))
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUserRegistration(t *testing.T) {
	router, _ := helpers.SetupRouter()
	defer func() {
		if err := recover(); err != nil {
			t.Logf("Recovered from panic: %v", err)
			helpers.CleanDB(t)
			t.FailNow()
		}
	}()
	defer helpers.CleanDB(t)

	t.Log("Handler started")

	tests := []struct {
		name         string
		userData     map[string]string
		expectedCode int
		expectedMsg  string
		filePath     string
	}{
		{
			name: "Successful Registration",
			userData: map[string]string{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "secret1234",
			},
			expectedCode: http.StatusCreated,
			expectedMsg:  "",
		},
		{
			name: "Missing Email",
			userData: map[string]string{
				"name":     "John Doe",
				"password": "secret123",
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  validation.ErrMissingEmail.Error(),
		},
		{
			name: "Invalid Email Format",
			userData: map[string]string{
				"name":     "John Doe",
				"email":    "invalid-email",
				"password": "secret123",
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  validation.ErrInvalidEmail.Error(),
		},
		{
			name: "Short Password",
			userData: map[string]string{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "short",
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  validation.ErrPassTooShort.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)

			for key, val := range tc.userData {
				_ = writer.WriteField(key, val)
			}

			if tc.filePath != "" {
				file, err := os.Open(tc.filePath)
				defer file.Close()

				if !assert.NoError(t, err) {
					t.Fatal(err)
				}

				part, err := writer.CreateFormFile("file", "testfile.pdf")
				if !assert.NoError(t, err) {
					t.Fatal(err)
				}

				_, err = io.Copy(part, file)
				if !assert.NoError(t, err) {
					t.Fatal(err)
				}
			}

			err := writer.Close()
			if !assert.NoError(t, err) {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, "/api/users", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			router.ServeHTTP(rr, req)
			if !assert.NoError(t, err) {
				t.Fatal(err)
			}
			assert.Equal(t, tc.expectedCode, rr.Code)

			r := helpers.UnmarshallResponse(t, rr.Body)

			if tc.expectedCode == http.StatusCreated {
				createdUser, ok := r["data"].(map[string]any)["user"].(map[string]any)
				assert.True(t, ok, "Expected a nested user object in the response")
				assert.Equal(t, tc.userData["name"], createdUser["name"])
				assert.Equal(t, tc.userData["email"], createdUser["email"])
				assert.NotEqual(t, tc.userData["password"], createdUser["password"])
				assert.NotEqual(t, "", createdUser["id"])

				cookies := rr.Result().Cookies()
				assert.Equal(t, 2, len(cookies))

				var access, refresh *http.Cookie
				for _, v := range cookies {
					switch v.Name {
					case "access":
						access = v
					case "refresh":
						refresh = v
					}
				}

				assert.NotNil(t, access, "Expected access cookie to be set")
				assert.NotEqual(t, "", access.Value, "Expected access cookie to have a value")

				assert.NotNil(t, refresh, "Expected refresh cookie to be set")
				assert.NotEqual(t, "", refresh.Value, "Expected refresh cookie to have a value")
			} else {
				errorMsg, ok := r["error"].(string)
				assert.True(t, ok, "Expected an error message in the response")
				assert.Contains(t, errorMsg, tc.expectedMsg)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	router, _ := helpers.SetupRouter()
	tests := []struct {
		name         string
		id           string
		expectedCode int
	}{
		{
			name:         "Invalid UUID",
			id:           "123",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Not Found User",
			id:           uuid.New().String(),
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, rr := helpers.SendHttpRequest(
				t, router, "", http.MethodGet, fmt.Sprintf("/api/users/%s", tc.id), nil,
			)
			assert.Equal(t, tc.expectedCode, rr.Code)

		})
	}
}

//func TestSendLoginCode(t *testing.T) {
//	router, _, clean := helpers.SetupRouter()
//	defer func() {
//		if err := recover(); err != nil {
//			t.Logf("Recovered from panic: %v", err)
//			clean(t)
//			t.FailNow()
//		}
//	}()
//	defer clean(t)
//
//	user, _ := helpers.CreateUser(router)
//
//	tests := []struct {
//		name         string
//		userData     map[string]string
//		expectedCode int
//		expectedMsg  string
//	}{
//		{
//			name: "Invalid Email",
//			userData: map[string]string{
//				"email":    "invalid-email",
//				"password": "test-password",
//			},
//			expectedCode: http.StatusBadRequest,
//			expectedMsg:  validation.ErrInvalidEmail.Error(),
//		},
//		{
//			name: "No password",
//			userData: map[string]string{
//				"email": user["email"].(string),
//			},
//			expectedCode: http.StatusBadRequest,
//			expectedMsg:  handler.ErrEmailAndPasswordRequired.Error(),
//		},
//		{
//			name: "Valid",
//			userData: map[string]string{
//				"email":    user["email"].(string),
//				"password": "secret1234",
//			},
//			expectedCode: http.StatusOK,
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			body, err := json.Marshal(tc.userData)
//			if !assert.NoError(t, err) {
//				t.Fatal(err)
//			}
//
//			rr := httptest.NewRecorder()
//			req, err := http.NewRequest(http.MethodPost, "/api/sso/send-login-code", bytes.NewBuffer(body))
//			req.Header.Set("Content-Type", "application/json")
//			router.ServeHTTP(rr, req)
//
//			if !assert.NoError(t, err) {
//				t.Fatal(err)
//			}
//			assert.Equal(t, tc.expectedCode, rr.Code)
//
//			data, err := io.ReadAll(rr.Body)
//			if !assert.NoError(t, err) {
//				t.Fatal(err)
//			}
//
//			var r map[string]any
//			if err = json.Unmarshal(data, &r); !assert.NoError(t, err) {
//				t.Fatal(err)
//			}
//
//			if tc.expectedMsg != "" {
//				errorMsg, ok := r["error"].(string)
//				assert.True(t, ok, "Expected an error message in the response")
//				assert.Contains(t, errorMsg, tc.expectedMsg)
//			}
//		})
//	}
//}

//	func TestCheckLoginCode(t *testing.T) {
//		ctx := context.Background()
//		router, cache, clean := helpers.SetupRouter()
//		defer func() {
//			if err := recover(); err != nil {
//				t.Logf("Recovered from panic: %v", err)
//				clean(t)
//				t.FailNow()
//			}
//		}()
//		defer clean(t)
//
//		user, _ := helpers.CreateUser(router)
//		code, err := cache.GetCode(ctx, fmt.Sprintf("code:%v", user["email"]))
//		if !assert.NoError(t, err) {
//			t.Fatal(err)
//		}
//
//		tests := []struct {
//			name         string
//			userData     map[string]string
//			expectedCode int
//			expectedMsg  string
//		}{
//			{
//				name: "Invalid Email",
//				userData: map[string]string{
//					"email": "invalid-email",
//					"code":  "1234",
//				},
//				expectedCode: http.StatusBadRequest,
//				expectedMsg:  validation.ErrInvalidEmail.Error(),
//			},
//			{
//				name: "No code",
//				userData: map[string]string{
//					"email": user["email"].(string),
//				},
//				expectedCode: http.StatusBadRequest,
//				expectedMsg:  handler.ErrEmailAndCodeRequired.Error(),
//			},
//			{
//				name: "No email",
//				userData: map[string]string{
//					"code": "1234",
//				},
//				expectedCode: http.StatusBadRequest,
//				expectedMsg:  handler.ErrEmailAndCodeRequired.Error(),
//			},
//			{
//				name: "Not found user",
//				userData: map[string]string{
//					"email": "test@gmail.com",
//					"code":  strconv.Itoa(code),
//				},
//				expectedCode: http.StatusNotFound,
//			},
//			{
//				name: "Valid",
//				userData: map[string]string{
//					"email": user["email"].(string),
//					"code":  strconv.Itoa(code),
//				},
//				expectedCode: http.StatusOK,
//			},
//		}
//
//		for _, tc := range tests {
//			t.Run(tc.name, func(t *testing.T) {
//				body, err := json.Marshal(tc.userData)
//				if !assert.NoError(t, err) {
//					t.Fatal(err)
//				}
//
//				rr := httptest.NewRecorder()
//				req, err := http.NewRequest(http.MethodPost, "/api/sso/check-login-code", bytes.NewBuffer(body))
//				req.Header.Set("Content-Type", "application/json")
//				router.ServeHTTP(rr, req)
//
//				if !assert.NoError(t, err) {
//					t.Fatal(err)
//				}
//				assert.Equal(t, tc.expectedCode, rr.Code)
//
//				data, err := io.ReadAll(rr.Body)
//				if !assert.NoError(t, err) {
//					t.Fatal(err)
//				}
//
//				var r map[string]any
//				if err = json.Unmarshal(data, &r); !assert.NoError(t, err) {
//					t.Fatal(err)
//				}
//
//				if tc.expectedMsg != "" {
//					errorMsg, ok := r["error"].(string)
//					assert.True(t, ok, "Expected an error message in the response")
//					assert.Contains(t, errorMsg, tc.expectedMsg)
//				}
//			})
//		}
//	}
func TestUpdateUsers(t *testing.T) {
	router, _ := helpers.SetupRouter()
	defer func() {
		if err := recover(); err != nil {
			t.Logf("Recovered from panic: %v", err)
			helpers.CleanDB(t)
			t.FailNow()
		}
	}()
	defer helpers.CleanDB(t)

	user, access := helpers.CreateUser(router)
	tests := []struct {
		name         string
		id           string
		expectedCode int
		expectedMsg  string
		userData     map[string]string
	}{
		{
			name:         "Invalid UUID",
			id:           "123",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Failed to decode request body",
			id:           user["id"].(string),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Missing Name",
			id:           user["id"].(string),
			expectedCode: http.StatusBadRequest,
			expectedMsg:  validation.ErrMissingName.Error(),
			userData: map[string]string{
				"email": "john@example.com",
			},
		},
		{
			name:         "Missing Email",
			id:           user["id"].(string),
			expectedCode: http.StatusBadRequest,
			expectedMsg:  validation.ErrMissingEmail.Error(),
			userData: map[string]string{
				"name": "John Doe",
			},
		},
		{
			name:         "Not Found User",
			id:           uuid.New().String(),
			expectedCode: http.StatusNotFound,
			expectedMsg:  controller.ErrNotFound.Error(),
			userData: map[string]string{
				"name":  "John Doe",
				"email": "john@example.com",
			},
		},
		{
			name:         "Valid",
			id:           user["id"].(string),
			expectedCode: http.StatusOK,
			userData: map[string]string{
				"name":  "John Doe UPDATED",
				"email": "johnUPDATED@example.com",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.userData)
			if !assert.NoError(t, err) {
				t.Fatal(err)
			}

			r, rr := helpers.SendHttpRequest(
				t, router, access, http.MethodPut, fmt.Sprintf("/api/users/%s", tc.id), bytes.NewBuffer(body),
			)

			assert.Equal(t, tc.expectedCode, rr.Code)
			if tc.expectedCode == http.StatusOK {
				createdUser, ok := r["data"].(map[string]any)
				assert.True(t, ok, "Expected a nested user object in the response")
				assert.Equal(t, tc.userData["name"], createdUser["name"])
				assert.Equal(t, tc.userData["email"], createdUser["email"])
				assert.NotEqual(t, user["name"], createdUser["name"])
				assert.NotEqual(t, user["email"], createdUser["email"])
			} else {
				errorMsg, ok := r["error"].(string)
				assert.True(t, ok, "Expected an error message in the response")
				assert.Contains(t, errorMsg, tc.expectedMsg)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	router, _ := helpers.SetupRouter()
	defer helpers.CleanDB(t)

	user, access := helpers.CreateUser(router)
	tests := []struct {
		name         string
		id           string
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "Invalid UUID",
			id:           "123",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Valid",
			id:           user["id"].(string),
			expectedCode: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, rr := helpers.SendHttpRequest(
				t, router, access, http.MethodDelete, fmt.Sprintf("/api/users/%s", tc.id), nil,
			)

			assert.Equal(t, tc.expectedCode, rr.Code)
		})
	}
}
