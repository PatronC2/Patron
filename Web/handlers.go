package main

import (
	"net/http"
)

func errorHandler(writer http.ResponseWriter, request *http.Request, status int) {
	writer.WriteHeader(status)

	switch status {
	case http.StatusNotFound:
		http.ServeFile(writer, request, "Pages/404.html")
	default:
		http.ServeFile(writer, request, "Pages/500.html") // internal server error
	}
}

func HandleRootPage(writer http.ResponseWriter, request *http.Request) {
	// This is also the default route for unknown paths, so check that it's "/" and not something like "/asdf"
	if request.URL.Path != "/" {
		errorHandler(writer, request, http.StatusNotFound)
		return
	}

	http.ServeFile(writer, request, "Pages/index.html")

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
	writer.Write([]byte("Hello, dashboard page! You're authenticated!"))
}

func HandleLoginPage(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, "Pages/login.html")
}

func HandleRegisterPage(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, "Pages/register.html")
}

func HandleAgentsPage(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, "Pages/Agents/agents.html")
}

func HandleLaunchersPage(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, "Pages/Launchers/launchers.html")
}

func HandleListenersPage(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, "Pages/Listeners/listeners.html")
}
