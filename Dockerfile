GO_VERSION=1.26.0

FROM golang:$GO_VERSION AS build
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN CGO_ENABLED=0 go build -a -ldflags="-s -w" -o modbus2mqtt .

FROM scratch
COPY --from=build /app/modbus2mqtt /
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 1000
ENTRYPOINT ["/modbus2mqtt"]
