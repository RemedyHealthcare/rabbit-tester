FROM debian:jessie

RUN apt-get update && apt-get install -y curl git

RUN curl -O https://storage.googleapis.com/golang/go1.7.1.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.7.1.linux-amd64.tar.gz

EXPOSE 8080
ENV PORT 8080

ENV GOPATH /opt/go
ADD . /opt/go/src/github.com/darksigma/mqtest
WORKDIR /opt/go/src/github.com/darksigma/mqtest
RUN /usr/local/go/bin/go get
RUN /usr/local/go/bin/go install

CMD /opt/go/bin/mqtest
