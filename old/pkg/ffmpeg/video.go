package ffmpeg

import (
	"math"
	"strconv"
	"strings"
)

// VideoInfo contiene información procesada del video
type VideoInfo struct {
	Width         int     `json:"width"`
	Height        int     `json:"height"`
	Duration      float64 `json:"duration"`
	FPS           float64 `json:"fps"`
	Bitrate       int     `json:"bitrate"`
	Codec         string  `json:"codec"`
	AudioCodec    string  `json:"audio_codec"`
	AudioChannels int     `json:"audio_channels"`
	AudioBitrate  int     `json:"audio_bitrate"`
	Format        string  `json:"format"`
	Rotation      int     `json:"rotation"` // Rotación en grados (-90, 0, 90, 180, 270)
}

// ProbeResult es el resultado crudo de ffprobe
type ProbeResult struct {
	Format  Format   `json:"format"`
	Streams []Stream `json:"streams"`
}

// Format contiene información del contenedor
type Format struct {
	Duration string            `json:"duration"`
	Bitrate  string            `json:"bit_rate"`
	Format   string            `json:"format_name"`
	Tags     map[string]string `json:"tags"`
}

// Stream contiene información de cada stream (video/audio)
type Stream struct {
	Index         int               `json:"index"`
	CodecName     string            `json:"codec_name"`
	CodecType     string            `json:"codec_type"`
	Width         int               `json:"width"`
	Height        int               `json:"height"`
	RFrameRate    string            `json:"r_frame_rate"`
	AvgFrameRate  string            `json:"avg_frame_rate"`
	Duration      string            `json:"duration"`
	Bitrate       string            `json:"bit_rate"`
	Channels      int               `json:"channels"`
	ChannelLayout string            `json:"channel_layout"`
	Tags          map[string]string `json:"tags"`
	SideDataList  []SideData        `json:"side_data_list"` // Para displaymatrix
}

// SideData contiene metadata adicional como displaymatrix
type SideData struct {
	SideDataType  string `json:"side_data_type"`
	DisplayMatrix string `json:"displaymatrix"`
	Rotation      int    `json:"rotation"` // FFmpeg a veces incluye esto directamente
}

// parseFrameRate parsea el frame rate de FFmpeg (formato "30/1")
func parseFrameRate(frameRate string) float64 {
	if frameRate == "" || frameRate == "0/0" {
		return 0
	}

	if strings.Contains(frameRate, "/") {
		parts := strings.Split(frameRate, "/")
		if len(parts) == 2 {
			numerator, err1 := strconv.ParseFloat(parts[0], 64)
			denominator, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil && denominator != 0 {
				return numerator / denominator
			}
		}
	} else {
		if fps, err := strconv.ParseFloat(frameRate, 64); err == nil {
			return fps
		}
	}

	return 0
}

// GetVideoRotation extrae la rotación del video desde tags o displaymatrix
func (p *ProbeResult) GetVideoRotation() int {
	for _, stream := range p.Streams {
		if stream.CodecType != "video" {
			continue
		}

		// Método 1: Buscar en tags (algunos videos lo tienen aquí)
		if rotStr, ok := stream.Tags["rotate"]; ok {
			if rot, err := strconv.Atoi(rotStr); err == nil {
				return normalizeRotation(rot)
			}
		}

		// Método 2: Parsear side_data_list
		for _, sideData := range stream.SideDataList {
			// Si FFmpeg ya calculó la rotación
			if sideData.Rotation != 0 {
				return normalizeRotation(sideData.Rotation)
			}

			// Si no, parsear displaymatrix manualmente
			if sideData.DisplayMatrix != "" {
				return parseDisplayMatrix(sideData.DisplayMatrix)
			}
		}
	}
	return 0 // Sin rotación
}

// parseDisplayMatrix convierte la displaymatrix a grados de rotación
func parseDisplayMatrix(matrix string) int {
	// La matriz tiene este formato para -90°:
	// 00000000:     0       65536      0
	// 00000001: -65536      0          0
	// 00000002:     0    47185920  1073741824

	lines := strings.Split(strings.TrimSpace(matrix), "\n")
	if len(lines) < 2 {
		return 0
	}

	var a, b, c, d float64

	// Primera línea: extraer componentes [a, b, tx]
	parts1 := strings.Fields(lines[0])
	if len(parts1) >= 3 {
		a, _ = strconv.ParseFloat(parts1[1], 64)
		b, _ = strconv.ParseFloat(parts1[2], 64)
	}

	// Segunda línea: extraer componentes [c, d, ty]
	parts2 := strings.Fields(lines[1])
	if len(parts2) >= 3 {
		c, _ = strconv.ParseFloat(parts2[1], 64)
		d, _ = strconv.ParseFloat(parts2[2], 64)
	}

	// Normalizar (65536 es 2^16, el factor de escala usado por FFmpeg)
	scale := 65536.0
	a /= scale
	b /= scale
	c /= scale
	d /= scale

	// Calcular ángulo: atan2(-c, d) en radianes -> grados
	angle := math.Atan2(-c, d) * 180.0 / math.Pi

	return normalizeRotation(int(math.Round(angle)))
}

// normalizeRotation normaliza el ángulo al rango [-180, 180]
func normalizeRotation(angle int) int {
	// Normalizar a rango [0, 360)
	angle = angle % 360
	if angle < 0 {
		angle += 360
	}

	// Convertir a rango [-180, 180]
	if angle > 180 {
		angle -= 360
	}

	return angle
}

// GetActualDimensions retorna las dimensiones visuales del video
// teniendo en cuenta la rotación de metadata
func (v *VideoInfo) GetActualDimensions() (width, height int) {
	// Si hay rotación de 90° o 270°, las dimensiones están intercambiadas
	if v.Rotation == 90 || v.Rotation == -90 || v.Rotation == 270 || v.Rotation == -270 {
		return v.Height, v.Width
	}
	return v.Width, v.Height
}

// NeedsTranspose indica si el video necesita transpose para corregir rotación
func (v *VideoInfo) NeedsTranspose() bool {
	return v.Rotation != 0
}

// GetTransposeFilter retorna el filtro de transpose apropiado para la rotación
func (v *VideoInfo) GetTransposeFilter() string {
	switch v.Rotation {
	case 90, -270:
		return "transpose=1" // 90° clockwise
	case -90, 270:
		return "transpose=2" // 90° counter-clockwise
	case 180, -180:
		return "transpose=2,transpose=2" // 180°
	default:
		return ""
	}
}
