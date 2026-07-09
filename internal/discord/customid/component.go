package customid

type ComponentInput struct {
	CustomID string
	Values   []string
}

func ParseComponentInput(input ComponentInput) (ID, error) {
	return ParseComponent(input.CustomID)
}
