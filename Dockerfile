FROM golang:1.25

WORKDIR /app

# RUN go install github.com/air-verse/air@latest
RUN curl -fsSL \
    https://raw.githubusercontent.com/pressly/goose/master/install.sh |\
    sh
# RUN apt-get -y update && apt-get install vim -y

COPY . .
COPY vendor/ ./vendor/

RUN go env -w GOFLAGS="-mod=vendor"
RUN go build -o user_api

CMD ["/app/user_api"]

# CMD air
