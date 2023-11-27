package services

import (
	"os/exec"
	"runtime"
)

func PlayErrorSound() {
	if runtime.GOOS == "windows" {
		_ = exec.Command("powershell", "-c", "(New-Object Media.SoundPlayer 'System.Media.SystemSounds::Hand.Play();").Run()
	} else if runtime.GOOS == "darwin" {
		_ = exec.Command("afplay", "/System/Library/Sounds/Sosumi.aiff").Run()
	}
}
