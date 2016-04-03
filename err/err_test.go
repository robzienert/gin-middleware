package err

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func ExampleController_Simple() {
	gin.SetMode(gin.TestMode)

	someFunc := func() error {
		return errors.New("some function that returns an error")
	}

	r := gin.New()
	r.Use(ErrorHandler())

	r.GET("/boom", func(c *gin.Context) {
		if err := someFunc(); err != nil {
			c.Error(err)
			c.Error(errors.New("This is a public-friendly error")).
				SetType(gin.ErrorTypePublic).
				SetMeta(GinErrorMeta{Code: "ABCD1234"})
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	})
}

func TestErrorHandler_InternalServerError_Public(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/boom", func(c *gin.Context) {
		c.Error(errors.New("something awful"))
		c.AbortWithError(http.StatusInternalServerError, errors.New("last error"))
	})
	req, _ := http.NewRequest("GET", "/boom", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, 500, resp.Code)

	expected, _ := json.Marshal(ErrorResponse{
		ID:    "",
		Title: "last error",
	})
	assert.JSONEq(t, string(expected), resp.Body.String())
}

func TestErrorHandler_InternalServerError_Private(t *testing.T) {
	gin.SetMode(gin.DebugMode)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/boom", func(c *gin.Context) {
		c.Error(errors.New("something awful"))
		c.Error(errors.New("get dat meta"))
		c.AbortWithError(http.StatusInternalServerError, errors.New("last error")).
			SetMeta(GinErrorMeta{Code: "wow such meta"})
	})
	req, _ := http.NewRequest("GET", "/boom", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, 500, resp.Code)

	expected, _ := json.Marshal(ErrorResponse{
		ID:    "",
		Code:  "wow such meta",
		Title: "last error",
		Detail: `something awful
get dat meta
last error`,
	})
	assert.JSONEq(t, string(expected), resp.Body.String())
}
