package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructToMap(t *testing.T) {
	t.Run("Struct", func(t *testing.T) {
		asserts := assert.New(t)

		m, err :=
			StructToMap(
				struct {
					Key1 string
					Key2 string      `json:"key2"`
					Key3 string      `json:"key3"`
					Key4 string      `json:"key4"`
					Key5 interface{} `json:"key5"`
					Key6 interface{} `json:"key6,omitempty"`
					Key7 interface{} `json:"-"`
				}{
					Key1: "value1",
					Key2: "value2",
					Key3: "",
					Key7: "Value7",
				},
				"json")
		if asserts.NoError(err) == false {
			return
		}

		if asserts.Equal("value1", m["Key1"]) == false {
			return
		}
		if asserts.Equal("value2", m["key2"]) == false {
			return
		}
		if asserts.Equal("", m["key3"]) == false {
			return
		}
		if asserts.Equal("", m["key4"]) == false {
			return
		}
		v5, ok := m["key5"]
		if asserts.Nil(v5) == false {
			return
		}
		if asserts.True(ok) == false {
			return
		}
		v6, ok := m["key6"]
		if asserts.Nil(v6) == false {
			return
		}
		if asserts.False(ok) == false {
			return
		}
		v7, ok := m["key7"]
		if asserts.Nil(v7) == false {
			return
		}
		if asserts.False(ok) == false {
			return
		}
	})

	t.Run("StructPtr", func(t *testing.T) {
		asserts := assert.New(t)

		m, err :=
			StructToMap(
				&struct {
					Key1 string
					Key2 string `json:"key2"`
				}{
					Key1: "value1",
					Key2: "value2",
				},
				"json")
		if asserts.NoError(err) == false {
			return
		}

		if asserts.Equal("value1", m["Key1"]) == false {
			return
		}
		if asserts.Equal("value2", m["key2"]) == false {
			return
		}
	})

	t.Run("NonStruct", func(t *testing.T) {
		asserts := assert.New(t)

		_, err :=
			StructToMap(
				[]struct{}{
					{},
					{},
				},
				"json")
		asserts.EqualError(err, "input value([]struct {}) must be a struct or pointer to struct")
	})
}
