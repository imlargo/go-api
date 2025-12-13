package enums

type Industry string

const (
	IndustryCreator      Industry = "creator"
	IndustryECommerce    Industry = "ecommerce"
	IndustryClipper      Industry = "clipper"
	IndustryCoaching     Industry = "coaching"
	IndustryAdultContent Industry = "adult_content"
)

func (i Industry) IsValid() bool {
	switch i {
	case IndustryCreator, IndustryECommerce, IndustryClipper, IndustryCoaching, IndustryAdultContent:
		return true
	default:
		return false
	}
}
