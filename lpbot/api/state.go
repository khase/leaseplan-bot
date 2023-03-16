package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/khase/leaseplan-bot/lpbot/lpcon"
)

func getState(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.Marshal(lpcon.GetStates())
	if err == nil {
		w.Write(bytes)
	} else {
		io.WriteString(w, "Something went wrong :(")
	}

}
