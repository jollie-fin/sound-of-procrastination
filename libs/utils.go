package libs

import (
	"fmt"
	"time"
)

type Audiostream chan float64
type AudiostreamOrFloat interface{}
type Signalstream chan bool
type SignalstreamOrBool interface{}
type IntstreamOrInt interface{}
type DurationstreamOrDuration interface{}

// Make a stream of type K
func MakeStream[K any]() chan K {
	return make(chan K, 10000)
}

// Make a audio stream (float)
func MakeAudioStream() Audiostream {
	return MakeStream[float64]()
}

// Make a signal stream
func MakeSignalStream() Signalstream {
	return MakeStream[bool]()
}

// Divide euclideanly, returns quot,rem 
func Divmod(d,m int) (int,int) {
	if d >= 0 {
		return d/m, d%m
	} else {
		return (d-m+1)/m,(d+1)%m+m-1
	}
}

// Listen to the whole stream and do nothing with it
func Sink[K any](in chan K) {
	go func (){
		for range in {}
	}()
}

// Get a value from :
// - a channel, by reading from it
// - a value, by returning it 
func ReadFromGeneric[K any](x interface{}) K {
	switch v := x.(type) {
	case K: return v
	case chan K: return <-v
	}
	panic(fmt.Sprintf("get can only read /(chan |)%T/; was %T",*new(K), x))
}

// Get a float from :
// - a channel, by reading from it
// - a float or an int, by returning it
func ReadFrom(x AudiostreamOrFloat) float64 {
	switch v := x.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case chan float64:
		return <-v
	case Audiostream:
		return <-v
	}
	panic(fmt.Sprintf("%T cannot be read into float64", x))
}

// Get a boolean from :
// - a signal channel, by reading from it
// - a bool, by returning it
func ReadTriggerFrom(x SignalstreamOrBool) bool {
	switch v := x.(type) {
	case bool:
		return v
	case chan bool:
		return <-v
	case Signalstream:
		return <-v
	}
	panic(fmt.Sprintf("%T cannot be read into float64", x))

}

// Get an int from :
// - an int channel, by reading from it
// - an int, by returning it
func ReadIntFrom(x IntstreamOrInt) int {
	switch v := x.(type) {
	case int:
		return v
	case chan int:
		return <-v
	}
	panic(fmt.Sprintf("%T cannot be read into float64", x))

}

// Get a duration from :
// - a duration channel, by reading from it
// - a duration, by returning it
func ReadDurationFrom(x DurationstreamOrDuration) time.Duration {
	switch v := x.(type) {
	case time.Duration:
		return v
	case chan time.Duration:
		return <-v
	}
	panic(fmt.Sprintf("%T cannot be read into float64", x))
}

