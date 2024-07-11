package model

// mosquitto_pub -h localhost -p 1883 -t v1/events/data/update/json -m
// '{"deviceID":"wb-sensor", "deviceType":"WB", "token":"g91fu9l7", "data":{"temp":400,"humid":20.3,"CO2":400, "PM2.5":22, "VOC":22, "error":0}}'
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
