package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/khase/leaseplan-bot/lpbot/lpcon"
)

type HealthData struct {
	Startup  string   `json:"Startup,omitempty"`
	Watchers []string `json:"Watchers,omitempty"`
}

func getHealth(w http.ResponseWriter, r *http.Request) {

	HealthData := &HealthData{
		Startup:  startTime.UTC().String(),
		Watchers: lpcon.GetWatcherKeys(),
	}

	bytes, err := json.Marshal(HealthData)
	if err == nil {
		w.Write(bytes)
	} else {
		io.WriteString(w, "Something went wrong :(")
	}

}
