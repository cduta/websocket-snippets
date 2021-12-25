class DBSocket {
  viewPort;
  #webSocket;

  static get NORMAL_CLOSURE() { return 1000; }

  constructor() {}

  connectViewPort(viewPort) {
    this.viewPort = viewPort;
  }

  connect() {
    var that = this;
    const websocketURI = "ws://" + location.host + "/connect";
    this.#webSocket = new WebSocket(websocketURI);

    this.#webSocket.onopen = function () {
      console.log("Connected WebSocket to " + websocketURI);
      that.viewPort.connect_button.disabled    = true;
      that.viewPort.disconnect_button.disabled = false;
    }

    this.#webSocket.onclose = function (e) {
      console.log("WebSocket closed with (" + e.code + ")");
      that.viewPort.connect_button.disabled    = false;
      that.viewPort.disconnect_button.disabled = true;
    }

    this.#webSocket.onerror = function (e) {
      console.error("WebSocket received an error: " + e);
    }
  }

  disconnect() {
    this.#webSocket.close(DBSocket.NORMAL_CLOSURE);
  }
}

class ViewPort {
  dbSocket;
  connect_button;
  disconnect_button;

  constructor(dbSocket) {
    var that = this;
    dbSocket.connectViewPort(this);
    this.dbSocket = dbSocket;
    this.connect_button    = document.getElementById("connect");
    this.disconnect_button = document.getElementById("disconnect");

    this.connect_button.onclick = function () {
      that.dbSocket.connect();
    }

    this.disconnect_button.onclick = function () {
      that.dbSocket.disconnect();
    }
  }
}

var dbSocket;
var viewPort;

window.onload = function () {
  dbSocket = new DBSocket();
  viewPort = new ViewPort(dbSocket);
}