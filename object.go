package realtime

import (
	"encoding/json"

	"github.com/mattbaird/jsonpatch"
)

type key string

type object struct {
	key         key
	value       interface{}
	bytes       []byte
	version     int64
	subscribers map[string]*user
	checked     bool
}

func (o *object) check() {
	if o.checked {
		return
	}
	newBytes, _ := json.Marshal(o.value)
	//calculate change set
	ops, _ := jsonpatch.CreatePatch(o.bytes, newBytes)
	if len(ops) == 0 {
		return
	}
	delta, _ := json.Marshal(ops)
	o.bytes = newBytes
	prev := o.version
	o.version++
	for _, u := range o.subscribers {
		update := &update{
			Key:     o.key,
			Version: o.version,
		}
		//calc update - send the smallest
		if u.versions[o.key] == prev && len(delta) < len(o.bytes) {
			update.Delta = true
			update.Data = delta
		} else {
			update.Delta = false
			update.Data = o.bytes
		}
		//insert pending update
		u.pending[o.key] = update
		//user now has this version
		u.versions[o.key] = o.version
	}
	//mark
	o.checked = true
}
