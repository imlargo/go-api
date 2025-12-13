package enums

type Goal string

const (
	GoalDfy Goal = "dfy" // Done For You Marketing
	GoalCrm Goal = "crm" // Customer Relationship Management
)

func (g Goal) IsValid() bool {
	switch g {
	case GoalDfy, GoalCrm:
		return true
	default:
		return false
	}
}
