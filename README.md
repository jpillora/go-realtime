
# go-realtime

Keep your Go structs in sync with your JS objects

:warning: This project is very beta.

### Features

* Simple API
* Delta updates using JSONPatch

### Usage

Server

``` go
type Foo struct {
	realtime.Object
	A, B int
}
foo := &Foo{}

//create handler and add foo to it
rt := realtime.NewHandler()
rt.Add("foo", foo)
http.Handle("/realtime", rt)

//...later...

//make changes
foo.A = 42
//push to client
foo.Update()
```

Client

``` js
var foo = {};

realtime.add("foo", foo, function onupdate() {
	//do stuff with foo...
});
```

See [example](example/)

#### MIT License

Copyright © 2015 Jaime Pillora &lt;dev@jpillora.com&gt;

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
'Software'), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.