FROM golang:onbuild
MAINTAINER Oleg Morozenkov <a@reo7sp.ru>

export GIN_MODE=release
EXPOSE 8080
