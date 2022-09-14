package common

func StringPtr(str string) *string {
	return &str
}

func IntPtr(i int) *int {
	return &i
}

func BoolPtr(b bool) *bool {
	return &b
}
