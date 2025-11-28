package main

import (
	"bytes"
	"io"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

// SoundEvent represents a sound event type
type SoundEvent int

const (
	SoundEventDot SoundEvent = iota
	SoundEventPowerPellet
	SoundEventEatGhost
	SoundEventGameOver
	SoundEventHighScore
)

// SoundManager manages sound effects
type SoundManager struct {
	enabled          bool
	sampleRate       beep.SampleRate
	dotSound         *beep.Buffer
	powerPelletSound *beep.Buffer
	eatGhostSound    *beep.Buffer
	gameOverSound    *beep.Buffer
	highScoreSound   *beep.Buffer
}

// NewSoundManager creates a new sound manager
func NewSoundManager() *SoundManager {
	sm := &SoundManager{
		enabled:    true,
		sampleRate: beep.SampleRate(44100),
	}

	// Initialize speaker
	speaker.Init(sm.sampleRate, sm.sampleRate.N(time.Second/10))

	// Load sounds from bundled resources (if available)
	// For now, we'll use placeholder sounds
	sm.loadSounds()

	return sm
}

// loadSounds loads all sound effects
func (sm *SoundManager) loadSounds() {
	// Try to load zip.mp3 for dot collection
	if resourceZipMp3Data != nil {
		sm.dotSound = sm.loadSound(resourceZipMp3Data)
	}

	// Try to load wheeHoo.mp3 for power pellet
	if resourceWheeHooMp3Data != nil {
		sm.powerPelletSound = sm.loadSound(resourceWheeHooMp3Data)
	}

	// Try to load boing.mp3 for eating ghost
	if resourceBoingMp3Data != nil {
		sm.eatGhostSound = sm.loadSound(resourceBoingMp3Data)
	}

	// Try to load uhOh.mp3 for game over
	if resourceUhOhMp3Data != nil {
		sm.gameOverSound = sm.loadSound(resourceUhOhMp3Data)
	}

	// Use wheeHoo.mp3 for high score
	if resourceWheeHooMp3Data != nil {
		sm.highScoreSound = sm.loadSound(resourceWheeHooMp3Data)
	}
}

// loadSound loads a sound from byte data
func (sm *SoundManager) loadSound(data []byte) *beep.Buffer {
	if len(data) == 0 {
		return nil
	}

	streamer, format, err := mp3.Decode(io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return nil
	}

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	streamer.Close()

	return buffer
}

// PlaySound plays a sound effect
func (sm *SoundManager) PlaySound(event SoundEvent) {
	if !sm.enabled {
		return
	}

	var buffer *beep.Buffer
	switch event {
	case SoundEventDot:
		buffer = sm.dotSound
	case SoundEventPowerPellet:
		buffer = sm.powerPelletSound
	case SoundEventEatGhost:
		buffer = sm.eatGhostSound
	case SoundEventGameOver:
		buffer = sm.gameOverSound
	case SoundEventHighScore:
		buffer = sm.highScoreSound
	}

	if buffer != nil {
		sound := buffer.Streamer(0, buffer.Len())
		speaker.Play(sound)
	}
}

// Toggle toggles sound on/off
func (sm *SoundManager) Toggle() {
	sm.enabled = !sm.enabled
}

// SetEnabled sets the sound enabled state
func (sm *SoundManager) SetEnabled(enabled bool) {
	sm.enabled = enabled
}

// IsEnabled returns whether sound is enabled
func (sm *SoundManager) IsEnabled() bool {
	return sm.enabled
}
