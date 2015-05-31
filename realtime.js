/* global window,WebSocket */
// Realtime - v0.1.0 - https://github.com/jpillora/go-realtime
// Jaime Pillora <dev@jpillora.com> - MIT Copyright 2015
(function(window) {

  var proto = "v1";

  //special merge - ignore $properties
  // x <- y
  var merge = function(x, y) {
    if (!x || typeof x !== "object" ||
      !y || typeof y !== "object")
      return y;
    var k;
    if (x instanceof Array && y instanceof Array)
      while (x.length > y.length)
        x.pop();
    else
      for (k in x)
        if (k[0] !== "$" && !(k in y))
          delete x[k];
    for (k in y)
      x[k] = merge(x[k], y[k]);
    return x;
  };

  var events = ["message","error","open","close"];

  //realtime class - represents a single websocket
  function Realtime(url) {
    if(!url)
      url = window.location.href;
    if(!/^http(s?:\/\/.+)$/.test(url))
      throw "Invalid URL";
    this.url = "ws" + RegExp.$1;
    this.connect();
    this.objs = {};
  }
  Realtime.prototype = {
    sync: function(key, obj) {
      this.objs[key] = obj;
    },
    connect: function() {
      if(this.ws)
        this.disconnect();
      if(!this.delay)
        this.delay = 100;
      this.ws = new WebSocket(this.url, "rt-"+proto);
      var _this = this;
      events.forEach(function(e) {
        e = "on"+e;
        _this.ws[e] = _this[e].bind(_this);
      });
      this.ping.t = setInterval(this.ping.bind(this), 30 * 1000);
    },
    disconnect: function() {
      if(!this.ws)
        return;
      var _this = this;
      events.forEach(function(e) {
        _this.ws["on"+e] = null;
      });
      if(this.ws.readyState !== WebSocket.CLOSED)
        this.ws.close();
      this.ws = null;
      clearInterval(this.ping.t);
    },
    ping: function() {
      this.send("ping");
    },
    send: function(data) {
      if(this.ws.readyState === WebSocket.OPEN)
        return this.ws.send(data);
    },
    onmessage: function(event) {
      var str = event.data;
      if (str === "ping") return;
      var msg = JSON.parse(str);

      if(!msg.data)
        return console.warn("missing msg data");

      for(var key in msg.data) {
        var dst = this.objs[key];
        var src = msg.data[key];
        if(!src || !dst)
          continue;
        merge(dst, src);
        if(typeof dst.$apply === "function")
          dst.$apply();
      }
    },
    onopen: function() {
      this.delay = 100;
      console.log("connected");
    },
    onclose: function() {
      this.delay *= 2;
      console.log("disconnected, reconnecting in %sms", this.delay);
      setTimeout(this.connect.bind(this), this.delay);
    },
    onerror: function(err) {
      console.error("websocket error: %s", err);
    }
  };

  //publicise
  window.realtime = function(url) {
    return new Realtime(url);
  };
}(window, undefined));