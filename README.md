# Modbus-2-MQTT

This program allow you to read data from a Modbus device, format it based on a
template and then publish it to a MQTT broker.

## Command line arguments
* `-c=<filename>` - Path to a TOML configuration file. Defaults to `config.toml` if not provided.
* `-g` - Generate a sample TOML configuration file and outputs it to the console.
* `-f=<filename>` - Used in conjunction with `-g` to instead write the sample configuration file to a file.
* `-logginglevel=<LEVEL>` - Logging level (TRACE, DEBUG, INFO, WARN, ERROR, FATAL). Defaults to ERROR.
* `-h` - Show help information

## Example TOML configuration file format

```toml
[modbus]
url = "tcp://localhost:502"
unit_id = 1
timeout = 1000
poll_rate = 1000
registers = [
    {holding_register = 40001, size = "UINT32", multiplier = 0.01, format = "%.2f", param_name = "my_param_name_1"},
    {holding_register = 40003, size = "SINT32", multiplier = 0.01, param_name = "my_param_name_2"},
    {holding_register = 40005, size = "UINT64", param_name = "my_param_name_3"}
]

[mqtt]
url = "mqtt://localhost:1883"
qos = 0
client_id = "some_unique_client_id" # 23 characters max
connect_retry = 1000
username = "my_username"
password = "my_password"
pub_topic = "my/special/topic"
pub_rate = 1000

[template_data]
template_file = "/some/file/path/filename.go.tmpl"

[template_data.template_kv]
key_name_1 = "key_value_1"
key_name_2 = "key_value_2"
key_name_3 = "key_value_3"
```

NOTE: All times in the configuration file are in milliseconds

### Configuration file definitions

#### [modbus]
* `url` - URL defining the IP or hostname of the Modbus server along with the port number.
* `unit_id` - Modbus unit id. Defaults to 1 if not supplied.
* `timeout` - Response period timeout value in milliseconds.
* `poll_rate` - Rate at which the Modbus server should be queried for updated data in milliseconds.
* `registers` - Array of registers to query for data.

    * `holding_register` - Modbus holding register.
    * `size` - Register size to read (currently supports: UINT16, SINT16, UINT32, SINT32, UINT64, SINT64).
    * `multiplier` - Decimal multiplier to apply to data after it is read (i.e. register contains kW but you want W, set multiplier=0.001). Defaults to 1 if not supplied.
    * `format` - Formatting to apply to data using Go's string formatting codes (i.e. maximum of 2 decimals, set format="%.2f"). Defaults to "%.0f" if not supplied.
    * `param_name` - Name to use for this register data.

#### [mqtt]
* `url` - URL defining the IP or hostname of the MQTT broker along with the port number.
* `qos` - The QoS to use for the MQTT publish (0, 1, 2).
* `client_id` - The MQTT Client ID to use. If not provided, a random one will be generated.
* `connect_retry` - Amount of time to wait before trying to automatically reconnect to the MQTT broker if the connection is lost.
* `username` - If provided, the username to set for the MQTT connection and requires `password`.
* `password` - If provided, the password to set for the MQTT connection and requires `username`.
* `pub_topic` - Topic to use for publishing MQTT payload.
* `pub_rate` - Rate at which to publish the MQTT payload.

#### [template_data]
* `template_file` - Path to the template file to use for formatting the MQTT payload.

#### [template_data.template_kv]
Array of key/value pairs to use in the template. Each entry must have the format of:
```
key_name = "key_value"
```

## Example templates

### Example #1
```tmpl
{
    "timestamp": {{.Timestamp.RFC3339Nano}},
    "kv_info": {
        "key_name_1": "{{index .TemplateKV "key_name_1"}}",
        "key_name_2": "{{index .TemplateKV "key_name_2"}}",
        "key_name_3": "{{index .TemplateKV "key_name_3"}}"
    },
    "keyvalues": [
    ]
    "readings": [
{{$l := dec (len .RegisterValues)}}{{range $i, $register := .RegisterValues}}
        {"param": "{{$register.ParamName}}", "value": "{{$register.Value}}"}{{if lt $i $l}},{{end}}
{{end}}
    ]
    }
}
```

### Example #2
```tmpl
Timestamp={{Timestamp.Seconds}}
{{range $k, $v := .TemplateKV}}
Key="{{$k}}", Value="{{$v}}"
{{end}}
````

## Available fields in a template

### Timestamp

The Timestamp is the time at the beginning of the poll of the Modbus read. There are 4 options available in the template:

* `Seconds` - This is the EPOCH seconds (i.e. number of seconds since 1/1/1970 00:00:00.000Z)
* `Nanoseconds` - Nanoseconds offset within the seconds, in the range [0,999999999]
* `RFC3339` - RFC3339 formatted time (YYYY-MM-DDTHH:MI:SSZ) at UTC
* `RFC3339Nanos` - RFC3339 formatted time with nanoseconds (YYYY-MM-DDTHH:MI:SS.000000000Z) at UTC

Each of these fields are referenced like this:

```
{{.Timestamp.<<option>}}
```

### Key/value pairs

In the `[template_data.template_kv]` section of the configuration file, you can define any key/value pairs that need to
be passed along to the template. You can see samples of their use in the example templates above.

### Data read from Modbus registers

Data that was read from the Modbus registers is available as a list called `.RegisterValues` and has the following fields
available:

* `ParamName` - This is the same value that is in the `param_name` of the `[modbus]registers` section from the config.
* `Value` - String value representing the data read from the Modbus register.
