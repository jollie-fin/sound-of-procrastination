package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cryptix/wav"
	libs "github.com/jollie-fin/sound-of-procrastination/libs"
)

func listen(filename string, duration time.Duration, in libs.Audiostream) error {
	var length = int32(duration.Seconds()*libs.Samplerate)

	wf := wav.File{
		SampleRate: libs.Samplerate,
		Channels: 1,
		Duration: duration,
		SignificantBits: 32,
	}
	dat, err := os.Create(filename)
	if err != nil {
		return err
	}	
	wr, err := wf.NewWriter(dat)
	if err != nil {
		return err
	}
	defer wr.Close()
	
	for i := 0; i < int(length);i++ {
		wr.WriteInt32(int32(<-in* (1<<31-1)))
	}
	return nil
}

// Synthesize a pumkatchaka drumline
func pumKaTchaKa(tempo libs.Signalstream) libs.Audiostream {
	// Split tempo
	tempos := libs.SplitIntoSlices(tempo, 4)
	drumtrigger1, drumtrigger2, consttrigger1, consttrigger2 := tempos[0], tempos[1], tempos[2], tempos[3]

	// bassdrum, snare, hihat
	drumpattern := [][]float64{
		{1. ,0. ,1. },
		{0. ,0. ,.9 },
		{0. ,1. ,.8 },
		{0. ,0. ,.6 },
		{.8 ,0. ,.7 },
		{0. ,0. ,.8 },
		{0. ,1. ,.9 },
		{0. ,0. ,.8 }}
	drums := libs.FilterDirac(drumtrigger1, drumpattern)
	drumsvelocity := libs.StepwiseMultiples(drumtrigger2, drumpattern)


	const triggerwidth = 50 * time.Millisecond

	// bassdrum
	slimtriggerbassdrum1,slimtriggerbassdrum2 := libs.Split(drums[0])
	triggerbassdrum	:= libs.DiracToImpulsion(slimtriggerbassdrum1,triggerwidth)
	velocitybassdrum := drumsvelocity[0]
	envelopebassdrum := libs.Adsr(
		triggerbassdrum,
		10*time.Millisecond, 40*time.Millisecond,
		.1, 300*time.Millisecond)
	bassdrum := libs.Saturate(
		libs.Vca(
			libs.Wave(
				libs.Stepwise(slimtriggerbassdrum2,[]float64{150,145}),
				libs.Sinus),
			envelopebassdrum,
			velocitybassdrum,
			1.7), .5)

	// hihat
	triggerhihat	:= libs.DiracToImpulsion(drums[2],triggerwidth)
	velocityhihat	 := drumsvelocity[2]
	
	envelopehihat := libs.Adsr(triggerhihat,
					10*time.Millisecond, 60*time.Millisecond,
					.1, 100*time.Millisecond)
	hihat := libs.Vca(libs.Wave(libs.Constant(consttrigger1, 1.), libs.Noise),
					  envelopehihat,
					  velocityhihat,
					  .3)

	// snare
	triggersnare	:= libs.DiracToImpulsion(drums[1],triggerwidth)
	velocitysnare	 := drumsvelocity[1]			  
	envelopesnare := libs.Adsr(triggersnare,
						10*time.Millisecond, 40*time.Millisecond,
						.1, 40*time.Millisecond)
	snare := libs.Vca(libs.Wave(libs.Constant(consttrigger2, 700.), libs.Square),
					  envelopesnare,
					  velocitysnare,
				      .2)

	return libs.Join(bassdrum, snare, hihat)
}

// Synthesize a melodic line
func synthSolo(tempo libs.Signalstream, vibrato_duration time.Duration) libs.Audiostream {
	tempos := libs.SplitIntoSlices(tempo,5)
	temponotes, tempovolume, temposhape, dummy, tempovibrato :=
		tempos[0], tempos[1], libs.DiracOnceIn(tempos[2],2), tempos[3], libs.DiracOnceIn(tempos[4],2)

	// vibrato
	vibrato_frequency := 1/vibrato_duration.Seconds()/2
	vibration_data := []float64{
		0,0,0,0,
		0,.1,2,0,
		0,0,0,0,
		0,.1,.5,.5,
	}
	vibrato_amount := libs.Slide(libs.Stepwise(tempovibrato, vibration_data),vibrato_duration*2)
	vibrato := libs.Vca(libs.Wave(libs.Constant(dummy, vibrato_frequency), libs.Sinus), vibrato_amount)

	// frequency
	note_data := []float64{
		0,0,7,7,7,7,7,7,
		0,-1,8,8,8,8,8,-1,
		-1,-1,8,8,8,8,8,8,
		0,-1,7,7,7,7,7,0}
	notes := libs.Stepwise(temponotes, note_data)
	slidy_notes := libs.Join(libs.Slide(notes, time.Second/2/12), vibrato)
	slidy_frequency := libs.TemperedScale(slidy_notes, 220, 6, 12)

	// volume
	volume_data := []float64{
		.6,.6, 3, 3, 3, 3,3,3,
	    .6,.6,.8,.8,.8,.8,0,0,
		.6,.6,5,5,5,5,5,5,
		.6,.6,.8,.8,.8,.8,0,0}
	volume := libs.Stepwise(tempovolume, volume_data)
	slidy_volume := libs.Slide(volume, time.Second/2/6)

	shape_data := []libs.Waveshape{
		libs.Sinus, libs.Triangle, libs.Triangle, libs.Triangle,
		libs.Sinus, libs.Triangle, libs.Triangle, libs.Triangle,
		libs.Sinus, libs.Sinus, libs.Sinus, libs.Sinus,
		libs.Sinus, libs.Sinus, libs.Sinus, libs.Sinus,
	}
	shapes := libs.Stepwise(temposhape, shape_data)

	// synth

	// shapes1,shapes2 := libs.Split(shapes)
	// slidy_frequency1, slidy_frequency2 := libs.Split(slidy_frequency)
	// slidy_volume1, slidy_volume2 := libs.Split(slidy_volume)

	// base := libs.Vca(libs.Wave(slidy_frequency1, shapes1),slidy_volume1)
	// saturated_base := libs.Saturate(base, 1)
	// harmony := libs.Vca(libs.Wave(libs.Vca(slidy_frequency2, 2.*4./3.), shapes2),slidy_volume2)
	// saturated_harmony := libs.Saturate(harmony, 1)
	base := libs.Vca(libs.Wave(slidy_frequency, shapes),slidy_volume)
	saturated_base := libs.Saturate(base, 1)

	reverbed := libs.Reverb(libs.Join(saturated_base/*, libs.Vca(saturated_harmony, .3)*/), time.Minute/90/4,.45)
	return reverbed
}

func main() {
	filename := os.Args[1]

	/* It is good practice to have a mean to stop goroutine
	   Here, quit is a channel that, if written to, will trigger a return
	   within TriggerEvery internal Routine
	   Every other routine depends indirectly from tempo, so they will
	   all exit ordingly */
	quit := make(chan bool)
	defer func(){quit <- true}()
	tempo_duration := time.Minute/90/8
	tempo := libs.DiracEvery(tempo_duration, quit)

	/* a Stream can only be consumed by one receiver, so let's split it */
	tempos := libs.SplitIntoSlices(tempo,4)
	tempodrum := libs.DiracOnceIn(tempos[0], 4)
	tempomain := libs.DiracOnceIn(tempos[1], 4)
	tempofadein := tempos[2]
	tempofadeout := tempos[3]

	/* Music lines */
	drumline := pumKaTchaKa(tempodrum)
	mainline := synthSolo(tempomain, tempo_duration)

	totallength := time.Minute/90*15+time.Second/3
	fadein := libs.Fadein(tempofadein, 0, 200*time.Millisecond)
	fadeout := libs.Fadeout(tempofadeout, totallength, 400*time.Millisecond)
	err := listen(
		filename,
		totallength,
		libs.Vca(
			libs.Join(
				libs.Vca(mainline, .5),
				libs.Vca(drumline, 1.)),
			.1,
			fadein,
			fadeout))
	if err != nil {
		fmt.Printf("Received error %s\n", err)
	}
}
