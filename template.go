package main

const JSON_TEMPLATE string = `{
    "timestamp": {{.Timestamp}},
    "gateway": {
        "name": "{{.GatewayName}}",
        "model": "{{.GatewayModel}}",
        "serial": "{{.GatewaySerial}}"
    },
    "device": {
        "name": "{{.DeviceName}}",
        "model": "{{.DeviceModel}}",
        "serial": "{{.DeviceSerial}}",
        "readings": [
{{$l := dec (len .Readings)}}{{range $i, $reading := .Readings}}            {"param": "{{$reading.ParamName}}", "value": "{{$reading.Value}}"}{{if lt $i $l}},{{end}}
{{end}}        ]
    }
}`
