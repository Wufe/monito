FROM ubuntu:18.04

COPY ./release /app

RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

RUN apt-get update && \
	apt-get install --yes libicu60 libssl1.1

WORKDIR /app