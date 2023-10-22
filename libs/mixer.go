package libs

import "math"

// Take a stream and returns two copies
func Split[K any](in chan K) (chan K, chan K) {
	a,b := MakeStream[K](), MakeStream[K]()
	go func (){
		for value := range in {
			a<-value
			b<-value
		}
	}()
	return a,b
}

// Take a stream and returns n copies in a slice
func SplitIntoSlices[K any](in chan K, n int) []chan K {
	outs := make([]chan K, n)
	for i := range(outs) {
		outs[i] = MakeStream[K]()
	}
	go func (){
		for value := range in {
			for _, channel := range(outs) {
				channel<-value
			}
		}
	}()
	return outs
}

// Multiply multiples audiostream as if it was a voltage controlled amplifier
func Vca(value Audiostream, values ...AudiostreamOrFloat) Audiostream {
	out := MakeAudioStream()
	go func (){
		for acc := range value {
			for _,channel := range values {
				acc *= ReadFrom(channel)
			}
			out<-acc
		}
	}()
	return out
}

// Sum multiples audiostream as if it was resistor node
func Join(in Audiostream, ins ...AudiostreamOrFloat) Audiostream {
	out := MakeAudioStream()
	go func (){
		for acc := range in{
			for _,channel := range ins {
				acc += ReadFrom(channel)
			}
			out<-acc
		}
	}()
	return out
}

// Return in[selection]
func ChooseGeneric[K any](selection Audiostream, in ...interface{}) chan K {
	out := MakeStream[K]()
	go func (){
		for sel := range selection {
			selected := int(math.Round(sel)) % len(in)
			for index, channel := range in {
				value := ReadFromGeneric[K](channel)
				if index == selected {
					out<-value
				}
			}
		}
	}()
	return out
}

// Return in[selection]
func Choose(selection Audiostream, in ...AudiostreamOrFloat) chan float64 {
	out := MakeAudioStream()
	go func (){
		for sel := range selection {
			selected := int(math.Round(sel)) % len(in)
			for index, channel := range in {
				value := ReadFrom(channel)
				if index == selected {
					out<-value
				}
			}
		}
	}()
	return out
}
