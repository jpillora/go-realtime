//go:generate go-bindata -pkg realtime -o realtime-embed.go realtime.js

package realtime

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
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
	r.userChanges = make(chan User)
	//continually batches and sends updates
	go r.flusher()
	return r
}

func NewSync(val interface{}, c Config) (*Realtime, error) {
	r := New(c)
	if err := r.Sync("default", val); err != nil {
		return nil, err
	}
	return r, nil
}

func Sync(val interface{}) (*Realtime, error) {
	return NewSync(val, Config{})
}

type Realtime struct {
	config Config
	ws     http.Handler
	mut    sync.Mutex //protects objects and users
	objs   map[key]*object
	users  map[string]*user

	watchingChanges bool
	userChanges     chan User
}

type User struct {
	Connected bool
	Address   string
}

func (r *Realtime) Changes() <-chan User {
	if r.watchingChanges {
		panic("Already watching user changes")
	}
	r.watchingChanges = true
	return r.userChanges
}

func (r *Realtime) flusher() {
	for {
		r.mut.Lock()
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
		r.mut.Unlock()
		//sleeeepp
		time.Sleep(r.config.Throttle)
	}
}

func (r *Realtime) Sync(k string, val interface{}) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	r.mut.Lock()
	r.objs[key(k)] = &object{
		key:         key(k),
		value:       val,
		bytes:       b,
		version:     1,
		subscribers: map[string]*user{},
		checked:     false,
	}
	r.mut.Unlock()
	return nil
}

func (r *Realtime) Update() {
	r.mut.Lock()
	for _, obj := range r.objs {
		obj.checked = false
	}
	r.mut.Unlock()
}

func (r *Realtime) UpdateVal(k string) {
	r.mut.Lock()
	if obj, ok := r.objs[key(k)]; ok {
		obj.checked = false
	}
	r.mut.Unlock()
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
		if err != io.EOF {
			log.Printf("invalid versions obj: %s", err)
		}
		return
	}
	//ready
	u := &user{
		id:       conn.Request().RemoteAddr,
		uptime:   time.Now(),
		conn:     conn,
		versions: vs,
		pending:  []*update{},
	}

	//check objs
	r.mut.Lock()
	for k, _ := range vs {
		if _, ok := r.objs[k]; !ok {
			conn.Write([]byte("missing object: " + k))
			r.mut.Unlock()
			return
		}
	}
	r.mut.Unlock()

	//add and subscribe to each obj
	r.mut.Lock()
	r.users[u.id] = u
	if r.watchingChanges {
		r.userChanges <- User{Connected: true, Address: u.id}
	}
	for k, _ := range vs {
		obj := r.objs[k]
		obj.subscribers[u.id] = u
		u.pending = append(u.pending, &update{
			Key: k, Version: obj.version, Data: obj.bytes,
		})
	}
	r.mut.Unlock()
	//block here during connection - pipe to null
	io.Copy(ioutil.Discard, conn)
	//remove and unsubscribe to each obj
	r.mut.Lock()
	delete(r.users, u.id)
	if r.watchingChanges {
		r.userChanges <- User{Connected: false, Address: u.id}
	}
	for k, _ := range vs {
		obj := r.objs[k]
		delete(obj.subscribers, u.id)
	}
	r.mut.Unlock()
	//disconnected
}

//embedded JS file
var JS = jsServe(_realtimeJs)

type jsServe []byte

func (j jsServe) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	b := []byte(j)
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "text/javascript")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.Write(b)
}
