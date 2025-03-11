package server

import (
	"encoding/json"
	"figenn/internal/database/mocks"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"
)

func TestHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockService(ctrl)

	mockDB.EXPECT().Health().Return(map[string]string{"message": "Hello World"})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	c := e.NewContext(req, resp)

	s := &Server{
		router: e,
		db:     mockDB,
		config: Config{
			JWTSecret: "test-secret",
		},
	}

	if err := s.healthHandler(c); err != nil {
		t.Errorf("handler() error = %v", err)
		return
	}
	if resp.Code != http.StatusOK {
		t.Errorf("handler() wrong status code = %v", resp.Code)
		return
	}
	expected := map[string]string{"message": "Hello World"}
	var actual map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Errorf("handler() error decoding response body: %v", err)
		return
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("handler() wrong response body. expected = %v, actual = %v", expected, actual)
		return
	}
}
