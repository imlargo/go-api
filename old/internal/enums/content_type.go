package enums

type ContentType string

const (
	ContentTypeVideo     ContentType = "video"
	ContentTypeStory     ContentType = "story"
	ContentTypeImage     ContentType = "image"
	ContentTypeSlideshow ContentType = "slideshow"
)

func (t ContentType) IsValid() bool {
	switch t {
	case ContentTypeVideo, ContentTypeStory, ContentTypeImage, ContentTypeSlideshow:
		return true
	default:
		return false
	}
}
