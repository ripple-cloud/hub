## Ripple Hub

Ripple Hub is the interface for local apps to connect to the Ripple Cloud.

It will allow local apps to:
* Listen to requests from the cloud
* Send data back to cloud

### How Hub works?

* It will connect to an upstream (in this case MQTT broker of Ripple Cloud).
* It will listen to downstream requests on a TCP port.
* Apps can register with Hub and notify the port it's listening on.
* When there's a request for an app from upstream, Hub will forward it to the app's port.
* When an app publish data Hub will send it to upstream.

### Development

```
$ make start-mosquitto
$ go install
$ hub
```
