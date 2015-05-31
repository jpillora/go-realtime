package main

import (
	"net/http"
	"time"

	"github.com/jpillora/go-realtime"
)

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
	rt := realtime.Sync(foo)

	go func() {
		i := 0
		for {
			//change foo
			foo.A++
			foo.B--
			i++
			if i > 100 {
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
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(":3000", nil)
}
