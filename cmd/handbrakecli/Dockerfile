FROM ubuntu:16.04

ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update
RUN apt-get install -y software-properties-common
RUN add-apt-repository ppa:stebbins/handbrake-releases
RUN apt-get update
RUN apt-get install -y handbrake-cli

RUN mkdir -p /config/ghb/
COPY presets.json /config/ghb/presets.json
ENTRYPOINT ["HandBrakeCLI"]
