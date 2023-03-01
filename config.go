package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type RegisterConfig struct {
	HoldingRegister int     `toml:"holding_register"`
	Size            string  `toml:"size"`
	Multiplier      float32 `toml:"multiplier"`
	Format          string  `toml:"format"`
	ParamName       string  `toml:"param_name"`
}

type ModbusInfo struct {
	URL            string           `toml:"url"`
	UnitId         int64            `toml:"unit_id"`
	Timeout        int64            `toml:"timeout"`
	PollRate       int64            `toml:"poll_rate"`
	ReconnectPause int64            `toml:"reconnect_pause"`
	Registers      []RegisterConfig `toml:"registers"`
}

type MqttInfo struct {
	URL          string `toml:"url"`
	Qos          int8   `toml:"qos"`
	ClientId     string `toml:"client_id"`
	ConnectRetry int64  `toml:"connect_retry"`
	Username     string `toml:"username"`
	Password     string `toml:"password"`
	PubTopic     string `toml:"pub_topic"`
	PubRate      int64  `toml:"pub_rate"`
}

type RegisterValue struct {
	ParamName string
	Value     string
}

type TimestampInfo struct {
	Seconds     int64
	Nanoseconds int32
	RFC3339     string
	RFC3339Nano string
}

type TemplateInfo struct {
	TemplateFile   string            `toml:"template_file"`
	TemplateKV     map[string]string `toml:"template_kv"`
	Timestamp      TimestampInfo
	RegisterValues []RegisterValue
}

type tomlConfig struct {
	Modbus   ModbusInfo   `toml:"modbus"`
	Mqtt     MqttInfo     `toml:"mqtt"`
	Template TemplateInfo `toml:"template_data"`
}

func loadConfig(filePath string) (tomlConfig, error) {
	var config tomlConfig
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return config, err
	}

	for i := 0; i < len(config.Modbus.Registers); i++ {
		if config.Modbus.Registers[i].Multiplier == 0.0 {
			config.Modbus.Registers[i].Multiplier = 1.0
		}
		if config.Modbus.Registers[i].Format == "" {
			config.Modbus.Registers[i].Format = "%.0f"
		}
	}

	return config, nil
}

func dumpConfig(config tomlConfig) {
	fmt.Printf("[modbus]\n")
	fmt.Printf("url = \"%s\"\n", config.Modbus.URL)
	fmt.Printf("unit_id = %d\n", config.Modbus.UnitId)
	fmt.Printf("timeout = %d\n", config.Modbus.Timeout)
	fmt.Printf("poll_rate = %d\n", config.Modbus.PollRate)
	fmt.Printf("reconnect_pause = %d\n", config.Modbus.ReconnectPause)
	fmt.Printf("registers = [\n")
	for i, reading := range config.Modbus.Registers {
		fmt.Printf("    {holding_register = %d, size = \"%s\", multiplier = %f, format = \"%s\" param_name = \"%s\"}", reading.HoldingRegister, reading.Size, reading.Multiplier, reading.Format, reading.ParamName)
		if i < len(config.Modbus.Registers)-1 {
			fmt.Printf(",")
		}
		fmt.Printf("\n")
	}
	fmt.Printf("]\n")
	fmt.Printf("\n")
	fmt.Printf("[mqtt]\n")
	fmt.Printf("url = \"%s\"\n", config.Modbus.URL)
	fmt.Printf("qos = %d\n", config.Mqtt.Qos)
	fmt.Printf("client_id = \"%s\"\n", config.Mqtt.ClientId)
	fmt.Printf("connect_retry = %d\n", config.Mqtt.ConnectRetry)
	fmt.Printf("username = \"%s\"\n", config.Mqtt.Username)
	fmt.Printf("password = \"********\"\n")
	fmt.Printf("pub_topic = \"%s\"\n", config.Mqtt.PubTopic)
	fmt.Printf("pub_rate = %d\n", config.Mqtt.PubRate)
	fmt.Printf("\n")
	fmt.Printf("[template_data]\n")
	fmt.Printf("template_file = \"%s\"\n", config.Template.TemplateFile)
	fmt.Printf("\n")
	fmt.Printf("[template_data.template_kv]\n")
	for k, v := range config.Template.TemplateKV {
		fmt.Printf("\"%s\" = \"%s\"\n", k, v)
	}

}

func generateExampleConfig() string {
	c := "[modbus]\n"
	c += "url = \"tcp://localhost:502\"\n"
	c += "unit_id = 1\n"
	c += "timeout = 1000\n"
	c += "poll_rate = 1000\n"
	c += "reconnect_pause = 5000\n"
	c += "registers = [\n"
	c += "    {holding_register = 40001, size = \"UINT32\", multiplier = 0.01, format = \"%.4f\", param_name = \"my_param_name_1\"},\n"
	c += "    {holding_register = 40003, size = \"SINT32\", multiplier = 0.01, format = \"%.4f\", param_name = \"my_param_name_2\"},\n"
	c += "    {holding_register = 40005, size = \"UINT64\", multiplier = 0.01, format = \"%.4f\", param_name = \"my_param_name_3\"}\n"
	c += "]\n"
	c += "\n"
	c += "[mqtt]\n"
	c += "url = \"tcp://localhost:1883\"\n"
	c += "qos = 0\n"
	c += "client_id = \"some_unique_client_id\" # 23 characters max\n"
	c += "connect_retry = 1000\n"
	c += "username = \"my_username\"\n"
	c += "password = \"my_password\"\n"
	c += "pub_topic = \"my/special/topic\"\n"
	c += "pub_rate = 1000\n"
	c += "\n"
	c += "[template_data]\n"
	c += "template_file = \"/some/file/path/filename.go.tmpl\"\n"
	c += "\n"
	c += "[template_data.template_kv]\n"
	c += "key_name_1 = \"key_value_1\"\n"
	c += "key_name_2 = \"key_value_2\"\n"
	c += "key_name_3 = \"key_value_3\"\n"

	return c
}
