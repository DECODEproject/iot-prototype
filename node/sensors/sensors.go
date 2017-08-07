package sensors

// SensorMessage defines an internal message struct for dealing with sensor data
type SensorMessage struct {
	Data      map[string]interface{}
	Schema    map[string]interface{}
	SensorUID string
}
