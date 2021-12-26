function formatJSON(type, data) {
  return JSON.stringify({ type: type, data: data });
}

class DBSocket {
  viewPort;
  #webSocket;

  static get NORMAL_CLOSURE() { return 1000; }

  constructor() {}

  connectViewPort(viewPort) {
    this.viewPort = viewPort;
  }

  sendJSON(json) {
    this.#webSocket.send(json);
  }

  connect() {
    var that = this;
    const websocketURI = "ws://" + location.host + "/messages";
    this.#webSocket = new WebSocket(websocketURI);

    this.#webSocket.onopen = function () {
      that.viewPort.readyViewPort();
      console.log("Connected WebSocket to " + websocketURI);
    }

    this.#webSocket.onclose = function (e) {
      that.viewPort.resetViewPort();
      console.log("WebSocket closed with (" + e.code + ")");
    }

    this.#webSocket.onerror = function (e) {
      that.viewPort.resetViewPort();
      console.error("WebSocket received an error: " + e);
    }

    this.#webSocket.onmessage = function (e) {
      const data = JSON.parse(e.data);
      switch (data.type) {
        case "user_change": 
          that.viewPort.setConnected(data.data);
          break;
        case "message_change":
          that.viewPort.setOutput(JSON.parse(data.data))
          break;
      }
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

  textBuffer;
  textSendButton;
  textOutput;

  constructor(dbSocket) {
    var that = this;
    dbSocket.connectViewPort(this);
    this.dbSocket = dbSocket;
    this.display          = document.getElementById("display");
    this.connectButton    = document.getElementById("connect");
    this.disconnectButton = document.getElementById("disconnect");
    this.textBuffer       = document.getElementById("buffer");
    this.textSendButton   = document.getElementById("send");
    this.textOutput       = document.getElementById("output");

    this.connectButton.onclick = function () {
      that.dbSocket.connect();
    }

    this.disconnectButton.onclick = function () {
      that.dbSocket.disconnect();
    }

    this.textSendButton.onclick = function () {
      that.dbSocket.sendJSON(formatJSON("send_text", that.textBuffer.value));
    }

    this.textBuffer.addEventListener("keyup", function (event) {
      event.preventDefault();
      if (event.key == "Enter") {
        that.dbSocket.sendJSON(formatJSON("send_text", that.textBuffer.value));
      }
    });
  }

  readyViewPort() {
    this.connectButton.disabled = !this.connectButton.disabled;
    this.disconnectButton.disabled = !this.disconnectButton.disabled;
    this.textBuffer.disabled = false;
    this.textSendButton.disabled = false;
  }

  resetViewPort() {
    this.connectButton.disabled = false;
    this.disconnectButton.disabled = true;
    this.display.innerText = "N/A";
    this.textBuffer.disabled = true;
    this.textSendButton.disabled = true;
  }

  setConnected(count) {
    this.display.innerText = count;
  }

  setOutput(outputs) {
    this.textOutput.innerHTML = "";
    for (const output of outputs) {
      let entry = document.createElement("div");
      entry.append(output);
      this.textOutput.append(entry);
    }
  }
}

var dbSocket;
var viewPort;

window.onload = function () {
  dbSocket = new DBSocket();
  viewPort = new ViewPort(dbSocket);
}