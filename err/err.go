// Package err provides a standard contextual JSON error handler.
package err

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is a jsonapi.org-compatible error object that should be used
// in all error conditions.
type ErrorResponse struct {
	ID     string            `json:"id"`
	Title  string            `json:"title"`
	Code   string            `json:"code,omitempty"`
	Detail string            `json:"detail,omitempty"`
	Meta   map[string]string `json:"meta,omitempty"`
}

// GinErrorMeta is a wrapper for the Gin Error Meta interface, which allows us
// to easily pass an error code if we want to the
type GinErrorMeta struct {
	Code  string
	Extra map[string]string
}

// ErrorHandler is the standard error handler for the application. This handler
// is not a recovery handler, however, so panics will continue to cause a blank
// 500 error while printing the panic to logs.
//
// Errors can have multiple types, but functionally this handler only concerns
// with public vs private. A public error will always be visible in responses,
// and will never be logged: This is good for validation error messages. Private
// errors will always show in logs, but will only be exposed in responses when
// the application is running in debug mode.
//
// Simple usage:
//
// foo, err := something.NotWorking()
// if err != nil {
//   c.Error(err)
//   c.Error(errors.New("This is a public error")).SetType(gin.ErrorTypePublic)
//   c.Error(errors.New("Bah")).Meta(GinErrorMeta{Code: "ABCD1234"})
//   c.AbortWithStatus(http.StatusInternalServerError)
//   return
// }
//
// Multiple errors can be chained together in a single request if necessary,
// where the last one will be used to display the final error title. Example:
//
// func FooIsValid(c *gin.Context, foo model.Foo) bool {
//   if foo.Bar == "" {
//     c.Error(errors.New("Bar cannot be empty")).Type(gin.ErrorTypePublic)
//   }
//   if foo.Baz != "baz" {
//     c.Error(errors.New("Baz must be baz")).Type(gin.ErrorTypePublic)
//   }
// }
//
// if !FooIsValid(c, &foo) {
//   c.AbortWithError(http.StatusBadRequest, errors.New("Foo is not valid"))
//   return
// }
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			uuid, ok := c.Get("uuid")
			if !ok {
				uuid = ""
			}
			if gin.IsDebugging() {
				errors := c.Errors
				last := c.Errors.Last()
				resp := ErrorResponse{
					ID:     uuid.(string),
					Title:  last.Error(),
					Detail: strings.Join(errors.Errors(), "\n"),
				}

				if last.Meta != nil {
					if errMeta, ok := last.Meta.(GinErrorMeta); ok {
						resp.Code = errMeta.Code
						resp.Meta = errMeta.Extra
					}
				}
				// -1 doesn't overwrite whatever the code was before.
				c.JSON(-1, resp)
			} else {
				errors := c.Errors.ByType(gin.ErrorTypePublic)
				last := c.Errors.Last()
				resp := ErrorResponse{
					ID:     uuid.(string),
					Title:  last.Error(),
					Detail: strings.Join(errors.Errors(), "\n"),
				}
				if last.Meta != nil {
					if errMeta, ok := last.Meta.(GinErrorMeta); ok {
						resp.Code = errMeta.Code
						resp.Meta = errMeta.Extra
					}
				}
				c.JSON(-1, resp)
			}
		}
	}
}
