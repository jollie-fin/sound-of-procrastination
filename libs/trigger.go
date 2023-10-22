package libs

import (
	"errors"
	"math/rand"
	"time"
)

// Stream a random element from `from` everytime a Dirac is received
func StreamRandom[K any](dirac Signalstream, from []K) chan K {
	out := MakeStream[K]()
	go func (){
		randomize := func() int {return rand.Int() % len(from)} //Bad random
		value := randomize()
		for shall_trigger := range dirac {
			if shall_trigger {
				value=randomize()
			}
			out<-from[value]
		}
	}()
	return out
}

// Stream a random integer in [|min, max|]. The value changes after every dirac
func StreamRandomint(
		dirac Signalstream,
		min IntstreamOrInt, max IntstreamOrInt) Audiostream {
	out := MakeAudioStream()
	go func (){
		value := 0.
		first := true
		for shall_trigger := range dirac {
			if shall_trigger || first {
				low := ReadIntFrom(min)
				high := ReadIntFrom(max)
				value = float64(rand.Int() % (high - low + 1) + low)
				first = false
			}
			out<-value
		}
	}()
	return out
}

// Listen for Dirac, and emit valued[i-1] after ith dirac.
func Stepwise[K any](trigger Signalstream, values []K) chan K{
	out := MakeStream[K]()
	go func (){
		i := len(values)-1
		for shall_trigger := range trigger {
			if shall_trigger {
				i++
				i%=len(values)
			}
			out<-values[i]
		}
	}()
	return out
}

// Return N such as len(values[i]) == N \forall N
func computeWidth[K any](values [][]K) (int, error) {
	n := -1
	for _, slice := range values {
		candidate := len(slice)
		if n == -1 {
			n = candidate
		} else if n != candidate {
			return 0, errors.New("heterogeneous array")
		}
	}
	return n, nil
}

// Listen for Dirac, and emit multiples signals according to the following rule
// - The jth channel will have values[i-1][j] after ith dirac
func StepwiseMultiples[K any](dirac Signalstream,
	values [][]K) []chan K {
	n, _ := computeWidth(values)
	out := make([]chan K, n)
	for i := range out {
		out[i] = MakeStream[K]()
	}
	stream := Stepwise(dirac, values)
	go func (){
		for value := range stream {
			for i, channel := range out {
				channel<-value[i]
			}
		}
	}()
	return out
}

// Listen for Dirac, and emit them back according to the following rule
// - The ith dirac will be sent to the jth channel if values[i][j] is positive
func FilterDirac(dirac Signalstream,
	values [][]float64) []Signalstream {
	n, _ := computeWidth(values)
	outs := make([]Signalstream, n)
	for i := range outs {
		outs[i] = MakeSignalStream()
	}
	go func (){
		i := len(values)-1
		for shall_trigger := range dirac {
			if shall_trigger {
				i++
				i%=len(values)
			}
			for j, channel := range(outs) {
				channel<- values[i][j]>1e-3 && shall_trigger
			}
		}
	}()
	return outs
}

// Skip diracs so that only one dirac is passed through every oncein diracs
func DiracOnceIn(dirac Signalstream, oncein IntstreamOrInt) Signalstream {
	out := MakeSignalStream()
	go func (){
		i := 0
		for state := range dirac {
			out <- state && i == 0
			if state {
				i++;
				i = i % ReadIntFrom(oncein)
			}
		}
	}()
	return out
}

// Everytime signal goes from false to true, emit a Dirac
func EdgeDetect(signal Signalstream) Signalstream {
	out := MakeSignalStream()
	go func (){
		previous := false
		for state := range signal {
			if state && !previous {
				out <- true
			} else {
				out <- false
			}
		}
	}()
	return out
}

// Produce a dirac cyclically, every duration, until quit has a value
func DiracEvery(duration DurationstreamOrDuration, quit chan bool) Signalstream {
	out := MakeSignalStream()
	go func(){
		time:=0*time.Second
		
		for {
			value := false
			time+=Timestep
			d:=ReadDurationFrom(duration)
			if time>d {
				time -= d
				value = true
			}
			select {
			case <-quit:
				return
			case out<-value:
			}
		}
	}()
	return out
}

// Produce a calibrated impulsion every time a dirac is received
func DiracToImpulsion(dirac Signalstream,
	minduration DurationstreamOrDuration) Signalstream {
	out := MakeSignalStream()
	go func (){
		step := time.Duration(0)
		previous := false
		for is_trigger := range dirac {
			if is_trigger && !previous {
				step = time.Duration(0)
			}
			step+= Timestep
			if !is_trigger && previous {
				out <- false
			} else {
				out<- step < ReadDurationFrom(minduration)
			}
			previous = is_trigger
		}
	}()
	return out
}

