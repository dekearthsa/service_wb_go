package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// var mqttClient mqtt.Client
var client *mongo.Client
var collection *mongo.Collection

// var DB_HOST_NAME string = "mongodb://127.0.0.1:27017/smartFM"
var DB_HOST_NAME string = "mongodb://root:root@database:27017/smartFM?ssl=false&authSource=admin"

// var MQTT_HOST string = "tcp://localhost:1883"
var MQTT_HOST string = "mqtt://host.docker.internal:1883"

var MQTT_CLIENT_NAME string = "go_mqtt_client"
var MQTT_SUB_TOPIC string = "v1/events/data/update/json"
var MQTT_PUB_TOPIC string = "v1/events/data/command/json"
var DATABASE_NAME string = "smartFM"
var COLLECTION_DEVICE_REGISTER string = "registerandconfigs"
var COLLECTION_DEVICE_WB_SETTING string = "settingwbparams"
var COLLECTION_OCC_STATE string = "occupanystates"
var COLLECTION_WB_STATE string = "wellbreaths"
var COLLECTION_WB_RLY_SETTING string = "settingrlies"

type ModelWBDataIn struct {
	DeviceID   string          `json:"deviceID"`
	DeviceType string          `json:"deviceType"`
	Token      string          `json:"token"`
	Data       ModelDeviceData `json:"data"`
}

type ModelDeviceData struct {
	Temp  int32   `json:"temp"`
	Humid float32 `json:"humid"`
	Co2   int32   `json:"CO2"`
	PM25  int32   `json:"PM2.5"`
	VOC   int32   `json:"VOC"`
	Err   int32   `json:"error"`
}

type ModelRangeSettingRLY struct {
	IsAck []bool
	IsDef []string
}

type ModelCommand struct {
	DeviceID   string             `json:"deviceID"`
	DeviceType string             `json:"deviceType"`
	Token      string             `json:"token"`
	Data       ModelChannelActive `json:"data"`
}

type ModelChannelActive struct {
	Channel1 int32 `json:"channel1"`
	Channel2 int32 `json:"channel2"`
	Channel3 int32 `json:"channel3"`
	Channel4 int32 `json:"channel4"`
}

func createCommand(system string, definition []string, token string, isOff bool) (ModelCommand, bool, error) {
	const SUBSYSTEM = "WB_RLY"
	const DEVICE_TYPE = "relay"
	var settingCommand ModelCommand
	// var arrayRLY ModelRangeSettingRLY
	var settingRLY bson.M

	var ch1 int32
	var ch2 int32
	var ch3 int32
	var ch4 int32
	collection = client.Database(DATABASE_NAME).Collection(COLLECTION_WB_RLY_SETTING)
	err := collection.FindOne(context.TODO(), bson.D{{Key: "System", Value: system}, {Key: "Subsystem", Value: SUBSYSTEM}}).Decode(&settingRLY)
	// fmt.Println("settingRLY => ", settingRLY)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return settingCommand, false, err
		}
	}
	deviceID, ok := settingRLY["DeviceId"].(string)
	if !ok {
		log.Println("cannot find deviceID")
		return settingCommand, false, nil
	}

	if !isOff {

		set_ch1, ok := settingRLY["Channel_1"].(bson.M)
		if !ok {
			log.Println("cannot find Channel_1")
			return settingCommand, false, nil
		}
		set_ch2, ok := settingRLY["Channel_2"].(bson.M)
		if !ok {
			log.Println("cannot find Channel_2")
			return settingCommand, false, nil
		}
		set_ch3, ok := settingRLY["Channel_3"].(bson.M)
		if !ok {
			log.Println("cannot find Channel_3")
			return settingCommand, false, nil
		}
		set_ch4, ok := settingRLY["Channel_4"].(bson.M)
		if !ok {
			log.Println("cannot find Channel_4")
			return settingCommand, false, nil
		}

		isAct_ch1, ok := set_ch1["Is_active"].(bool)
		if !ok {
			log.Println("cannot find is active Channel_1")
			return settingCommand, false, nil
		}
		isAct_ch2, ok := set_ch2["Is_active"].(bool)
		if !ok {
			log.Println("cannot find is active Channel_2")
			return settingCommand, false, nil
		}
		isAct_ch3, ok := set_ch3["Is_active"].(bool)
		if !ok {
			log.Println("cannot find is active Channel_3")
			return settingCommand, false, nil
		}
		isAct_ch4, ok := set_ch4["Is_active"].(bool)
		if !ok {
			log.Println("cannot find is active Channel_4")
			return settingCommand, false, nil
		}

		def_ch1, ok := set_ch1["Definition"].(string)
		if !ok {
			log.Println("cannot find is Definition Channel_1")
			return settingCommand, false, nil
		}
		def_ch2, ok := set_ch2["Definition"].(string)
		if !ok {
			log.Println("cannot find is Definition Channel_2")
			return settingCommand, false, nil
		}
		def_ch3, ok := set_ch3["Definition"].(string)
		if !ok {
			log.Println("cannot find is Definition Channel_3")
			return settingCommand, false, nil
		}
		def_ch4, ok := set_ch4["Definition"].(string)
		if !ok {
			log.Println("cannot find is Definition Channel_4")
			return settingCommand, false, nil
		}

		if isAct_ch1 {
			for _, def := range definition {
				if def == def_ch1 {
					ch1 = 1
				}
			}

			if ch1 != 1 {
				ch1 = 0
			}
		}

		if isAct_ch2 {
			for _, def := range definition {
				if def == def_ch2 {
					ch2 = 1
				}
			}
			if ch2 != 1 {
				ch2 = 0
			}
		}

		if isAct_ch3 {
			for _, def := range definition {
				if def == def_ch3 {
					ch3 = 1
				}
			}
			if ch3 != 1 {
				ch3 = 0
			}
		}

		if isAct_ch4 {
			for _, def := range definition {
				if def == def_ch4 {
					ch4 = 1
				}
			}
			if ch4 != 1 {
				ch4 = 0
			}
		}

		var arrayRLY = ModelCommand{
			DeviceID:   deviceID,
			DeviceType: DEVICE_TYPE,
			Token:      token,
			Data: ModelChannelActive{
				Channel1: ch1,
				Channel2: ch2,
				Channel3: ch3,
				Channel4: ch4,
			},
		}
		// fmt.Println("arrayRLY => ", arrayRLY)
		return arrayRLY, false, nil
	} else {
		var offRelay = ModelCommand{
			DeviceID:   deviceID,
			DeviceType: DEVICE_TYPE,
			Token:      token,
			Data: ModelChannelActive{
				Channel1: 0,
				Channel2: 0,
				Channel3: 0,
				Channel4: 0,
			},
		}
		// fmt.Println("offRelay => ", offRelay)
		return offRelay, false, nil
	}
}

func sendCommand(command ModelCommand, mqtt_client mqtt.Client) error {
	// fmt.Println("command => ", command)
	b, err := json.Marshal(command)
	if err != nil {
		return err
	}
	// fmt.Println("sendCommand => ", string(b))
	token := mqtt_client.Publish(MQTT_PUB_TOPIC, 0, false, string(b))
	token.Wait()
	// mqtt_client.Disconnect(250)
	return nil
}

func occupanyState(system string) (bool, error) {
	var stateResult bson.M
	var ms int64 = time.Now().UnixNano() / 1e6
	collection = client.Database(DATABASE_NAME).Collection(COLLECTION_OCC_STATE)
	err := collection.FindOne(context.TODO(), bson.D{{Key: "System", Value: system}}).Decode(&stateResult)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, mongo.ErrNoDocuments
		} else {
			log.Fatal(err)
		}
	}

	setTime, ok := stateResult["TimeRemaining"].(float64)
	if !ok {
		fmt.Println("error get TimeRemaining in func occupanyState.")
	}
	if int64(setTime) <= ms {
		return true, nil
	}

	return false, nil
}

func operatLogic(
	system string,
	token string,

	co2 int32,
	humid float32,
	temp float32,

	cos_mor_thes int32,
	cos_les_thes int32,

	humid_mor_thes float32,
	humid_les_thes float32,

	temp_mor_thes int32,
	temp_les_thes int32,
) (ModelCommand, error) {
	var setStringCommand []string
	var isCommand ModelCommand

	if co2 >= cos_mor_thes && temp < float32(temp_les_thes) && humid < humid_les_thes {
		setStringCommand = append(setStringCommand, "WB_EXHAUST_FAN")
		setStringCommand = append(setStringCommand, "WB_SUPPLY_FAN_LOW")
		command, errStatus, err := createCommand(system, setStringCommand, token, false)
		// fmt.Println("step 1 => ", command)
		if err != nil {
			log.Panic(err)
			return command, err
		}
		if !errStatus {
			return command, nil
		}
	} else if co2 >= cos_mor_thes && (temp >= float32(temp_mor_thes) || humid >= humid_mor_thes) {
		isEnd, err := occupanyState(system)
		if err != nil {
			log.Panic(err)
			return isCommand, err
		}
		// fmt.Println("isEnd 1 => ", isEnd)
		if !isEnd {
			setStringCommand = append(setStringCommand, "WB_EXHAUST_FAN")
			setStringCommand = append(setStringCommand, "WB_SUPPLY_FAN_HIGH")
			command, errStatus, err := createCommand(system, setStringCommand, token, false)
			// fmt.Println("step 2 => ", command)
			if err != nil {
				log.Panic(err)
				return command, err
			}

			if !errStatus {
				// fmt.Println("step 2 errStatus  => ", errStatus)
				return command, nil
			}
			// fmt.Println("step 2 none => ", command)
		} else {
			setStringCommand = append(setStringCommand, "WB_EXHAUST_FAN")
			setStringCommand = append(setStringCommand, "WB_SUPPLY_FAN_LOW")
			command, errStatus, err := createCommand(system, setStringCommand, token, false)
			// fmt.Println("step 3 => ", command)
			if err != nil {
				log.Panic(err)
				return command, err
			}
			if !errStatus {
				return command, nil
			}
		}
	} else if (temp >= float32(temp_mor_thes) || humid >= humid_mor_thes) && co2 < cos_les_thes {
		isEnd, err := occupanyState(system)
		// fmt.Println("isEnd 2 => ", isEnd)
		if err != nil {
			log.Panic(err)
			return isCommand, err
		}
		if !isEnd {
			setStringCommand = append(setStringCommand, "WB_EXHAUST_FAN")
			setStringCommand = append(setStringCommand, "WB_SUPPLY_FAN_HIGH")
			command, errStatus, err := createCommand(system, setStringCommand, token, false)
			// fmt.Println("step 4 => ", command)
			if err != nil {
				log.Panic(err)
				return command, err
			}
			if !errStatus {
				return command, nil
			}
		} else {
			setStringCommand = append(setStringCommand, "WB_EXHAUST_FAN")
			setStringCommand = append(setStringCommand, "WB_SUPPLY_FAN_LOW")
			command, errStatus, err := createCommand(system, setStringCommand, token, false)
			// fmt.Println("step 5 => ", command)
			if err != nil {
				log.Panic(err)
				return command, err
			}
			if !errStatus {
				return command, nil
			}
		}
	} else if co2 < cos_les_thes && temp < float32(temp_les_thes) && humid < humid_les_thes {
		command, errStatus, err := createCommand(system, setStringCommand, token, true)
		// fmt.Println("step 6 end => ", command)
		if err != nil {
			log.Panic(err)
			return command, err
		}
		if !errStatus {
			// isCommand = command
			return command, nil
		}
	}
	return isCommand, nil
}

func fetchThesHoldWB(system string) (int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, error) {
	var result bson.M
	collection = client.Database(DATABASE_NAME).Collection(COLLECTION_DEVICE_WB_SETTING)
	err := collection.FindOne(context.TODO(), bson.D{{Key: "System", Value: system}}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
		} else {
			log.Fatal(err)
		}
	}
	Is_thres_co2_more_than, ok := result["Is_thres_co2_more_than"].(int32)
	if !ok {
		log.Println("Is_thres_co2_more_than field is not of type int32")
	}
	Is_thres_pm2_more_than, ok := result["Is_thres_pm2_more_than"].(int32)
	if !ok {
		log.Println("Is_thres_pm2_more_than field is not of type int32")
	}
	Is_thres_temp_more_than, ok := result["Is_thres_temp_more_than"].(int32)
	if !ok {
		log.Println("Is_thres_temp_more_than field is not of type int32")
	}
	Is_thres_humid_more_than, ok := result["Is_thres_humid_more_than"].(int32)
	if !ok {
		log.Println("Is_thres_humid_more_than field is not of type int32")
	}
	Is_thres_voc_more_than, ok := result["Is_thres_voc_more_than"].(int32)
	if !ok {
		log.Println("Is_thres_voc_more_than field is not of type int32")
	}
	Is_thres_co2_lower_than, ok := result["Is_thres_co2_lower_than"].(int32)
	if !ok {
		log.Println("Is_thres_co2_lower_than field is not of type int32")
	}
	Is_thres_pm2_lower_than, ok := result["Is_thres_pm2_lower_than"].(int32)
	if !ok {
		log.Println("Is_thres_pm2_lower_than field is not of type int32")
	}
	Is_thres_temp_lower_than, ok := result["Is_thres_temp_lower_than"].(int32)
	if !ok {
		log.Println("Is_thres_temp_lower_than field is not of type int32")
	}
	Is_thres_humid_lower_than, ok := result["Is_thres_humid_lower_than"].(int32)
	if !ok {
		log.Println("Is_thres_humid_lower_than field is not of type int32")
	}
	Is_thres_voc_lower_than, ok := result["Is_thres_voc_lower_than"].(int32)
	if !ok {
		log.Println("Is_thres_voc_lower_than field is not of type int32")
	}
	return Is_thres_co2_more_than, Is_thres_pm2_more_than, Is_thres_temp_more_than, Is_thres_humid_more_than, Is_thres_voc_more_than, Is_thres_co2_lower_than, Is_thres_pm2_lower_than, Is_thres_temp_lower_than, Is_thres_humid_lower_than, Is_thres_voc_lower_than, nil
}

func checkingDevice(deivceID string, deviceType string, token string) (bool, string, string, error) {
	if deviceType == "WB" {
		var result bson.M
		collection = client.Database(DATABASE_NAME).Collection(COLLECTION_DEVICE_REGISTER)
		err := collection.FindOne(context.TODO(), bson.D{{Key: "DeviceId", Value: deivceID}}).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return false, "", token, err
			} else {
				log.Fatal(err)
			}
		}

		tokenConvert, ok := result["Token"]
		if !ok {
			log.Println("Token is not string")
			return false, "", token, nil
		}

		if tokenConvert != token {
			log.Println("Invalid token")
			return false, "", token, nil
		}
		system, ok := result["System"].(string)
		if !ok {
			log.Println("System is not string")
			return false, "", token, nil
		}

		isAuto, ok := result["IsAuto"].(bool)
		if !ok {
			log.Println("isAuto is not bool")
			return false, "", token, nil
		}
		if !isAuto {
			return false, "", token, nil
		}
		return true, system, token, nil
	}
	return false, "", token, nil
}

func init() {
	fmt.Println("Starting service_go_wb...")
	fmt.Println("Connecting MongoDB...")
	clientOptions := options.Client().ApplyURI(DB_HOST_NAME)
	var err error
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
}

func main() {
	fmt.Println("Connecting mqtt host " + MQTT_HOST + "....")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(MQTT_HOST)
	opts.SetClientID(MQTT_CLIENT_NAME)
	opts.SetDefaultPublishHandler(messageHandler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	fmt.Println("Connected mqtt host " + MQTT_HOST)
	fmt.Println("Subscribe topic " + MQTT_SUB_TOPIC + "....")
	if token := client.Subscribe(MQTT_SUB_TOPIC, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	fmt.Println("Subscribe topic " + MQTT_SUB_TOPIC)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Service_wb started to shut down service press control+C")
	s := <-sigc
	fmt.Printf("Received signal: %s, shutting down...\n", s)

	client.Disconnect(250)
}

var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var result ModelWBDataIn
	err := json.Unmarshal([]byte(msg.Payload()), &result)
	if err != nil {
		log.Println("cannot Unmarshal or wrong format data")
		log.Panicln(err)
	}
	status, system, token, _ := checkingDevice(result.DeviceID, result.DeviceType, result.Token)
	if status {
		Is_thres_co2_more_than, _, Is_thres_temp_more_than, Is_thres_humid_more_than, _, Is_thres_co2_lower_than, _, Is_thres_temp_lower_than, Is_thres_humid_lower_than, _, err := fetchThesHoldWB(system)
		if err != nil {
			log.Println(err)
		}
		command, err := operatLogic(system, token, result.Data.Co2, float32(result.Data.Humid), float32(result.Data.Temp)/10, Is_thres_co2_more_than, Is_thres_co2_lower_than, float32(Is_thres_humid_more_than), float32(Is_thres_humid_lower_than), Is_thres_temp_more_than, Is_thres_temp_lower_than)
		if err != nil {
			log.Panic(err)
		}
		err = sendCommand(command, client)
		if err != nil {
			log.Println(err)
		}
	}
}
