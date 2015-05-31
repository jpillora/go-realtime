
:warning: This project is currently in progress.

---

<!-- 
https://github.com/evanphx/json-patch
https://github.com/Starcounter-Jack/JSON-Patch
 -->

# go-realtime

Keep your Go structs in sync with your JS objects

### Usage

Server

``` go
rt := realtime.New(realtime.Config{
	Throttle: 100 * time.Millisecond, //limit update
	Delta: true, //generate diffs with jsonpatch
})

//any json.Marshaller
foo := struct{}

rt.Object("foo", foo)

//make changes
foo.A = 42

//push to client
rt.Update()
```

Client

``` js
var rt = realtime();

var foo = rt.object("foo")

foo.$on("update", function() {
	//do something on update...
});
```

### Todo

* Allow clients to write