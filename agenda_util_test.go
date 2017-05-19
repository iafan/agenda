package agenda

import (
	"encoding/json"
	"errors"
	"testing"
)

// TestSerializableError is a traditional (non agenda-based) test
// that tests SerializableError function
func TestSerializableError(t *testing.T) {
	bytes, err := json.Marshal(SerializableError(nil))
	if err != nil {
		t.Error(err.Error())
	}
	if string(bytes) != "null" {
		t.Errorf("Expected 'null', got '%s'", string(bytes))
	}

	bytes, err = json.Marshal(SerializableError(errors.New("test")))
	if err != nil {
		t.Error(err.Error())
	}
	if string(bytes) != `"test"` {
		t.Errorf("Expected 'test', got '%s'", string(bytes))
	}
}
