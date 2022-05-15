package web

import (
	"net/http"
	"strings"

	"github.com/PatronC2/Patron/lib/logger"
)

func GetUserIP(request *http.Request) string {
	IPAddress := request.Header.Get("X-Real-Ip")
	if IPAddress != "" {
		IPAddress = strings.Split(IPAddress, ", ")[0]
	} else { // header was empty
		IPAddress = request.Header.Get("X-Forwarded-For")
		if IPAddress == "" { // header was empty
			IPAddress = request.RemoteAddr // extract IP from request (usually least accurate)
		}
	}
	return IPAddress
}

func ClearAuthCookieAndRedirect(writer http.ResponseWriter, request *http.Request, err error) {
	logger.Log(logger.Warning, GetUserIP(request), "| Invalid token -", err.Error())

	// Delete the expired auth cookie
	newAuthCookie := http.Cookie{Name: "auth", Value: "", MaxAge: -1, Secure: true, HttpOnly: true}
	http.SetCookie(writer, &newAuthCookie)

	// Redirect to the `/login` page
	http.Redirect(writer, request, "/login", http.StatusFound)
}
