package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ModelWBDataIn struct {
	DeviceID   string          `json:"deviceID"`
	DeviceType string          `json:"deviceType"`
	Token      string          `json:"token"`
	Data       ModelDeviceData `json:"data"`
}

type ModelDeviceData struct {
	Temp  int32   `json:"temp"`
	Humid float32 `json:"humod"`
	Co2   int32   `json:"CO2"`
	PM25  int32   `json:"PM2.5"`
	VOC   int32   `json:"VOC"`
	Err   int32   `json:"error"`
}

var client *mongo.Client
var collection *mongo.Collection

func fetchThesHoldWB(system string) (int32, int32, int32, int32, int32, int32, int32, int32, int32, int32, error) {
	var result bson.M
	collection = client.Database("smartFM").Collection("settingwbparams")
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

func checkingDevice(deivceID string, deviceType string, token string) (bool, string, error) {
	if deviceType == "WB" {
		var result bson.M
		collection = client.Database("smartFM").Collection("registerandconfigs")
		err := collection.FindOne(context.TODO(), bson.D{{Key: "DeviceId", Value: deivceID}}).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return false, "", err
			} else {
				log.Fatal(err)
			}
		}

		tokenConvert, ok := result["Token"]
		if !ok {
			log.Println("Token is not string")
			return false, "", nil
		}

		if tokenConvert != token {
			log.Println("Invalid token")
			return false, "", nil
		}
		system, ok := result["System"].(string)
		if !ok {
			log.Println("System is not string")
			return false, "", nil
		}

		return true, system, nil
	}
	return false, "", nil
}

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017/smartFM")
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
	fmt.Println("Service_wb running ...")
	MQTT_HOST := "tcp://localhost:1883"
	MQTT_CLIENT_NAME := "go_mqtt_client"
	MQTT_SUB_TOPIC := "go-mqtt/sample"
	fmt.Println("connecting mqtt host " + MQTT_HOST + "....")
	opts := MQTT.NewClientOptions()
	opts.AddBroker(MQTT_HOST)
	opts.SetClientID(MQTT_CLIENT_NAME)
	opts.SetDefaultPublishHandler(messageHandler)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	fmt.Println("Subscribe topic " + MQTT_SUB_TOPIC + "....")
	if token := client.Subscribe(MQTT_SUB_TOPIC, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Service_wb started to shut down service press control+C")
	// Block until a signal is received
	s := <-sigc
	fmt.Printf("Received signal: %s, shutting down...\n", s)

	client.Disconnect(250)
}

var messageHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("Received message on topic: %s\nMessage: %s\n", msg.Topic(), msg.Payload())
	var result ModelWBDataIn
	err := json.Unmarshal([]byte(msg.Payload()), &result)
	if err != nil {
		log.Println("cannot Unmarshal or wrong format data")
		log.Panicln(err)
	}
	status, system, _ := checkingDevice(result.DeviceID, result.DeviceType, result.Token)
	if status != false {
		Is_thres_co2_more_than, Is_thres_pm2_more_than, Is_thres_temp_more_than, Is_thres_humid_more_than, Is_thres_voc_more_than, Is_thres_co2_lower_than, Is_thres_pm2_lower_than, Is_thres_temp_lower_than, Is_thres_humid_lower_than, Is_thres_voc_lower_than, err := fetchThesHoldWB(system)
		if err != nil {
			log.Println(err)
		}
	}
	// if result.DeviceType == "WB" {

	// }
}
