from flask import Flask, render_template
from flask_socketio import SocketIO, emit

app = Flask(__name__)
app.config['SECRET_KEY'] = 'secret!'
socketio = SocketIO(app)

shared_text = ""


@app.route('/')
def index():
    # Render the index.html template with the current text
    return render_template('index.html', text=shared_text)


@socketio.on('connect')
def on_connect():
    # When a client connects, send them the current text
    emit('update_text', {'text': shared_text})


@socketio.on('text_update')
def on_text_update(data):
    global shared_text
    shared_text = data.get('text', "")
    # Broadcast the updated text to all connected clients (except the sender)
    emit('update_text', {'text': shared_text}, broadcast=True, include_self=False)


if __name__ == '__main__':
    # Use eventlet to support WebSocket connections
    socketio.run(app, debug=True)
