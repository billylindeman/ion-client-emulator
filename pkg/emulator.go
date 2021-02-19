package pkg

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/giongto35/cloud-game/v2/pkg/config"
	encoderConfig "github.com/giongto35/cloud-game/v2/pkg/config/encoder"
	"github.com/giongto35/cloud-game/v2/pkg/config/worker"
	"github.com/giongto35/cloud-game/v2/pkg/emulator/libretro/nanoarch"
	"github.com/giongto35/cloud-game/v2/pkg/encoder"
	"github.com/giongto35/cloud-game/v2/pkg/encoder/h264encoder"
	"github.com/giongto35/cloud-game/v2/pkg/encoder/opus"
	"github.com/giongto35/cloud-game/v2/pkg/games"
	itc "github.com/giongto35/cloud-game/v2/pkg/webrtc/interceptor"
	"github.com/lucsky/cuid"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

var (
	viewportWidth  int = 320
	viewportHeight int = 240
)

// getMetadata returns game info from a path
func getMetadata(path string) games.GameMetadata {
	name := filepath.Base(path)
	ext := filepath.Ext(name)

	return games.GameMetadata{
		Name: strings.TrimSuffix(name, ext),
		Type: ext[1:],
		Path: path,
	}
}

type emulatorProducer struct {
	cfg worker.Config

	videoChannel <-chan nanoarch.GameFrame
	audioChannel <-chan []int16
	inputChannel chan<- nanoarch.InputEvent

	videoTrack *webrtc.TrackLocalStaticSample
	audioTrack *webrtc.TrackLocalStaticSample

	ReTime  itc.ReTime
	running bool
}

func newEmulatorProducer(file string) *emulatorProducer {
	roomID := "ion-emulator"
	game := getMetadata(file)

	log.Printf("starting emulator for game %v: %#v", file, game)

	cfg := worker.Config{}
	config.LoadConfig(&cfg, "")
	emuName := cfg.Emulator.GetEmulatorByRom(game.Type)
	libretroConfig := cfg.Emulator.GetLibretroCoreConfig(emuName)

	store := nanoarch.Storage{
		Path:     cfg.Emulator.Storage,
		MainSave: roomID + ".dat",
	}

	inputChannel := make(chan nanoarch.InputEvent)
	emulator, videoChannel, audioChannel := nanoarch.Init(roomID, true, inputChannel, store, libretroConfig)
	emulator.SetViewport(viewportWidth, viewportHeight)

	stream := fmt.Sprintf("gst-screen-%v", cuid.New())
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/h264", ClockRate: 90000}, cuid.New(), stream)
	if err != nil {
		panic(err)
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000}, cuid.New(), stream)
	if err != nil {
		panic(err)
	}

	go func() {
		emulator.LoadMeta(game.Path)
		emulator.Start()
	}()

	return &emulatorProducer{
		cfg:          cfg,
		videoChannel: videoChannel,
		audioChannel: audioChannel,
		inputChannel: inputChannel,
		videoTrack:   videoTrack,
		audioTrack:   audioTrack,
		ReTime:       itc.ReTime{},
		running:      true,
	}
}

func (e *emulatorProducer) Start() {
	go e.startVideo()
	go e.startAudio(48000, e.cfg.Encoder.Audio)
}

func (e *emulatorProducer) Stop() {
	e.running = false
}

func (e *emulatorProducer) startVideo() {
	videoEnc, err := h264encoder.NewH264Encoder(viewportWidth, viewportHeight, 1)
	defer videoEnc.Stop()
	if err != nil {
		fmt.Println("error create new encoder", err)
		panic("could not create h264encoder")
	}

	encoderInput := videoEnc.GetInputChan()
	encoderOutput := videoEnc.GetOutputChan()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered when sent to close Image Channel")
			}
		}()

		for data := range encoderOutput {
			// log.Printf("h264frame encoded, writing sample duration: %v")
			e.ReTime.SetTimestamp(data.Timestamp)
			if err := e.videoTrack.WriteSample(media.Sample{Data: data.Data}); err != nil {
				panic(err)
			}
		}
	}()

	for frame := range e.videoChannel {
		if !e.running {
			log.Printf("emulatorProducer closed, shutting down encoder")
			return
		}
		if len(encoderInput) < cap(encoderInput) {
			// log.Printf("got frame: %d", frame.Timestamp)
			encoderInput <- encoder.InFrame{Image: frame.Image, Timestamp: frame.Timestamp}
		}
	}
}

func (e *emulatorProducer) startAudio(sampleRate int, audio encoderConfig.Audio) {
	sound, err := opus.NewEncoder(
		sampleRate,
		audio.Frequency,
		audio.Channels,
		opus.SampleBuffer(audio.Frame, sampleRate != audio.Frequency),
		// we use callback on full buffer in order to
		// send data to all the clients ASAP
		opus.CallbackOnFullBuffer(e.broadcastAudio),
	)
	if err != nil {
		log.Fatalf("error: cannot create audio encoder, %v", err)
	}
	log.Printf("OPUS: %v", sound.GetInfo())

	for samples := range e.audioChannel {
		sound.BufferWrite(samples)
	}

	log.Println("audio channel closed")
}

func (e *emulatorProducer) broadcastAudio(audio []byte) {
	audioDuration := time.Duration(e.cfg.Encoder.Audio.Frame) * time.Millisecond
	err := e.audioTrack.WriteSample(media.Sample{Data: audio, Duration: audioDuration})
	if err != nil {
		log.Println("Warn: Err write sample: ", err)
	}
}

//AudioTrack returns the audio track for the pipeline
func (e *emulatorProducer) AudioTrack() *webrtc.TrackLocalStaticSample {
	return e.audioTrack
}

//VideoTrack returns the video track for the pipeline
func (e *emulatorProducer) VideoTrack() *webrtc.TrackLocalStaticSample {
	return e.videoTrack
}
