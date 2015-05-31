//go:generate go-bindata -pkg realtime -o realtime-embed.go realtime.js

package realtime

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

var proto = "v1"

type Config struct {
	Throttle time.Duration
}

func New(c Config) *Realtime {
	if c.Throttle == 0 {
		c.Throttle = 100 * time.Millisecond
	}
	r := &Realtime{config: c}
	r.ws = websocket.Handler(r.serveWS)
	r.objs = map[key]*object{}
	r.users = map[string]*user{}
	//continually batches and sends updates
	go r.flusher()
	return r
}

func NewSync(val interface{}, c Config) *Realtime {
	r := New(c)
	r.Sync("default", val)
	return r
}

func Sync(val interface{}) *Realtime {
	return NewSync(val, Config{})
}

type Realtime struct {
	config Config
	ws     http.Handler
	objs   map[key]*object
	users  map[string]*user
}

func (r *Realtime) flusher() {
	for {
		//calculate all updates for all users
		for _, o := range r.objs {
			if !o.checked {
				o.check()
			}
		}
		//send all calculated updates
		for _, u := range r.users {
			if len(u.pending) > 0 {
				b, _ := json.Marshal(u.pending)
				u.conn.Write(b)
				u.pending = nil
			}
		}
		//sleeeepp
		time.Sleep(r.config.Throttle)
	}
}

func (r *Realtime) Sync(k string, val interface{}) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	r.objs[key(k)] = &object{
		key:         key(k),
		value:       val,
		bytes:       b,
		version:     1,
		subscribers: map[string]*user{},
		checked:     false,
	}
	return nil
}

func (r *Realtime) Update() {
	for _, obj := range r.objs {
		obj.checked = false
	}
}

func (r *Realtime) UpdateVal(k string) {
	if obj, ok := r.objs[key(k)]; ok {
		obj.checked = false
	}
}

func (r *Realtime) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := req.Header.Get("Sec-Websocket-Protocol")
	if p == "rt-"+proto {
		r.ws.ServeHTTP(w, req)
	} else if strings.HasSuffix(req.URL.Path, ".js") {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "text/javascript")
		w.Write(JS)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request"))
	}
}

func (r *Realtime) serveWS(conn *websocket.Conn) {
	vs := versions{}
	//only 1 decode
	if err := json.NewDecoder(conn).Decode(&vs); err != nil {
		log.Printf("invalid versions obj: %s", err)
		return
	}
	//ready
	u := &user{
		id:       conn.Request().RemoteAddr,
		conn:     conn,
		versions: vs,
		pending:  []*update{},
	}
	//add and subscribe to each obj
	r.users[u.id] = u
	for k, _ := range vs {
		obj, ok := r.objs[k]
		if !ok {
			conn.Write([]byte("missing object: " + k))
			return
		}
		obj.subscribers[u.id] = u
		u.pending = append(u.pending, &update{
			Key: k, Version: obj.version, Data: obj.bytes,
		})
	}
	//pipe to null
	io.Copy(ioutil.Discard, conn)
	//remove and unsubscribe to each obj
	delete(r.users, u.id)
	for k, _ := range vs {
		obj := r.objs[k]
		delete(obj.subscribers, u.id)
	}
	//disconnected
}

//embedded JS file
var JS = byteServe(_realtimeJs)

type byteServe []byte

func (b byteServe) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "text/javascript")
	w.Write([]byte(b))
}
