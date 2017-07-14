package sensors

import (
	"context"
	"math/rand"
	"time"
)

type temp_humidity struct {
	ctx  context.Context
	out  chan<- SensorMessage
	name string
}

func NewTemperatureHumiditySensor(ctx context.Context, name string, out chan<- SensorMessage) *temp_humidity {
	return &temp_humidity{
		ctx:  ctx,
		name: name,
		out:  out,
	}
}

func (t *temp_humidity) Start() {
	go t.loop()
}

func (t *temp_humidity) loop() {

	temp := rand.Float32() * 22
	humidity := rand.Float32() * 34

	// schema for the data
	schema := map[string]interface{}{
		"@context": map[string]interface{}{
			"decode":   "http://decode.eu#",
			"m3-lite":  "http://purl.org/iot/vocab/m3-lite#",
			"humidity": "m3-lite:AirHumidity",
			"temp":     "m3-lite:AirTemperature",
			"domain":   "decode:hasDomain",
		},
		"@type": "m3-lite:Sensor",
		"domain": map[string]interface{}{
			"@type": "m3-lite:Environment",
		},
	}
	ticker := time.Tick(10 * time.Second)

	for {

		select {
		case <-t.ctx.Done():
			return
		case <-ticker:

			diff := rand.Float32()

			if diff > 0.5 {
				temp += diff
				humidity += diff
			} else {
				temp -= diff
				humidity -= diff
			}

			data := map[string]interface{}{
				"temp":     temp,
				"humidity": humidity,
			}

			t.out <- SensorMessage{
				Data:      data,
				Schema:    schema,
				SensorUID: t.name,
			}
		}
	}
}
