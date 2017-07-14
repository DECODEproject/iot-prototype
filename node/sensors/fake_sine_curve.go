package sensors

import (
	"context"
	"math"
	"time"
)

type sine_curve struct {
	ctx   context.Context
	out   chan<- SensorMessage
	name  string
	value float64
}

func NewSineCurveEmitterSensor(ctx context.Context, name string, out chan<- SensorMessage) *sine_curve {
	return &sine_curve{
		ctx:  ctx,
		name: name,
		out:  out,
	}
}

func (t *sine_curve) Start() {
	go t.loop()
}

func (t *sine_curve) loop() {

	// schema for the data
	schema := map[string]interface{}{
		"@context": map[string]interface{}{
			"decode":  "http://decode.eu#",
			"m3-lite": "http://purl.org/iot/vocab/m3-lite#",
			"xsd":     "http://www.w3.org/2001/XMLSchema#",
			"value":   "xsd:float",
			"domain":  "decode:hasDomain",
		},
		"@type": "m3-lite:Sensor",
		"domain": map[string]interface{}{
			"@type": "decode:Fun",
		},
	}
	ticker := time.Tick(10 * time.Second)

	for {

		select {
		case <-t.ctx.Done():
			return
		case <-ticker:
			increase := 90 / 180 * math.Pi / 9
			v := 180 - math.Sin(t.value)*120
			t.value += increase

			data := map[string]interface{}{
				"value": v,
			}

			t.out <- SensorMessage{
				Data:      data,
				Schema:    schema,
				SensorUID: t.name,
			}
		}
	}
}
