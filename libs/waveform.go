package libs

import (
	"math"
	"math/rand"
	"time"
)

type Waveshape int64
const (
	Sinus Waveshape = iota
	Square
	Triangle
	Sawtooth
	Noise
)
type WaveshapestreamOrWaveshape interface{}

// Produce a soundwave of a specific frequency with a specific shape.
// Frequency is ignored when producing Noise
func Wave(frequencyc Audiostream, shapec WaveshapestreamOrWaveshape) Audiostream {
	result := MakeAudioStream()
	execute := func(){
		phase := 0.
		value:= 0.
		for frequency := range frequencyc {
			shape := ReadFromGeneric[Waveshape](shapec)

			switch shape {
			case Sinus:
				value = math.Sin(phase)
			case Square:
				if phase > math.Pi {
					value = -1
				} else {
					value = 1
				}
			case Triangle:
				value = 2. * phase / math.Pi - 1
				if value > 1 {
					value = 2 - value
				}
			case Sawtooth:
				value = phase / math.Pi - 1
			case Noise:
				value = rand.Float64()*2.-1.
			}

			result<-value

			phase += frequency * phasestep
			switch {
			case phase >= Tau:
					phase-=Tau
			case phase < 0:
					phase+=Tau
			}
		}
	}
	
	go execute()
	return result
}

//Compute ADSR at every impulsion. https://en.wikipedia.org/wiki/Envelope_(music)#ADSR
func Adsr(impulsion Signalstream,
	attack DurationstreamOrDuration,
	decay DurationstreamOrDuration,
	sustain AudiostreamOrFloat,
	release DurationstreamOrDuration) Audiostream {
	out := MakeAudioStream()
	timestep := Timestep.Seconds()

	go func(){
		phase:=0
		value:=0.
		for is_trigger := range impulsion {
			attack_step := timestep/ReadDurationFrom(attack).Seconds()
			decay_step := timestep/ReadDurationFrom(decay).Seconds()
			release_step := timestep/ReadDurationFrom(release).Seconds()
			sustain_level := ReadFrom(sustain)
			if is_trigger {
				switch phase {
					case 0:
						phase = 1
					case 1:
						value += attack_step
						if value > 1. {
							value = 1.
							phase = 2
						}
					case 2:
						value -= decay_step
						if value < sustain_level {
							value = sustain_level
							phase = 3
						}
					default:
					}
			} else {

				phase = 0
				value -= release_step
				if value < 0 {
					value = 0
				}
			}
			out <- value
	}
	}()
	return out
}

// Dampen input below max by adding distortion
func Saturate(input Audiostream, max AudiostreamOrFloat) Audiostream {
	out := MakeAudioStream()
	go func(){
		for i := range input {
			m := ReadFrom(max)
			out<-math.Sin(i/m)*m
		}
	}()
	return out
}

// Add reverb to `input`. Be careful not to have a too high `coeff`
func Reverb(input Audiostream,
			reverb time.Duration,
			coeff AudiostreamOrFloat) Audiostream {
	nb_samples := int(reverb / Timestep)
	window := make([]float64, nb_samples)
	out := MakeAudioStream()
	go func(){
		i := 0
		for in := range input {
			window[i] = in + window[i] * ReadFrom(coeff)
			out<-window[i]
			i++
			i %= nb_samples
		}
	}()
	return out
}

// Return a fadein shape while `dummy` is active,
// starting at `position`, and with a `length` duration
func Fadein(dummy Signalstream,
			position time.Duration,
			length time.Duration) Audiostream {
	out:=MakeAudioStream()
	go func(){
		var t time.Duration
		for range dummy {
			if t < position {
				out<-0
			} else if t > position + length {
				out<-1
			} else {
				out<-(t-position).Seconds()/length.Seconds()
			}

			t += Timestep
		}
	}()
	return out
}

// Return a fadeout shape while `dummy` is active,
// ending at `position`, and with a `length` duration
func Fadeout(dummy Signalstream,
	position time.Duration,
	length time.Duration) Audiostream {
	return Join(Vca(Fadein(dummy,position-length,length),-1.), 1.)
}