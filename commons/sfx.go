package commons

import (
	"fmt"
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

type soundType int

const (
	ShortBeep soundType = iota
	LongBeep
	ShortDelay
)

var (
	AudioFormat beep.Format
	SoundAssets = map[soundType]*beep.Buffer{}
)

type audioBuffer struct {
	buffer beep.Buffer
}

func init() {
	soundFilePrefixes := map[soundType]string{
		ShortBeep:  "short-beep",
		LongBeep:   "long-beep",
		ShortDelay: "short-delay",
	}

	for soundType, filePrefix := range soundFilePrefixes {
		fileName := fmt.Sprintf("./assets/sfx_%v.mp3", filePrefix)

		file, err := os.Open(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %v: %v", fileName, err)
			os.Exit(1)
		}

		var streamer beep.StreamSeekCloser
		streamer, AudioFormat, err = mp3.Decode(file)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding %v: %v", fileName, err)
			os.Exit(1)
		}

		buffer := beep.NewBuffer(AudioFormat)
		buffer.Append(streamer)
		streamer.Close()

		SoundAssets[soundType] = buffer
	}

	speaker.Init(AudioFormat.SampleRate, AudioFormat.SampleRate.N(time.Second/10))
}

func MorseCharSound(str string) beep.Streamer {
	buffer := beep.NewBuffer(AudioFormat)

	for _, r := range str {
		loopCount := 1

		var sound *beep.Buffer
		switch r {
		case '.':
			sound = SoundAssets[ShortBeep]
		case ',':
			sound = SoundAssets[LongBeep]
		default:
			sound = SoundAssets[ShortDelay]
			loopCount = 3
		}

		for range loopCount {
			soundStreamer := sound.Streamer(0, sound.Len())
			buffer.Append(soundStreamer)
		}
	}

	return buffer.Streamer(0, buffer.Len())
}
