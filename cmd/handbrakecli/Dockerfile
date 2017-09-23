FROM ubuntu:16.04

ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update
RUN apt-get install -y software-properties-common
RUN add-apt-repository ppa:stebbins/handbrake-releases
RUN apt-get update
RUN apt-get install -y handbrake-cli=1.0.4-zhb-1ppa1~xenial1

RUN mkdir -p /config/ghb/
COPY presets.json /config/ghb/presets.json
ENTRYPOINT ["HandBrakeCLI"]
