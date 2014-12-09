DOCKER_HOST?=tcp://0.0.0.0:2376
DOCKER_TLS_VERIFY?=1
DOCKER_CERT_PATH?=$(PWD)/certs/

start-mosquitto:
	docker build -t mosquitto mosquitto
	docker run -it -p 60000:60000 mosquitto
.PHONY: start-mosquitto
