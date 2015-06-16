package realtime

import (
	"time"

	"golang.org/x/net/websocket"
)

type versions map[key]int64

type update struct {
	Key     key
	Delta   bool
	Version int64
	Data    jsonBytes
}

type user struct {
	id       string
	uptime   time.Time
	conn     *websocket.Conn
	versions versions
	pending  []*update
}
