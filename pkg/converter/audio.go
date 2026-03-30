package converter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/g3n/engine/audio"
	"github.com/g3n/engine/core"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
)

func convertSound(snd *node.Sound, parent *core.Node, baseDir string) {
	if snd.Source == nil || len(snd.Source.URL) == 0 {
		return
	}

	audioPath := resolveAudioPath(snd.Source.URL, baseDir)
	if audioPath == "" {
		return
	}

	player, err := audio.NewPlayer(audioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "audio: failed to load %s: %v\n", audioPath, err)
		return
	}

	player.SetLooping(snd.Source.Loop)
	if snd.Source.Pitch > 0 {
		player.SetPitch(float32(snd.Source.Pitch))
	}
	player.SetGain(float32(snd.Intensity))

	if snd.Spatialize {
		player.SetPosition(
			float32(snd.Location.X),
			float32(snd.Location.Y),
			float32(snd.Location.Z),
		)
		player.SetRolloffFactor(1.0)
		if snd.MaxFront > 0 {
			player.SetMaxGain(float32(snd.MaxFront))
		}
		if snd.MinFront > 0 {
			player.SetMinGain(float32(snd.MinFront))
		}
		if snd.Direction.X != 0 || snd.Direction.Y != 0 || snd.Direction.Z != 0 {
			player.SetInnerCone(90)
			player.SetOuterCone(180)
		}
	}

	player.SetName(snd.GetName())
	parent.Add(player)

	if snd.Source.StartTime == 0 {
		if err := player.Play(); err != nil {
			fmt.Fprintf(os.Stderr, "audio: play error: %v\n", err)
		}
	}
}

func resolveAudioPath(urls []string, baseDir string) string {
	for _, url := range urls {
		path := url
		if !filepath.IsAbs(path) && baseDir != "" {
			path = filepath.Join(baseDir, url)
		}
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}
