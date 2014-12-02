FROM ubuntu:14.04

RUN apt-get update \
    && apt-get install -y build-essential wget libwrap0-dev libssl-dev python-distutils-extra libc-ares-dev\
    && rm -rf /var/lib/apt/lists/*

RUN locale-gen en_US.utf8

RUN mkdir -p /usr/local/src
WORKDIR /usr/local/src
RUN wget http://mosquitto.org/files/source/mosquitto-1.3.5.tar.gz
RUN tar xvzf ./mosquitto-1.3.5.tar.gz
WORKDIR /usr/local/src/mosquitto-1.3.5
RUN make
RUN make install
RUN adduser --system --disabled-password --disabled-login mosquitto

EXPOSE 60000
CMD /usr/local/sbin/mosquitto -p 60000
