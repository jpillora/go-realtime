//go:generate go-bindata -pkg realtime -o realtime-embed.go realtime.js

package realtime

import (
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

var proto = "v1"

//embedded JS file
var JS = _realtimeJs

type Config struct {
	Throttle time.Duration
	Delta    bool
}

func New(c Config) *Realtime {
	r := &Realtime{config: c}
	r.ws = websocket.Handler(r.serveWS)
	return r
}

type Realtime struct {
	config Config
	ws     http.Handler
}

func (r *Realtime) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := req.Header.Get("Sec-Websocket-Protocol")
	if p == "rt-"+proto {
		r.ws.ServeHTTP(w, req)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (r *Realtime) serveWS(conn *websocket.Conn) {
	conn.Write([]byte("ping"))
	select {}
}
