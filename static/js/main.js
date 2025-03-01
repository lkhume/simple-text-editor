// Connect to the server
var socket = io();

// Get a reference to the textarea element
var editor = document.getElementById("editor");
var timeout = null;

// Listen for changes in the textarea
editor.addEventListener("input", function () {
  clearTimeout(timeout);
  timeout = setTimeout(function () {
    socket.emit("text_update", { text: editor.value }); // Delay of 300ms to batch updates
  }, 300);
});

// Listen for changes in the server
socket.on("update_text", function (data) {
  // Update the text area with the latest text from other clients
  editor.value = data.text;
});
