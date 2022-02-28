package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/lib/sqlite"
	"github.com/PatronC2/Patron/lib/system"
	"github.com/PatronC2/Patron/lib/utils"
)

func main() {
	var argTest bool
	var argPort int
	flag.BoolVar(&argTest, "test", false, "Listen on localhost instead of the default interface's IP address")
	flag.IntVar(&argPort, "port", 443, "Port to listen on")
	flag.Parse()

	logger.Log(logger.Debug, "----------Initializing----------")

	var currentDirectory, certPath, privateKeyPath, listenAddress string
	var db *sql.DB
	var listenIP net.IP

	currentDirectory, err := os.Getwd()
	if err != nil {
		logger.LogError(err)
		os.Exit(logger.ERR_UNKNOWN)
	}

	db = sqlite.GetDatabaseHandle()
	defer utils.Close(db)

	// cert, err := tls.LoadX509KeyPair(utils.CurrentDirectory+"/pwnts.red.pem", utils.CurrentDirectory+"/pwnts_server_key.pem")
	// if err != nil {
	// 	logger.Log(utils.Error, "Couldn't load X509 keypair")
	// 	os.Exit(1)
	// }

	// tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
	// listener, err := tls.Listen("tcp", listenPort, &tlsConfig)
	// if err != nil {
	// 	logger.Log(utils.Error, "Couldn't set up listener")
	// 	os.Exit(1)
	// }
	// defer listener.Close()

	certPath = currentDirectory + "/patron_cert.pem"
	privateKeyPath = currentDirectory + "/patron_key.pem"

	// cert, err := os.ReadFile(certPath)
	// utils.CheckErrorExit(utils.Error, err, utils.ERR_GENERIC, "Cannot read certificate file '"+certPath+"'")
	// privateKey, err := os.ReadFile(privateKeyPath)
	// utils.CheckErrorExit(utils.Error, err, utils.ERR_GENERIC, "Cannot read private key file '"+privateKeyPath+"'")

	/*
		--- Main site ---
	*/
	if argTest {
		listenIP = net.ParseIP("127.0.0.1")
	} else {
		listenIP = system.GetHostIP()
	}

	listenAddress = fmt.Sprintf("%s:%d", listenIP.String(), argPort)

	logger.Log(logger.Done, "Running HTTPS server at", listenAddress)
	logger.Log(logger.Debug, "----------Activity Logs---------")

	// https://pkg.go.dev/net/http#FileServer
	// Allow the hosting of static files like our images and stylesheets
	staticFileServer := http.FileServer(http.Dir(currentDirectory + "/static"))
	// TODO: Figure out why the FileServer isn't setting the correct Content-Type header MIME type on JavaScript files.
	//		 Should be "text/javascript" but is "text/plain"
	http.Handle("/static/", http.StripPrefix("/static/", staticFileServer))

	// Register page handlers
	handleRequests()

	if err = http.ListenAndServeTLS(listenAddress, certPath, privateKeyPath, nil); err != nil {
		logger.LogError(err)
		logger.Log(logger.Error, "Couldn't start HTTPS listener at", listenAddress)
		os.Exit(logger.ERR_GENERIC)
	}
}

func handleRequests() {
	// TODO: Add request logging
	http.HandleFunc("/", HandleRootPage)
	http.Handle("/dashboard", IsAuthorized(HandleDashboardPage))
	http.HandleFunc("/login", HandleLoginPage)
	http.HandleFunc("/register", HandleRegisterPage)
	http.HandleFunc("/Agents/agents.html", HandleAgentsPage)
	http.HandleFunc("/Launchers/launchers.html", HandleLaunchersPage)
	http.HandleFunc("/Listeners/listeners.html", HandleListenersPage)
}
