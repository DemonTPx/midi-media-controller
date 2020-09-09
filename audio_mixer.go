package main

import (
	"github.com/mafik/pulseaudio"
	"log"
)

type AudioMixer struct {
	client               *pulseaudio.Client
	volumeChangeCallback func(volume float32)
	volume               float32
}

func NewAudioMixer() *AudioMixer {
	return &AudioMixer{}
}

func (m *AudioMixer) Init() error {
	client, err := pulseaudio.NewClient()

	if err != nil {
		return err
	}

	m.client = client
	m.volume, _ = m.client.Volume()

	updates, err := m.client.Updates()
	if err != nil {
		return err
	}

	go func() {
		for range updates {
			volume, _ := m.client.Volume()
			if m.volume != volume {
				m.volume = volume
				if m.volumeChangeCallback != nil {
					m.volumeChangeCallback(m.volume)
				}
			}
		}
	}()

	return nil
}

func (m *AudioMixer) SetVolume(volume float32) {
	err := m.client.SetVolume(volume)

	if err != nil {
		log.Printf("error while setting volume %v", err)
	}
}

func (m *AudioMixer) SetOnVolumeChangeCallback(callback func(volume float32)) {
	m.volumeChangeCallback = callback
}
