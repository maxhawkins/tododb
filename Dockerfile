FROM golang:1.7

RUN mkdir -p /go/src/github.com/maxhawkins/tododb
WORKDIR /go/src/github.com/maxhawkins/tododb

COPY . /go/src/github.com/maxhawkins/tododb
RUN go get -u github.com/maxhawkins/tododb
RUN go install github.com/maxhawkins/tododb

CMD ["tododb"]
