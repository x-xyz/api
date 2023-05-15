package ptr

// String return a pointer to the input value
func String(value string) *string {
	return &value
}

// Int return a pointer to the input value
func Int(value int) *int {
	return &value
}

// Int32 return a pointer to the input value
func Int32(value int32) *int32 {
	return &value
}

// Int64 return a pointer to the input value
func Int64(value int64) *int64 {
	return &value
}

// Float32 return a pointer to the input value
func Float32(value float32) *float32 {
	return &value
}

// Float64 return a pointer to the input value
func Float64(value float64) *float64 {
	return &value
}

// Bool return a pointer to the input value
func Bool(value bool) *bool {
	return &value
}
