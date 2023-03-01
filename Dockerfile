FROM golang:1.20.1 AS build
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o modbus2mqtt .

FROM scratch
COPY --from=build /app/modbus2mqtt /
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 1000
ENTRYPOINT ["/modbus2mqtt"]
