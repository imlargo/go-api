package enums

type Platform string

const (
	PlatformInstagram Platform = "instagram"
	PlatformTikTok    Platform = "tiktok"
	PlatformThreads   Platform = "threads"
	PlatformX         Platform = "x"
)

func (p Platform) IsValid() bool {
	switch p {
	case PlatformInstagram, PlatformTikTok, PlatformThreads, PlatformX:
		return true
	default:
		return false
	}
}
