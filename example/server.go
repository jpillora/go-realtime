package main

import (
	"log"
	"net/http"
	"time"

	"github.com/jpillora/go-realtime"
)

var indexhtml = []byte(`
<pre id="out"></pre>
<script src="realtime.js"></script>
<script>
var rt = realtime("/realtime");
var foo = {};
foo.$onupdate = function() {
	out.innerHTML = JSON.stringify(foo, null, 2);
};
//keep in sync with the server
rt.sync(foo);
</script>
`)

type Foo struct {
	A, B int
	C    []int
	D    string
	E    Bar
}

type Bar struct {
	X, Y int
}

func main() {

	var foo = &Foo{A: 21, B: 42, D: "0"}

	//publish foo
	rt, _ := realtime.Sync(foo)

	go func() {
		i := 0
		for {
			//change foo
			foo.A++
			foo.B--
			i++
			if i > 10 {
				foo.C = foo.C[1:]
			}
			foo.C = append(foo.C, i)
			//mark updated
			rt.Update()
			//do other stuff...
			time.Sleep(100 * time.Millisecond)
		}
	}()

	http.Handle("/realtime", rt)
	http.Handle("/realtime.js", realtime.JS)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexhtml)
	})
	log.Printf("Listening on localhost:4000...")
	http.ListenAndServe(":4000", nil)
}
