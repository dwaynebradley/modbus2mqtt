package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/template"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/simonvetter/modbus"
)

type SafeJson struct {
	mu         sync.Mutex
	JsonString string
}

var showDebugInfo bool

func main() {
	configPath := flag.String("c", "config.toml", "Path to configuration TOML file.")
	generateConfigPath := flag.String("g", "", "Path where the default configuration should be generated.")
	debug := flag.Bool("d", false, "Show debugging information.")

	var err error

	showDebugInfo = false
	if debug != nil {
		showDebugInfo = *debug
	}

	log.SetFlags(log.LstdFlags)
	flag.Parse()

	if generateConfigPath != nil && len(*generateConfigPath) != 0 {
		generateDefaultConfig(*generateConfigPath)
		os.Exit(0)
	}

	if configPath == nil || len(*configPath) == 0 {
		log.Fatal("You must supply a configuration file")
	}

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Can't load configuration file: %v\n", err)
	} else {
		if showDebugInfo {
			log.Println("===========================")
			log.Println("Configuration file: BEGIN")
			log.Println("===========================")
			dumpConfig(config)
			log.Println("===========================")
			log.Println("Configuration file: END")
			log.Println("===========================")
		}
	}

	// Connect to the MQTT broker where messages will be published
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.Mqtt.URL)
	opts.SetClientID("modbus2mqtt")
	opts.SetCleanSession(true)
	opts.SetUsername(config.Mqtt.Username)
	opts.SetPassword(config.Mqtt.Password)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(time.Duration(config.Mqtt.ConnectRetry))
	opts.OnConnectionLost = connectionLostHandler
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Connect the modbus server
	var modbusClient *modbus.ModbusClient

	modbusClient, err = createModbusClient(config.Modbus.URL, time.Duration(config.Modbus.Timeout)*time.Millisecond)
	if err != nil {
		log.Fatalf("Can't establish modbus connection: %v\n", err)
	} else {
		log.Printf("Modbus connection successful: %s\n", config.Modbus.URL)
	}
	if config.Modbus.UnitId > 0 {
		modbusClient.SetUnitId(uint8(config.Modbus.UnitId))
	}
	defer modbusClient.Close()

	wg := sync.WaitGroup{}

	// Define a "dec" function that we can use inside of the template since Go templates
	// cannot do simple "math" functions such as + or - by default
	funcMap := template.FuncMap{
		"dec": func(i int) int {
			return i - 1
		},
	}

	tmpl, err := template.New("test").Funcs(funcMap).Parse(JSON_TEMPLATE)
	if err != nil {
		panic(err)
	}

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

				templateKV := config.TemplateKV

				templateKV.Timestamp = uint(time.Now().Unix())

				rdgs, err := GetDeviceModbusData(modbusClient, templateKV.Readings)
				if err != nil {
					panic(err)
				}
				templateKV.Readings = rdgs

				var json_buf bytes.Buffer

				err = tmpl.Execute(&json_buf, templateKV)
				if err != nil {
					panic(err)
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
				// Copy source data to use for publishing. Lock the data before copying
				// so the other routines can keep flowing uninterrupted.
				json.mu.Lock()
				j := json.JsonString
				json.mu.Unlock()

				token := mqttClient.Publish(config.Mqtt.PubTopic, byte(config.Mqtt.Qos), false, j)
				token.Wait()
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
	fmt.Printf("Connect lost: %v", err)
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
