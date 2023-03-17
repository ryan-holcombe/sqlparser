package parse

type Validator interface {
	Valid() bool
}

func Validate(typ interface{}) bool {
	if valid, ok := typ.(Validator); ok {
		return valid.Valid()
	}

	return true
}
