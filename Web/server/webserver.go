package main

import (
	"log"
	"net/http"

	"github.com/PatronC2/Patron/data"
	"github.com/go-chi/chi"
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

	http.ListenAndServe(":3000", r)
}
