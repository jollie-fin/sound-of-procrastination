package libs

import (
	"math"
	"time"
)

const Tau = math.Pi * 2
const (
	Samplerate = 44100
	Timestep = 1*time.Second/Samplerate
	phasestep = Tau / Samplerate
)

// Stream value forever until dummy_channel no longer has any value
// It is used to pass a single value where is stream is expected
func Constant[K any, U any](dummy_channel chan U, value K) chan K {
	result := MakeStream[K]()
	go func(){
		for range dummy_channel {
			result<-value
		}
	}()
	return result
}
