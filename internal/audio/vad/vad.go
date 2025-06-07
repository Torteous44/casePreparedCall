package vad

import (
	"encoding/binary"
	"fmt"
	"math"
)

// VAD represents a Voice Activity Detector
type VAD struct {
	energyThreshold    float64
	silenceCounter     int
	voiceCounter       int
	minVoiceDuration   int    // minimum consecutive frames for voice detection
	minSilenceDuration int    // minimum consecutive frames for silence detection
	buffer             []bool // circular buffer for smoothing detection
	bufferSize         int    // size of the circular buffer
	bufferIndex        int    // current position in the buffer
}

// NewVAD creates a new Voice Activity Detector
func NewVAD() *VAD {
	bufferSize := 5 // Adjust this value to change smoothing window
	return &VAD{
		energyThreshold:    1000.0, // Adjustable threshold for energy detection
		minVoiceDuration:   3,      // Need 3+ frames of voice to confirm
		minSilenceDuration: 5,      // Need 5+ frames of silence to confirm
		buffer:             make([]bool, bufferSize),
		bufferSize:         bufferSize,
		bufferIndex:        0,
	}
}

// DetectActivity detects voice activity in audio data
// Assumes 16-bit PCM audio data
func (v *VAD) DetectActivity(audioData []byte) (bool, error) {
	if len(audioData) < 2 {
		return false, fmt.Errorf("insufficient audio data")
	}

	// Convert bytes to 16-bit samples
	samples := make([]int16, len(audioData)/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(audioData[i*2 : i*2+2]))
	}

	// Calculate energy (RMS)
	energy := v.calculateEnergy(samples)

	// Determine if voice is present based on energy threshold
	hasVoice := energy > v.energyThreshold

	// Update buffer
	v.buffer[v.bufferIndex] = hasVoice
	v.bufferIndex = (v.bufferIndex + 1) % v.bufferSize

	// Count true values in buffer for smoothing
	trueCount := 0
	for _, val := range v.buffer {
		if val {
			trueCount++
		}
	}

	// Smooth decision based on buffer majority
	smoothedVoice := trueCount > v.bufferSize/2

	// Apply temporal smoothing to reduce false positives/negatives
	if smoothedVoice {
		v.voiceCounter++
		v.silenceCounter = 0
		// Confirm voice only if we have enough consecutive voice frames
		return v.voiceCounter >= v.minVoiceDuration, nil
	} else {
		v.silenceCounter++
		v.voiceCounter = 0
		// Confirm silence only if we have enough consecutive silence frames
		return v.silenceCounter < v.minSilenceDuration, nil
	}
}

// calculateEnergy computes the RMS energy of audio samples
func (v *VAD) calculateEnergy(samples []int16) float64 {
	if len(samples) == 0 {
		return 0.0
	}

	var sum float64
	for _, sample := range samples {
		sum += float64(sample) * float64(sample)
	}

	rms := math.Sqrt(sum / float64(len(samples)))
	return rms
}

// SetEnergyThreshold allows adjustment of the energy threshold
func (v *VAD) SetEnergyThreshold(threshold float64) {
	v.energyThreshold = threshold
}

// GetEnergyThreshold returns the current energy threshold
func (v *VAD) GetEnergyThreshold() float64 {
	return v.energyThreshold
}

// SetVoiceDuration sets the minimum voice duration for detection
func (v *VAD) SetVoiceDuration(frames int) {
	v.minVoiceDuration = frames
}

// SetSilenceDuration sets the minimum silence duration for detection
func (v *VAD) SetSilenceDuration(frames int) {
	v.minSilenceDuration = frames
}

// Reset resets the VAD internal counters
func (v *VAD) Reset() {
	v.voiceCounter = 0
	v.silenceCounter = 0
}
