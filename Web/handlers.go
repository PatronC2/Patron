package main

import (
	"net/http"
)

func errorHandler(writer http.ResponseWriter, request *http.Request, status int) {
	writer.WriteHeader(status)

	switch status {
	case http.StatusNotFound:
		http.ServeFile(writer, request, "StatusPages/404.html")
	default:
		writer.Write([]byte("Unknown error"))
	}
}

func HandleRootPage(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		errorHandler(writer, request, http.StatusNotFound)
		return
	}

	http.ServeFile(writer, request, "index.html")

	/* --- TODO: fix later
	// The templated HTML of type template.HTML for proper rendering on the DOM
	homeHTML := returnTemplateHTML(writer, request, "index.html", "handleHomePage", homeContent)

	// Must use template.HTML for proper DOM rendering, otherwise it will be plain text
	layoutContent := map[string]template.HTML{"title": "Scoreboard", "pageContent": homeHTML}
	// Fill the layout with our page-specific templated HTML
	// The layout template automatically includes the header info, navbar, and general layout
	serveLayoutTemplate(writer, request, "handleHomePage", layoutContent)
	*/
}

func HandleDashboardPage(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Hello, dashboard page!"))
}

func HandleLoginPage(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Hello, login page!"))
}

func HandleAgentsPage(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Hello, agents page!"))
}

func HandleLaunchersPage(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Hello, launchers page!"))
}

func HandleListenersPage(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Hello, listeners page!"))
}
