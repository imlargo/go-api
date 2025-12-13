package content

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/nicolailuther/butter/pkg/apiclient"
)

type shotstackService struct {
	apiClient *apiclient.ApiClient
}

func NewShotstackMediaService(shotstackApiKey string) MediaService {
	client := apiclient.NewClient("https://api.shotstack.io", time.Minute*2, map[string]string{
		"Accept":    "application/json",
		"x-api-key": shotstackApiKey,
	})

	return &shotstackService{
		apiClient: client,
	}
}

type shotstackResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Response struct {
		Status    string `json:"status"`
		ID        string `json:"id"`
		Owner     string `json:"owner"`
		URL       string `json:"url"`
		Thumbnail string `json:"thumbnail"`
	} `json:"response"`
}

type shotstackProbeResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Response struct {
		Metadata struct {
			Streams []struct {
				Index     int    `json:"index"`
				CodecType string `json:"codec_type"`
				Tags      struct {
					CreationTime string `json:"creation_time"`
				} `json:"tags"`
				SideDataList []struct {
					SideDataType string `json:"side_data_type"`
					Rotation     int    `json:"rotation"`
				} `json:"side_data_list,omitempty"`
			} `json:"streams"`
			Format struct {
				Filename string `json:"filename"`
				Duration string `json:"duration"`
			} `json:"format"`
		} `json:"metadata"`
	} `json:"response"`
}

func (s *shotstackService) GenerateThumbnail(fileUrl string) (string, error) {

	metadata, err := s.getVideoMetadata(fileUrl)
	if err != nil {
		return "", fmt.Errorf("error getting video metadata: %w", err)
	}

	if err := IsValidVideo(*metadata); err != nil {
		return "", err
	}

	rotate, angle := GetRotationFromProbe(*metadata)

	endpoint := "/edit/v1/render"
	var result shotstackResponse
	err = s.apiClient.Post(endpoint, getThumbnailOptions(fileUrl, rotate, angle), &result)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("error in Shotstack response: %s", result.Message)
	}

	isDone := false
	renderingID := result.Response.ID

	for !isDone {
		taskUrl := fmt.Sprintf("%s/%s", endpoint, renderingID)
		s.apiClient.Get(taskUrl, &result)
		if !result.Success {
			return "", fmt.Errorf("error checking Shotstack rendering status: %s", result.Message)
		}

		isDone = result.Response.Status == "done" || result.Response.Status == "error"
		time.Sleep(1 * time.Second)
	}

	return result.Response.URL, nil
}

func (s *shotstackService) RenderVideoWithThumbnail(videoUrl string, textOverlay string, useMirror bool, useOverlay bool) (string, string, error) {
	metadata, err := s.getVideoMetadata(videoUrl)
	if err != nil {
		return "", "", fmt.Errorf("error getting video metadata: %w", err)
	}

	if err := IsValidVideo(*metadata); err != nil {
		return "", "", err
	}

	rotate, angle := GetRotationFromProbe(*metadata)

	endpoint := "/edit/v1/render"
	var result shotstackResponse
	err = s.apiClient.Post(endpoint, generateApiOptions(videoUrl, textOverlay, useMirror, rotate, angle), &result)
	if err != nil {
		return "", "", err
	}

	isDone := false
	renderingID := result.Response.ID

	for !isDone {
		taskUrl := fmt.Sprintf("%s/%s", endpoint, renderingID)
		s.apiClient.Get(taskUrl, &result)
		if !result.Success {
			return "", "", fmt.Errorf("error checking Shotstack rendering status: %s", result.Message)
		}

		isDone = result.Response.Status == "done" || result.Response.Status == "error"
		time.Sleep(1 * time.Second)
	}

	return result.Response.URL, result.Response.Thumbnail, nil
}

func (s *shotstackService) RenderVideoWithThumbnailV2(videoID int, textOverlay string, useMirror bool, useOverlay bool, isMain bool) (*ButterRepurposerTaskResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *shotstackService) GenerateThumbnailV2(fileID int) (*ButterRepurposerTaskResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *shotstackService) getVideoMetadata(input string) (*shotstackProbeResponse, error) {
	encodedUrl := url.QueryEscape(input)
	var result shotstackProbeResponse
	err := s.apiClient.Get(fmt.Sprintf("/edit/v1/probe/%s", encodedUrl), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func createCaptionTrack(overlay string) map[string]interface{} {
	estimatedCharsPerLine := int(math.Floor(float64(MaxTextWidth) / float64(ApproxCharWidth)))
	visualLineCount := int(math.Ceil(float64(len(overlay)) / float64(estimatedCharsPerLine)))
	actualLineCount := len(strings.Split(overlay, "\n"))
	lineCount := int(math.Max(float64(visualLineCount), float64(actualLineCount)))

	adjustedFontSize := int(math.Round(BaseFontSize * (1 - FontReductionFactor*float64(lineCount-1))))
	adjustedLineSize := BaseLineHeight * (1 - LineReductionFactor*float64(lineCount-1))

	fontSize := int(math.Max(float64(adjustedFontSize), 40))
	lineHeight := math.Max(adjustedLineSize, 0.7)

	return map[string]interface{}{
		"clips": []map[string]interface{}{
			{
				"fit":   "none",
				"scale": 1,
				"asset": map[string]interface{}{
					"type": "text",
					"text": overlay,
					"alignment": map[string]interface{}{
						"horizontal": "center",
						"vertical":   "center",
					},
					"font": map[string]interface{}{
						"color":      "#ffffff",
						"family":     "Montserrat ExtraBold",
						"size":       fontSize,
						"lineHeight": lineHeight,
					},
					"width":  MaxTextWidth,
					"height": 1920,
					"stroke": map[string]interface{}{
						"width": 3,
						"color": "#000000",
					},
				},
				"start":    0,
				"length":   "end",
				"position": "center",
			},
		},
	}
}

func getFlipTransformation() map[string]interface{} {
	return map[string]interface{}{
		"flip": map[string]interface{}{
			"horizontal": true,
			"vertical":   false,
		},
	}
}

func createVideoTrack(videoSource string, useMirror bool, rotate bool, angle int) map[string]interface{} {
	clip := map[string]interface{}{
		"asset": map[string]interface{}{
			"type":      "video",
			"src":       videoSource,
			"transcode": true,
		},
		"start":  0,
		"length": "auto",
	}

	if useMirror {
		transform := map[string]interface{}{
			"flip": map[string]interface{}{
				"horizontal": true,
			},
		}

		if rotate {
			transform = map[string]interface{}{
				"flip": map[string]interface{}{
					"horizontal": true,
				},
				"rotate": map[string]interface{}{
					"angle": -angle,
				},
			}
		}

		clip["transform"] = transform
	} else if rotate {
		clip["transform"] = map[string]interface{}{
			"rotate": map[string]interface{}{
				"angle": angle,
			},
		}
	}

	return map[string]interface{}{
		"clips": []map[string]interface{}{clip},
	}
}

func generateApiOptions(videoSource, caption string, useMirror bool, rotate bool, angle int) map[string]interface{} {
	tracks := []map[string]interface{}{}

	if caption != "" {
		tracks = append(tracks, createCaptionTrack(caption))
	}

	tracks = append(tracks, createVideoTrack(videoSource, useMirror, rotate, angle))

	baseConfig := map[string]interface{}{
		"timeline": map[string]interface{}{
			"background": "#000000",
			"tracks":     tracks,
		},
		"output": map[string]interface{}{
			"format": "mp4",
			"fps":    30,
			"size": map[string]interface{}{
				"width":  1080,
				"height": 1920,
			},
			"thumbnail": map[string]interface{}{
				"capture": 1,
				"scale":   1,
			},
		},
	}

	return baseConfig
}

func getThumbnailOptions(contentSource string, rotate bool, angle int) map[string]interface{} {
	clip := map[string]interface{}{
		"length": "auto",
		"start":  0,
		"asset": map[string]interface{}{
			"type":      "video",
			"src":       contentSource,
			"transcode": true,
		},
	}

	if rotate {
		clip["transform"] = map[string]interface{}{
			"rotate": map[string]interface{}{
				"angle": angle,
			},
		}
	}

	return map[string]interface{}{
		"output": map[string]interface{}{
			"format":  "png",
			"fps":     25,
			"scaleTo": "preview",
			"size": map[string]interface{}{
				"width":  1080,
				"height": 1920,
			},
		},
		"timeline": map[string]interface{}{
			"background": "#FFFFFF",
			"tracks": []map[string]interface{}{
				{
					"clips": []map[string]interface{}{
						clip,
					},
				},
			},
		},
	}
}

func GetRotationFromProbe(probe shotstackProbeResponse) (bool, int) {
	for _, stream := range probe.Response.Metadata.Streams {
		if stream.CodecType == "video" {
			for _, sideData := range stream.SideDataList {
				if sideData.SideDataType == "Display Matrix" && sideData.Rotation != 0 {
					return true, -sideData.Rotation
				}
			}
		}
	}
	return false, 0
}

func IsValidVideo(probe shotstackProbeResponse) error {
	if !probe.Success {
		return fmt.Errorf("error in Shotstack response: %s", probe.Message)
	}

	for _, stream := range probe.Response.Metadata.Streams {
		if stream.CodecType == "video" {
			return nil
		}
	}

	return errors.New("no video stream found in media file")
}
