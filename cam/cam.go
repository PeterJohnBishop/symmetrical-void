package cam

import (
	"github.com/pion/mediadevices"
	_ "github.com/pion/mediadevices/pkg/driver/camera"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

func StartRawCamera() (video.Reader, error) {
	stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			// Low resolution is crucial for ASCII art
			c.Width = prop.Int(160)
			c.Height = prop.Int(120)
		},
	})
	if err != nil {
		return nil, err
	}

	track := stream.GetVideoTracks()[0].(*mediadevices.VideoTrack)
	return track.NewReader(false), nil
}
