FROM golang AS build
WORKDIR /go/src/alinea
COPY ./app/ . 
RUN go get -v ./...
RUN go build -v -o app

FROM debian
WORKDIR /app
COPY --from=build /go/src/alinea/app .
COPY templates templates/
ENTRYPOINT ["./app"]