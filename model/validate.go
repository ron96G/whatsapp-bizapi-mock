package model

type ValidationStatus []Error

func (v ValidationStatus) IsValid() bool {
	return len(v) == 0
}

func (m *Message) Validate() ValidationStatus {
	v := ValidationStatus{}

	return v
}

func (r *RegistrationRequest) Validate() ValidationStatus {
	v := ValidationStatus{}

	return v
}

// helper functions

func notNull(s interface{}) bool {
	return s == nil
}

func notEmpty(s string) bool {
	return len(s) == 0
}

func notZero(s int) bool {
	return s == 0
}
