FROM golang:1.8-onbuild
MAINTAINER Oleg Morozenkov <a@reo7sp.ru>

ENV GIN_MODE release
EXPOSE 8080
