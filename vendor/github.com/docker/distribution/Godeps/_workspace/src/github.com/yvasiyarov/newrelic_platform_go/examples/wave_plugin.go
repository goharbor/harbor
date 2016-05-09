package main

import (
	"github.com/yvasiyarov/newrelic_platform_go"
)

type WaveMetrica struct {
	sawtoothMax     int
	sawtoothCounter int
}

func (metrica *WaveMetrica) GetName() string {
	return "Wave_Metrica"
}
func (metrica *WaveMetrica) GetUnits() string {
	return "Queries/Second"
}
func (metrica *WaveMetrica) GetValue() (float64, error) {
	metrica.sawtoothCounter++
	if metrica.sawtoothCounter > metrica.sawtoothMax {
		metrica.sawtoothCounter = 0
	}
	return float64(metrica.sawtoothCounter), nil
}

type SquareWaveMetrica struct {
	squarewaveMax     int
	squarewaveCounter int
}

func (metrica *SquareWaveMetrica) GetName() string {
	return "SquareWave_Metrica"
}
func (metrica *SquareWaveMetrica) GetUnits() string {
	return "Queries/Second"
}
func (metrica *SquareWaveMetrica) GetValue() (float64, error) {
	returnValue := 0
	metrica.squarewaveCounter++

	if metrica.squarewaveCounter < (metrica.squarewaveMax / 2) {
		returnValue = 0
	} else {
		returnValue = metrica.squarewaveMax
	}

	if metrica.squarewaveCounter > metrica.squarewaveMax {
		metrica.squarewaveCounter = 0
	}
	return float64(returnValue), nil
}

func main() {
	plugin := newrelic_platform_go.NewNewrelicPlugin("0.0.1", "7bceac019c7dcafae1ef95be3e3a3ff8866de246", 60)
	component := newrelic_platform_go.NewPluginComponent("Wave component", "com.exmaple.plugin.gowave")
	plugin.AddComponent(component)

	m := &WaveMetrica{
		sawtoothMax:     10,
		sawtoothCounter: 5,
	}
	component.AddMetrica(m)

	m1 := &SquareWaveMetrica{
		squarewaveMax:     4,
		squarewaveCounter: 1,
	}
	component.AddMetrica(m1)

	plugin.Verbose = true
	plugin.Run()
}
