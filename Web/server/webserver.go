package main

import (
	"log"
	"net/http"

	"github.com/PatronC2/Patron/data"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/qkgo/yin"
	"github.com/s-christian/gollehs/lib/logger"
)

func main() {
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

	http.ListenAndServe(":3001", r)
}
