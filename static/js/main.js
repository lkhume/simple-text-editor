// Connect to the WebSocket endpoint
const protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
const socket = new WebSocket(protocol + window.location.host + "/ws");

// Reference to the textarea
const editor = document.getElementById("editor");
let timeout = null;

// On connection open.
socket.onopen = () => {
  console.log("WebSocket connection established.");
};

// Handle incoming messages.
socket.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.text !== undefined) {
    editor.value = data.text;
  }
};

// Send updates to the server with a delay
editor.addEventListener("input", () => {
  clearTimeout(timeout);
  timeout = setTimeout(() => {
    const msg = { text: editor.value };
    socket.send(JSON.stringify(msg));
  }, 300);
});
