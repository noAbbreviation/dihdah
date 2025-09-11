package commons

import (
	"fmt"
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/generators"
	"github.com/gopxl/beep/speaker"
)

type soundType int

const (
	ShortBeep soundType = iota
	LongBeep
	ShortDelay
)

const DefaultDitDuration = time.Millisecond * 60
const MorseSpaceIndicator = '_'

var SoundAssets map[soundType]*beep.Buffer

var AudioFormat = beep.Format{
	SampleRate:  24_000,
	NumChannels: 2,
	Precision:   2,
}

type audioBuffer struct {
	buffer beep.Buffer
}

func init() {
	initSoundAssets(DefaultDitDuration)
	speaker.Init(AudioFormat.SampleRate, AudioFormat.SampleRate.N(time.Second/10))
}

var currentDitDuration = time.Duration(0)

func initSoundAssets(ditDuration time.Duration) {
	if max(ditDuration, currentDitDuration)-min(ditDuration, currentDitDuration) < time.Millisecond*5 {
		return
	}

	currentDitDuration = ditDuration
	shortBeepSamples := AudioFormat.SampleRate.N(ditDuration)

	audioTone, err := generators.SineTone(AudioFormat.SampleRate, 1_000)
	tamedAudioTone := &effects.Volume{
		Streamer: audioTone,
		Base:     2,
		Volume:   -1,
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating the audioTone: %v\n", err)
		os.Exit(1)
	}

	sBeepBuffer := beep.NewBuffer(AudioFormat)
	sBeepBuffer.Append(beep.Take(shortBeepSamples, tamedAudioTone))

	lBeepBuffer := beep.NewBuffer(AudioFormat)
	lBeepBuffer.Append(beep.Take(shortBeepSamples*3, tamedAudioTone))

	sDelayBuffer := beep.NewBuffer(AudioFormat)
	sDelayBuffer.Append(generators.Silence(shortBeepSamples))

	SoundAssets = map[soundType]*beep.Buffer{
		ShortBeep:  sBeepBuffer,
		LongBeep:   lBeepBuffer,
		ShortDelay: sDelayBuffer,
	}
}

func MorseCharSound(str string, speed float64) beep.Streamer {
	buffer := beep.NewBuffer(AudioFormat)

	initSoundAssets(time.Duration(float64(DefaultDitDuration) / speed))
	resampledSounds := SoundAssets

	for _, r := range str {
		loopCount := 1

		var sound *beep.Buffer
		switch r {
		case '.':
			sound = resampledSounds[ShortBeep]
		case ',':
			sound = resampledSounds[LongBeep]
		case ' ', '-':
			sound = resampledSounds[ShortDelay]
			loopCount = 3
		case MorseSpaceIndicator:
			sound = resampledSounds[ShortDelay]

			// This is accounting for a space and delaySound before this
			loopCount = 4
		default:
			continue
		}

		for range loopCount {
			soundStreamer := sound.Streamer(0, sound.Len())
			buffer.Append(soundStreamer)
		}

		if r == '.' || r == ',' {
			delaySound := resampledSounds[ShortDelay]

			soundStreamer := delaySound.Streamer(0, delaySound.Len())
			buffer.Append(soundStreamer)
		}
	}

	return buffer.Streamer(0, buffer.Len())
}
