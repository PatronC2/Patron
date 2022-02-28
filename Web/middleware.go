package main

import (
	"errors"
	"net/http"

	"github.com/PatronC2/Patron/lib/web"
)

// JWT authentication middleware to authenticated pages and endpoints
func IsAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			tokenClaims, err := web.GetAuthClaims(writer, request)
			if err != nil {
				web.ClearAuthCookieAndRedirect(writer, request, err)
				return
			}

			// Ensure the claims we need exist in the first place
			if tokenClaims["user"] == nil || tokenClaims["perms"] == nil {
				web.ClearAuthCookieAndRedirect(writer, request, errors.New("token claims don't exist"))
				return
			}

			// NOTE: Go's JSON unmarshalling decodes JSON numbers to type float64

			// Give the claims a Go type
			var tokenUser string = tokenClaims["user"].(string)

			var tokenPermissions web.UserPermissions = tokenClaims["perms"].(web.UserPermissions)

			// Validate the claims
			if tokenUser == "" {
				web.ClearAuthCookieAndRedirect(writer, request, errors.New("token claim 'user' is invalid"))
				return
			}
			if tokenPermissions != web.PERMS_ADMIN {
				web.ClearAuthCookieAndRedirect(writer, request, errors.New("token claim 'perms' is invalid"))
				return
			}

			// If everything was successful, navigate to the page
			endpoint(writer, request)
		},
	)
}
