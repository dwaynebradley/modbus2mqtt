package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/template"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/simonvetter/modbus"
	"gitlab.com/mthollylab/modbus2mqtt/logging"
)

type SafeJson struct {
	mu         sync.Mutex
	JsonString string
}

func showUsageAndExit(exitcode int) {
	fmt.Println("Usage: modbus2mqtt [-c file] [-g] [-f file] [-loglevel LEVEL] [-h]")
	flag.PrintDefaults()

	os.Exit(exitcode)
}

func main() {
	configPath := flag.String("c", "config.toml", "Path to configuration TOML file.")
	generateSampleConfig := flag.Bool("g", false, "Generate a sample configuration.")
	generateConfigPath := flag.String("f", "", "Path where to write example configuration file when used with -g.")
	loggingLevel := flag.String("loglevel", "ERROR", "Logging level (TRACE, DEBUG, INFO, WARN, ERROR, FATAL)")
	showHelp := flag.Bool("h", false, "Show help message")
	flag.Parse()

	if *showHelp {
		showUsageAndExit(0)
	}

	logging.SetLoggingLevel(*loggingLevel)

	var err error

	if *generateSampleConfig {
		logging.Info("Generating sample configuration file")

		ec := generateExampleConfig()

		if generateConfigPath == nil || len(*generateConfigPath) == 0 {
			// Just print the example to the screen
			fmt.Print(ec)
		} else {
			if _, err := os.Stat(*generateConfigPath); err == nil {
				// path/to/whatever exists - delete it so we can replace it
				err = os.Remove(*generateConfigPath)
				if err != nil {
					fields := logging.NewFieldMap("err", err.Error())
					logging.Fatalf("Could not remove previous sample configuration file", fields)
				}
			} else if os.IsNotExist(err) {
				// path/to/whatever does *not* exist - good to go
			} else {
				// Schrodinger: file may or may not exist. See err for details.
				fields := logging.NewFieldMap("err", err.Error())
				logging.Fatalf("Unknown error", fields)
			}

			f, err := os.Create(*generateConfigPath)
			if err != nil {
				fields := logging.NewFieldMap("err", err.Error())
				logging.Fatalf("Could not create sample configuration file", fields)
			}

			defer f.Close()

			f.WriteString(ec)

			fields := logging.NewFieldMap("file", *generateConfigPath)
			logging.Infof("Created sample configuration file", fields)
		}
		os.Exit(0)
	}

	if configPath == nil || len(*configPath) == 0 {
		logging.Fatal("Path to configuration file not provided")
	}

	config, err := loadConfig(*configPath)
	if err != nil {
		fields := logging.NewFieldMap("err", err.Error())
		logging.AddField(fields, "file", *configPath)
		logging.Fatalf("Can't load configuration file", fields)
	} else {
		if logging.GetLoggingLevel() <= logging.DebugLevel {
			logging.Debug("===========================")
			logging.Debug("Configuration file: BEGIN")
			logging.Debug("===========================")
			dumpConfig(config)
			logging.Debug("===========================")
			logging.Debug("Configuration file: END")
			logging.Debug("===========================")
		}
	}

	// Define a "dec" function that we can use inside of the template since Go templates
	// cannot do simple "math" functions such as + or - by default
	funcMap := template.FuncMap{
		"dec": func(i int) int {
			return i - 1
		},
	}

	tmpl_bytes, err := os.ReadFile(config.Template.TemplateFile)
	if err != nil {
		fields := logging.NewFieldMap("err", err.Error())
		logging.Fatalf("Could not read template file", fields)
	}
	tmpl_content := string(tmpl_bytes)

	tmpl, err := template.New("payload").Funcs(funcMap).Parse(tmpl_content)
	if err != nil {
		fields := logging.NewFieldMap("err", err.Error())
		logging.Fatalf("Could not parse template file", fields)
	}

	// Connect to the MQTT broker where messages will be published
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.Mqtt.URL)
	opts.SetClientID(config.Mqtt.ClientId)
	opts.SetCleanSession(true)
	if len(config.Mqtt.Username) > 0 || len(config.Mqtt.Password) > 0 {
		opts.SetUsername(config.Mqtt.Username)
		opts.SetPassword(config.Mqtt.Password)
	}
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(time.Duration(config.Mqtt.ConnectRetry))
	opts.OnConnectionLost = connectionLostHandler
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		fields := logging.NewFieldMap("err", token.Error().Error())
		logging.Fatalf("MQTT client connection failed", fields)
	} else {
		logging.Info("MQTT client connection successful")
	}

	// Connect the modbus server
	var modbusClient *modbus.ModbusClient

	modbusClient, err = createModbusClient(config.Modbus.URL, time.Duration(config.Modbus.Timeout)*time.Millisecond)
	if err != nil {
		fields := logging.NewFieldMap("err", err.Error())
		logging.Fatalf("Modbus client connection failed", fields)
	} else {
		logging.Info("Modbus client connection successful")
	}
	if config.Modbus.UnitId > 0 {
		modbusClient.SetUnitId(uint8(config.Modbus.UnitId))
	}
	defer modbusClient.Close()

	wg := sync.WaitGroup{}

	var allDone []chan bool

	// ===========================================================
	// Poll the modbus server for updated data
	// ===========================================================
	modbusTicker := time.NewTicker(time.Duration(config.Modbus.PollRate) * time.Millisecond)
	modbusDone := make(chan bool)

	json := SafeJson{JsonString: ""}

	wg.Add(1)
	allDone = append(allDone, modbusDone)

	go func() {
		for {
			select {
			case <-modbusDone:
				modbusTicker.Stop()
				wg.Done()
				return
			case <-modbusTicker.C:
				// ================================================================
				// TODO: Actually poll the modbus device to get data. Just putting
				// random data into the readings for now.
				// ================================================================

				templateData := config.Template

				time_now := time.Now().UTC()
				tsInfo := TimestampInfo{
					Seconds:     time_now.Unix(),
					Nanoseconds: int32(time_now.Nanosecond()),
					RFC3339:     time_now.Format(time.RFC3339),
					RFC3339Nano: time_now.Format(time.RFC3339Nano),
				}
				templateData.Timestamp = tsInfo

				rvs, err := GetDeviceModbusData(modbusClient, config.Modbus.Registers)
				if err != nil {
					fields := logging.NewFieldMap("err", err.Error())
					logging.Fatalf("Error retrieving modbus data. Exiting.", fields)
				}
				templateData.RegisterValues = rvs

				var json_buf bytes.Buffer

				err = tmpl.Execute(&json_buf, templateData)
				if err != nil {
					fields := logging.NewFieldMap("err", err.Error())
					logging.Errorf("Could not execute template substition", fields)
				}

				json.mu.Lock()
				json.JsonString = json_buf.String()
				json.mu.Unlock()
			}
		}
	}()

	// ===========================================================
	// Publish the JSON template with updated data
	// ===========================================================
	mqttTicker := time.NewTicker(time.Duration(config.Mqtt.PubRate) * time.Millisecond)
	mqttDone := make(chan bool)

	wg.Add(1)
	allDone = append(allDone, mqttDone)

	go func() {
		for {
			select {
			case <-mqttDone:
				mqttTicker.Stop()
				wg.Done()
				return
			case <-mqttTicker.C:
				json.mu.Lock()
				j := json.JsonString
				json.mu.Unlock()

				if mqttClient.IsConnected() {
					// Copy source data to use for publishing. Lock the data before copying
					// so the other routines can keep flowing uninterrupted.
					fields := logging.NewFieldMap("payload", j)
					logging.Debugf("Payload to be published", fields)

					token := mqttClient.Publish(config.Mqtt.PubTopic, byte(config.Mqtt.Qos), false, j)
					token.Wait()

					logging.Debug("Payload published to MQTT")
				} else {
					fields := logging.NewFieldMap("payload", j)
					logging.Warningf("MQTT client is not connected. Cannot publish payload", fields)
				}
			}
		}
	}()

	signalChan := make(chan os.Signal, 1)
	doneChan := make(chan bool)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		// logging.Debug("Received an interrupt")

		// Stop all of the go routines
		for _, done := range allDone {
			done <- true
		}

		// Disconnect the MQTT client
		if mqttClient != nil && mqttClient.IsConnected() {
			mqttClient.Disconnect(500)
		}

		// Close modbus connection
		if modbusClient != nil {
			modbusClient.Close()
		}

		doneChan <- true
	}()
	<-doneChan
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fields := logging.NewFieldMap("err", err.Error())
	logging.Warningf("Connection lost to MQTT broker", fields)
}

func createModbusClient(url string, timeout time.Duration) (*modbus.ModbusClient, error) {
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     url,
		Timeout: timeout,
	})
	if err != nil {
		return nil, err
	}

	err = client.Open()
	if err != nil {
		return nil, err
	}

	return client, nil
}
