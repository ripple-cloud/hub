# Ripple

Easy way to talk with your Raspberry PI using the cloud.

## Key Components

* Cloud Server
* Client running on RaspberryPI

## How this works?

You can write apps using any language that can run on a Raspberry PI. These apps can:
* Listen for requests and respond
* Send periodic datapoints.

* Each app must with register first with the cloud server.
* This will give the app an UUID.
* Then at the time of initialization, apps must register with Ripple client
  - can specify a port it's listening for requests (if no port is provided, it can only send msgs to server)
* Cloud server exposes an endpoint for each app, which can accept a payload through HTTP POST. (each request will have a unique ID)
* When an app is initialized, it can receive these messages from server (this includes ones that were received before the client was initialized)
* After processing the request or when data is available, apps can send results back to server.
* Server will store these results under the app ID (for a limited period) and can optionally call a webhook endpoint speicified by the user.
