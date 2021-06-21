package web

import (
	"net/http"
	"product_code/check_stream/config"

	log4plus "common/log4go"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

type ResponseCommon struct {
	Result int    `json:"result"`
	Msg    string `json:"msg"`
}

func handlerWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers",
				"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}

		h.ServeHTTP(w, r)
	})
}

func StartWeb() {
	go startWeb()
}

func startWeb() {
	m := mux.NewRouter()
	m.Handle("/checkstream/add_task", handlerWrap(http.HandlerFunc(HttpAddTask)))
	m.Handle("/checkstream/stream_status", handlerWrap(http.HandlerFunc(HttpStreamStatus)))
	n := negroni.Classic()
	n.UseHandler(m)
	log4plus.Debug("startWeb listen=%s", config.GetInstance().Listen)
	n.Run(config.GetInstance().Listen)
}
