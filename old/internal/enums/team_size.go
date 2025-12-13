package enums

type TeamSize string

const (
	TeamSizeJustMe      TeamSize = "just_me"
	TeamSizeTwoToFive   TeamSize = "two_to_five"
	TeamSizeSixToTwenty TeamSize = "six_to_twenty"
	TeamSizeTwentyPlus  TeamSize = "twenty_plus"
)

func (t TeamSize) IsValid() bool {
	switch t {
	case TeamSizeJustMe, TeamSizeTwoToFive, TeamSizeSixToTwenty, TeamSizeTwentyPlus:
		return true
	default:
		return false
	}
}
