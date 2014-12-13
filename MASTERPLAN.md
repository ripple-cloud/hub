# Ripple

Easy way to write and distribute Raspberry PI apps.

## Why use Ripple?

* Write your apps in any language.
* Port existing apps with minimal modifications.
* You can control, fetch data using a simple REST API.
* You will not need to assign a static IP address to bring your PI to the internet.
* You won't lose requests or data if your network or power fails.

## Key Components

* Cloud server (which exposes a REST API)
* Client running on Raspberry PI (which will talk with your apps)

## How it Works?

|----------|           |----------------|           |---------------|            |-------------|
|          |           |                |           |               |            |             |
| Your App |  <----->  |     Hub        |  <------> |    Cloud      |  <------>  | REST Client |
|          |           |                |           |               |            |             |
|----------|           |----------------|           |---------------|            |-------------|

You can write apps using any language that can run on a Raspberry Pi. these apps can:
* Listen for requests and respond
* Send periodic datapoints. (eg: sensor sending data)

* Each app must register first with the cloud server. (you have to assign a unique slug to your app)
* Then at the time of initialization, apps must register with Ripple client
  - can specify a port it's listening for requests (if no port is provided, it can only send msgs to server)
* Cloud server exposes an endpoint for each app, which can accept a payload through HTTP POST. (each request will have a unique job ID)
* When an app is initialized, it can receive these messages from server (including ones that were received before the client was initialized)
* After processing the request or when data is available, apps can send results back to server.
* Server will store these results (for a limited period) and can optionally call a webhook endpoint specified by the user.

## Your App

* Talks to Ripple Hub, through TCP.
* Initialize - specify the port app it is listening
* Listen on that port for messages
* Send data back to client.

pseudo-code for an app:
```
ripple.connect("app-slug", { port: 4567 })
server.listen(4567, function(data) {
  // message received
  // process message
  job := json.parse(data)

  // do some internal action

  // send a reply
  ripple.send(json.stringify({job: 30}))
})
```

## REST API

* POST /app/{app-slug}  - send a new request to app (responds with 202 job-id )
* GET /app/{app-slug}   - lists results of all jobs of the app
* GET /app/{app-slug}/job/{job-id}  - list results for the given job

## server

* Account creation
* Manage devices (a Pi that have installed the client)
* Manage tokens

## Built-in apps (optional)

* Access GPIO pins over REST API
  post /app/gpio
  request:
  {
    digitalwrite: 'd7', '0'
  }
* Play audio

## How to use?

* Signup for an account in cloud server
* Download and install Ripple Hub in a Raspberry Pi
* Write an app conforming to the spec.
* Call it and access data using the REST interface.

## Security

* server will issue authentication tokens you can use in clients
* when a Pi is lost or become stale you can revoke these tokens
* Apps will be scoped by client tokens

## TODO

* Ripple Cloud
* Ripple Hub
* Ripple Hub Client library (that talks to Ripple Hub over TCP)
* REST endpoints to Ripple Cloud
* Built-in apps

## Implementation

* client app
  - mqtt client
  - use leveldb as the store
  - dispatch mqtt msgs to appropriate services
  - listens on a port for an incoming requests forward them as mqtt msgs
* server
  - mosquitto server
  - http service listening for api requests - stores them in a db and forwards as mqtt msgs
  - listens to incoming mqtt messages - stores them in a db or calls given webhooks
  - handle signup/login process

## Distribution

* server is packaged as a docker image.
* client will be a binary that can be downloaded and executed.
  - in future, build sd card images containing a full os with the client and built-in apps

### Security

* Make Hub Listen only using Unix Domain Sockets.
* Whitelist based app priviledges policy.

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
