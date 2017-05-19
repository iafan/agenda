package agenda

// SerializableError is a helper function that returns either
// nil or string value of the provided error as interface{},
// which makes it serializable by e.g. json.Marshal
func SerializableError(err error) interface{} {
	if err != nil {
		return err.Error()
	}
	return nil
}
