package sensors

type SensorMessage struct {
	Data      map[string]interface{}
	Schema    map[string]interface{}
	SensorUID string
}
