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
      that.viewPort.toggleConnectButton();
    }

    this.#webSocket.onclose = function (e) {
      console.log("WebSocket closed with (" + e.code + ")");
      that.viewPort.toggleConnectButton();
      that.viewPort.resetConnected();
    }

    this.#webSocket.onerror = function (e) {
      console.error("WebSocket received an error: " + e);
    }

    this.#webSocket.onmessage = function (e) {
      that.viewPort.setConnected(e.data);
    }
  }

  disconnect() {
    this.#webSocket.close(DBSocket.NORMAL_CLOSURE);
  }
}

class ViewPort {
  dbSocket;
  display;
  connectButton;
  disconnectButton;

  constructor(dbSocket) {
    var that = this;
    dbSocket.connectViewPort(this);
    this.dbSocket = dbSocket;
    this.display           = document.getElementById("display");
    this.connectButton    = document.getElementById("connect");
    this.disconnectButton = document.getElementById("disconnect");

    this.connectButton.onclick = function () {
      that.dbSocket.connect();
    }

    this.disconnectButton.onclick = function () {
      that.dbSocket.disconnect();
    }
  }

  toggleConnectButton() {
    this.connectButton.disabled    = !this.connectButton.disabled;
    this.disconnectButton.disabled = !this.disconnectButton.disabled;
  }

  setConnected(count) {
    this.display.innerText = count;
  }

  resetConnected() {
    this.display.innerText = "N/A";
  }
}

var dbSocket;
var viewPort;

window.onload = function () {
  dbSocket = new DBSocket();
  viewPort = new ViewPort(dbSocket);
}