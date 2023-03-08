package main

import (
	"crypto/rand"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/simonvetter/modbus"
	"gitlab.com/mthollylab/modbus2mqtt/logging"
)

type RegisterConfig struct {
	HoldingRegister  int    `toml:"holding_register"`
	Size             string `toml:"size"`
	Endianness       string `toml:"endianness"`
	ModbusEndianness modbus.Endianness
	WordOrder        string `toml:"word_order"`
	ModbusWordOrder  modbus.WordOrder
	Multiplier       float32 `toml:"multiplier"`
	Format           string  `toml:"format"`
	ParamName        string  `toml:"param_name"`
}

type ModbusInfo struct {
	URL       string           `toml:"url"`
	UnitId    int64            `toml:"unit_id"`
	Timeout   int64            `toml:"timeout"`
	PollRate  int64            `toml:"poll_rate"`
	Registers []RegisterConfig `toml:"registers"`
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

type MonitoringInfo struct {
	Enabled bool   `toml:"enabled"`
	Port    uint16 `toml:"port"`
}

type tomlConfig struct {
	Modbus     ModbusInfo     `toml:"modbus"`
	Mqtt       MqttInfo       `toml:"mqtt"`
	Template   TemplateInfo   `toml:"template_data"`
	Monitoring MonitoringInfo `toml:"monitoring"`
}

func loadConfig(filePath string) (tomlConfig, error) {
	var config tomlConfig
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return config, err
	}

	// Add some sane defaults for fields that are not provided

	// Modbus
	if len(config.Modbus.URL) == 0 {
		logging.Fatal("Modbus URL is required")
	}
	if config.Modbus.UnitId == 0 {
		config.Modbus.UnitId = 1
	}
	if config.Modbus.Timeout == 0 {
		config.Modbus.Timeout = 5000
	}
	if config.Modbus.PollRate == 0 {
		config.Modbus.PollRate = 1000
	}
	if config.Modbus.Registers == nil || len(config.Modbus.Registers) == 0 {
		logging.Fatal("You must define at least 1 Modbus register to read")
	}
	for i := 0; i < len(config.Modbus.Registers); i++ {
		if config.Modbus.Registers[i].Multiplier == 0.0 {
			config.Modbus.Registers[i].Multiplier = 1.0
		}
		if config.Modbus.Registers[i].Format == "" {
			config.Modbus.Registers[i].Format = "%.0f"
		}

		// Default to BIG_ENDIAN
		switch config.Modbus.Registers[i].Endianness {
		case "BIG", "BIG_ENDIAN":
			config.Modbus.Registers[i].ModbusEndianness = modbus.BIG_ENDIAN
		case "LITTLE", "LITTLE_ENDIAN":
			config.Modbus.Registers[i].ModbusEndianness = modbus.LITTLE_ENDIAN
		default:
			config.Modbus.Registers[i].Endianness = "BIG_ENDIAN"
			config.Modbus.Registers[i].ModbusEndianness = modbus.BIG_ENDIAN
		}

		// Default to HIGH_WORK_FIRST
		switch config.Modbus.Registers[i].WordOrder {
		case "HIGH", "HIGH_WORD", "HIGH_WORD_FIRST":
			config.Modbus.Registers[i].ModbusWordOrder = modbus.HIGH_WORD_FIRST
		case "LOW", "LOW_WORD", "LOW_WORD_FIRST":
			config.Modbus.Registers[i].ModbusWordOrder = modbus.LOW_WORD_FIRST
		default:
			config.Modbus.Registers[i].WordOrder = "HIGH_WORD_FIRST"
			config.Modbus.Registers[i].ModbusWordOrder = modbus.HIGH_WORD_FIRST
		}
	}

	// MQTT
	if len(config.Mqtt.URL) == 0 {
		logging.Fatal("MQTT URL is required")
	}
	// QOS will default to 0 if not provided
	if config.Mqtt.ClientId == "" {
		// Generate a random client id
		baseStr := "modbus2mqtt-"

		chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
		ll := len(chars)
		length := 23 - len(baseStr)
		b := make([]byte, length)
		rand.Read(b) // generates len(b) random bytes
		for i := 0; i < length; i++ {
			b[i] = chars[int(b[i])%ll]
		}

		randClientId := baseStr + string(b)

		fields := logging.NewFieldMap("ClientId", randClientId)
		logging.Infof("Generated random client id for MQTT connection", fields)

		config.Mqtt.ClientId = randClientId
	}
	if config.Mqtt.ConnectRetry == 0 {
		config.Mqtt.ConnectRetry = 1000
	}
	if len(config.Mqtt.Username) > 0 || len(config.Mqtt.Password) > 0 {
		if len(config.Mqtt.Username) == 0 {
			logging.Fatal("MQTT username is required when providing a password")
		}
		if len(config.Mqtt.Password) == 0 {
			logging.Fatal("MQTT password is required when providing a username")
		}
	}
	if len(config.Mqtt.PubTopic) == 0 {
		logging.Fatal("MQTT topic is required")
	}
	if config.Mqtt.PubRate == 0 {
		config.Mqtt.PubRate = 1000
	}

	// Template Data
	if len(config.Template.TemplateFile) == 0 {
		logging.Fatal("Path to template file is required")
	}

	// Monitoring
	if config.Monitoring.Port <= 0 {
		config.Monitoring.Port = 62112
	}

	return config, nil
}

func dumpConfig(config tomlConfig) {
	fmt.Printf("[modbus]\n")
	fmt.Printf("url = \"%s\"\n", config.Modbus.URL)
	fmt.Printf("unit_id = %d\n", config.Modbus.UnitId)
	fmt.Printf("timeout = %d\n", config.Modbus.Timeout)
	fmt.Printf("poll_rate = %d\n", config.Modbus.PollRate)
	fmt.Printf("registers = [\n")
	for i, register := range config.Modbus.Registers {
		fmt.Printf("    {holding_register = %d, size = \"%s\", endianness = \"%s\", word_order = \"%s\", multiplier = %f, format = \"%s\" param_name = \"%s\"}", register.HoldingRegister, register.Size, register.Endianness, register.WordOrder, register.Multiplier, register.Format, register.ParamName)
		if i < len(config.Modbus.Registers)-1 {
			fmt.Printf(",")
		}
		fmt.Printf("\n")
	}
	fmt.Printf("]\n")
	fmt.Printf("\n")
	fmt.Printf("[mqtt]\n")
	fmt.Printf("url = \"%s\"\n", config.Mqtt.URL)
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
	fmt.Printf("\n")
	fmt.Printf("[monitoring]\n")
	fmt.Printf("enabled = %v\n", config.Monitoring.Enabled)
	fmt.Printf("port = %d\n", config.Monitoring.Port)

}

func generateExampleConfig() string {
	c := "[modbus]\n"
	c += "url = \"tcp://localhost:502\"\n"
	c += "unit_id = 1\n"
	c += "timeout = 1000\n"
	c += "poll_rate = 1000\n"
	c += "registers = [\n"
	c += "    {holding_register = 40001, size = \"UINT32\", endianness = \"BIG_ENDIAN\", word_order = \"HIGH_WORD_FIRST\", multiplier = 0.01, format = \"%.4f\", param_name = \"my_param_name_1\"},\n"
	c += "    {holding_register = 40003, size = \"SINT32\", endianness = \"BIG_ENDIAN\", word_order = \"HIGH_WORD_FIRST\", multiplier = 0.01, format = \"%.4f\", param_name = \"my_param_name_2\"},\n"
	c += "    {holding_register = 40005, size = \"UINT64\", endianness = \"BIG_ENDIAN\", word_order = \"HIGH_WORD_FIRST\", multiplier = 0.01, format = \"%.4f\", param_name = \"my_param_name_3\"}\n"
	c += "    {holding_register = 40005, size = \"FLOAT32\", endianness = \"BIG_ENDIAN\", word_order = \"HIGH_WORD_FIRST\", multiplier = 0.01, format = \"%.4f\", param_name = \"my_param_name_4\"}\n"
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
	c += "\n"
	c += "[monitoring]\n"
	c += "enabled = true\n"
	c += "port = 62112\n"

	return c
}
