package model

type ValidationStatus []Error

func (v ValidationStatus) IsValid() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (m *Message) Validate() ValidationStatus {
	v := ValidationStatus{}

	return v
}

// helper functions

func notNull(s interface{}) bool {
	if s == nil {
		return false
	}
	return true
}

func notEmpty(s string) bool {
	if len(s) == 0 {
		return false
	}
	return true
}

func notZero(s int) bool {
	if s == 0 {
		return false
	}
	return true
}
