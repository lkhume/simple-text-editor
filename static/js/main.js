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

let prevText = editor.value;

// Send updates to the server with a delay
editor.addEventListener("input", () => {
  clearTimeout(timeout);
  timeout = setTimeout(() => {
    const currText = editor.value;
    if (currText.length > prevText.length) {
      // Insertion occurred
      let pos = editor.selectionStart - (currText.length - prevText.length);
      let insertedText = currText.slice(
        pos,
        pos + (currText.length - prevText.length),
      );

      if (insertedText.length === 1) {
        const op = { type: "insert", pos: pos, char: insertedText };
        socket.send(JSON.stringify(op));
      }
    } else if (currText.length < prevText.length) {
      // Deletion occurred
      let pos = editor.selectionStart;
      const op = { type: "delete", pos: pos };
      socket.send(JSON.stringify(op));
    }

    prevText = currText;
  }, 300);
});
