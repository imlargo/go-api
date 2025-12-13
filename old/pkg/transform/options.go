package transform

import "github.com/nicolailuther/butter/pkg/ffmpeg"

type TransformationOptions struct {
	VideoInfo   *ffmpeg.VideoInfo
	TextOverlay string
	UseOverlay  bool
	UseMirror   bool
	IsMain      bool
}
