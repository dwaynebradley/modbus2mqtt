package main

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type ModbusInfo struct {
	URL            string `toml:"url"`
	UnitId         int64  `toml:"unit_id"`
	Timeout        int64  `toml:"timeout"`
	PollRate       int64  `toml:"poll_rate"`
	ReconnectPause int64  `toml:"reconnect_pause"`
}

type MqttInfo struct {
	URL          string `toml:"url"`
	Qos          int8   `toml:"qos"`
	ConnectRetry int64  `toml:"connect_retry"`
	Username     string `toml:"username"`
	Password     string `toml:"password"`
	PubTopic     string `toml:"pub_topic"`
	PubRate      int64  `toml:"pub_rate"`
}

type ReadingInfo struct {
	HoldingRegister int     `toml:"holding_register"`
	Size            string  `toml:"size"`
	Multiplier      float32 `toml:"multiplier"`
	Format          string  `toml:"format"`
	ParamName       string  `toml:"param_name"`
	Value           string
}

type TemplateKV struct {
	Timestamp     uint
	GatewayName   string        `toml:"gateway_name"`
	GatewayModel  string        `toml:"gateway_model"`
	GatewaySerial string        `toml:"gateway_serial"`
	DeviceName    string        `toml:"device_name"`
	DeviceModel   string        `toml:"device_model"`
	DeviceSerial  string        `toml:"device_serial"`
	Readings      []ReadingInfo `toml:"readings"`
}

type tomlConfig struct {
	Modbus     ModbusInfo `toml:"modbus"`
	Mqtt       MqttInfo   `toml:"mqtt"`
	TemplateKV TemplateKV `toml:"template_kv"`
}

func loadConfig(filePath string) (tomlConfig, error) {
	var config tomlConfig
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return config, err
	}

	for i := 0; i < len(config.TemplateKV.Readings); i++ {
		if config.TemplateKV.Readings[i].Multiplier == 0.0 {
			config.TemplateKV.Readings[i].Multiplier = 1.0
		}
		if config.TemplateKV.Readings[i].Format == "" {
			config.TemplateKV.Readings[i].Format = "%.0f"
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
	fmt.Printf("\n")
	fmt.Printf("[mqtt]\n")
	fmt.Printf("url = \"%s\"\n", config.Modbus.URL)
	fmt.Printf("qos = %d\n", config.Mqtt.Qos)
	fmt.Printf("connect_retry = %d\n", config.Mqtt.ConnectRetry)
	fmt.Printf("username = \"%s\"\n", config.Mqtt.Username)
	fmt.Printf("password = \"********\"\n")
	fmt.Printf("pub_topic = \"%s\"\n", config.Mqtt.PubTopic)
	fmt.Printf("pub_rate = %d\n", config.Mqtt.PubRate)
	fmt.Printf("\n")
	fmt.Printf("[template_kv]\n")
	fmt.Printf("gateway_name = \"%s\"\n", config.TemplateKV.GatewayName)
	fmt.Printf("gateway_model = \"%s\"\n", config.TemplateKV.GatewayModel)
	fmt.Printf("gateway_serial = \"%s\"\n", config.TemplateKV.GatewaySerial)
	fmt.Printf("device_name = \"%s\"\n", config.TemplateKV.DeviceName)
	fmt.Printf("device_model = \"%s\"\n", config.TemplateKV.DeviceModel)
	fmt.Printf("device_serial = \"%s\"\n", config.TemplateKV.DeviceSerial)
	fmt.Printf("readings = [\n")
	for i, reading := range config.TemplateKV.Readings {
		fmt.Printf("    {holding_register = %d, size = \"%s\", multiplier = %f, format = \"%s\" param_name = \"%s\"}", reading.HoldingRegister, reading.Size, reading.Multiplier, reading.Format, reading.ParamName)
		if i < len(config.TemplateKV.Readings)-1 {
			fmt.Printf(",")
		}
		fmt.Printf("\n")
	}
	fmt.Printf("]\n")

}

func generateDefaultConfig(filename string) {
	if _, err := os.Stat(filename); err == nil {
		// path/to/whatever exists - delete it
		err = os.Remove(filename)
		if err != nil {
			log.Fatal(err)
		}
	} else if os.IsNotExist(err) {
		// path/to/whatever does *not* exist - good to go
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		log.Fatal(err)
	}

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	f.WriteString("[modbus]\n")
	f.WriteString("url = \"tcp://localhost:502\"\n")
	f.WriteString("unit_id = 1\n")
	f.WriteString("timeout = 1000\n")
	f.WriteString("poll_rate = 1000\n")
	f.WriteString("reconnect_pause = 5000\n")
	f.WriteString("\n")
	f.WriteString("[mqtt]\n")
	f.WriteString("url = \"tcp://localhost:1883\"\n")
	f.WriteString("qos = 0\n")
	f.WriteString("connect_retry = 1000\n")
	f.WriteString("username = \"my_username\"\n")
	f.WriteString("password = \"my_password\"\n")
	f.WriteString("pub_topic = \"my/special/topic\"\n")
	f.WriteString("pub_rate = 1000\n")
	f.WriteString("\n")
	f.WriteString("[template_kv]\n")
	f.WriteString("gateway_name = \"my_gateway_name\"\n")
	f.WriteString("gateway_model = \"my_gateway_model\"\n")
	f.WriteString("gateway_serial = \"my_gateway_serial\"\n")
	f.WriteString("device_name = \"my_device_name\"\n")
	f.WriteString("device_model = \"my_device_model\"\n")
	f.WriteString("device_serial = \"my_device_serial\"\n")
	f.WriteString("readings = [\n")
	f.WriteString("    {holding_register = 40001, size = \"UINT32\", multiplier = 0.01, format = \"%.4f\", param_name = \"my_param_name_1\"},\n")
	f.WriteString("    {holding_register = 40003, size = \"UINT32\", multiplier = 0.01, format = \"%.4f\", param_name = \"my_param_name_2\"},\n")
	f.WriteString("    {holding_register = 40005, size = \"UINT32\", multiplier = 0.01, format = \"%.4f\", param_name = \"my_param_name_3\"}\n")
	f.WriteString("]\n")

	f.Sync()
}
