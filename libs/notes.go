package libs

import (
	"math"
)

// Limit derivative
func Slide(notes Audiostream, duration DurationstreamOrDuration) Audiostream {
	out:=MakeAudioStream()
	timestep:=Timestep.Seconds()
	go func (){
		current := 0.
		first := true
		for new := range notes {
			factor:=timestep/ReadDurationFrom(duration).Seconds()
			if first {
				current = new
				first = false
			} else if new > current {
				current += factor
			} else {
				current -= factor
			}
			out <- current
		}
	}()
	return out
}

// Transform note into matching frequency through a tempered scale
func TemperedScale(notes Audiostream,
					 base AudiostreamOrFloat,
					 offset AudiostreamOrFloat,
					 scale AudiostreamOrFloat) Audiostream {
	out := MakeAudioStream()
	go func (){
		for note := range notes {
			note += ReadFrom(offset)
			note /= ReadFrom(scale)
			frequency := math.Pow(2, note) * ReadFrom(base)
			out <- frequency
		}
	}()
	return out
}

// Take input an pass it through a pentatonic scale :
// - input = k => return kth pentatonic note
func Pentatonic(input Audiostream, use_major bool) Audiostream {
	minor := []int{0,2,4,7,9}
	major := []int{-3,0,2,4,7}
	scale := []int{}
	if use_major {
		scale = major
	} else {
		scale = minor
	}

	result := MakeAudioStream()

	go func(){
		for in := range input {
			note := int(math.Ceil(in))
			d,m:=Divmod(note,5)
			note_penta := scale[m]
			note_penta += d*12
			result<-float64(note_penta)
		}
	}()
	return result
}

