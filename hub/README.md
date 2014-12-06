## Ripple Hub

Ripple Hub is the interface for local apps to connect to the Ripple Server.

It will allow local apps to:
* Receive data from the server
* Send data back to server

### Workflow

* Listen on a port for client requests
* Start a MQTT client
* Clients can notify their availability (app-id, port)
* When there's a message for a specific client, forward it to the port 
* When a client publish data send it to MQTT server (app-id, data)

### Security

* (TODO) Listen only using Unix Domain Sockets.
* (TODO) Has whitelist based app priviledges policy.

```
  listen:
    app-A 
    app-B
  send:
    app-C
    app-D
  send-listen:
    app-E
```
