package logger

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
)

func MapMarshalerFuncString(m map[string]string) zapcore.ObjectMarshalerFunc {
	return zapcore.ObjectMarshalerFunc(func(inner zapcore.ObjectEncoder) error {
		for k, v := range m {
			inner.AddString(k, v)
		}
		return nil
	})
}

func MapMarshalerFuncStringp(m map[string]*string) zapcore.ObjectMarshalerFunc {
	return zapcore.ObjectMarshalerFunc(func(inner zapcore.ObjectEncoder) error {
		for k, v := range m {
			inner.AddString(k, *v)
		}
		return nil
	})
}

func MapMarshalerFuncAny(m map[string]interface{}) zapcore.ObjectMarshalerFunc {
	return zapcore.ObjectMarshalerFunc(func(inner zapcore.ObjectEncoder) error {
		for key, value := range m {
			switch val := value.(type) {
			case zapcore.ObjectMarshaler:
				if err := inner.AddObject(key, val); err != nil {
					return fmt.Errorf("failed to add zapcore.ObjectMarshaler : %w", err)
				}
			case zapcore.ArrayMarshaler:
				if err := inner.AddArray(key, val); err != nil {
					return fmt.Errorf("failed to add zapcore.ArrayMarshaler : %w", err)
				}
			case bool:
				inner.AddBool(key, val)
			case []bool:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendBool(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add byte array : %w", err)
				}
			case complex128:
				inner.AddComplex128(key, val)
			case []complex128:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendComplex128(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add complex128 array : %w", err)
				}
			case complex64:
				inner.AddComplex64(key, val)
			case []complex64:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendComplex64(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add complex64 array : %w", err)
				}
			case float64:
				inner.AddFloat64(key, val)
			case []float64:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendFloat64(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add float64 array : %w", err)
				}
			case float32:
				inner.AddFloat32(key, val)
			case []float32:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendFloat32(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add float32 array : %w", err)
				}
			case int:
				inner.AddInt(key, val)
			case []int:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendInt(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add int array : %w", err)
				}
			case int64:
				inner.AddInt64(key, val)
			case []int64:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendInt64(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add int64 array : %w", err)
				}
			case int32:
				inner.AddInt32(key, val)
			case []int32:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendInt32(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add int32 array : %w", err)
				}
			case int16:
				inner.AddInt16(key, val)
			case []int16:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendInt16(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add int16 array : %w", err)
				}
			case int8:
				inner.AddInt8(key, val)
			case []int8:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendInt8(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add int8 array : %w", err)
				}
			case string:
				inner.AddString(key, val)
			case []string:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendString(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add string array : %w", err)
				}
			case uint:
				inner.AddUint(key, val)
			case []uint:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendUint(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add uint array : %w", err)
				}
			case uint64:
				inner.AddUint64(key, val)
			case []uint64:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendUint64(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add uint64 array : %w", err)
				}
			case uint32:
				inner.AddUint32(key, val)
			case []uint32:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendUint32(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add uint32 array : %w", err)
				}
			case uint16:
				inner.AddUint16(key, val)
			case []uint16:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendUint16(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add uint16 array : %w", err)
				}
			case uint8:
				inner.AddUint8(key, val)
			case []byte:
				inner.AddBinary(key, val)
			case uintptr:
				inner.AddUintptr(key, val)
			case []uintptr:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendUintptr(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add uintptr array : %w", err)
				}
			case time.Time:
				inner.AddTime(key, val)
			case []time.Time:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendTime(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add time.Time array : %w", err)
				}
			case time.Duration:
				inner.AddDuration(key, val)
			case []time.Duration:
				if err := inner.AddArray(key, zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
					for _, v := range val {
						inner.AppendDuration(v)
					}
					return nil
				})); err != nil {
					return fmt.Errorf("failed to add time.Duration array : %w", err)
				}
			case fmt.Stringer:
				inner.AddString(key, val.String())
			default:
				if err := inner.AddReflected(key, val); err != nil {
					return fmt.Errorf("failed to add reflected : %w", err)
				}
			}
		}
		return nil
	})
}

type StructKeyExtractor func(tag string) string

func ForceStructToMap(v interface{}, tag string, extractor ...StructKeyExtractor) map[string]interface{} {
	m, _ := StructToMap(v, tag, extractor...)
	return m
}

func StructToMap(v interface{}, tag string, extractor ...StructKeyExtractor) (r map[string]interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	ke := func(tag string) string {
		return tag
	}
	if len(extractor) > 0 && extractor[0] != nil {
		ke = extractor[0]
	}

	r = make(map[string]interface{})

	elem := reflect.ValueOf(v)
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	if elem.Kind() != reflect.Struct {
		return r, fmt.Errorf("input value(%T) must be a struct or pointer to struct", v)
	}

	for i := 0; i < elem.NumField(); i++ {
		f := elem.Type().Field(i)

		var key string
		isOmitEmpty := false
		if ft, ok := f.Tag.Lookup(tag); ok {
			if ft == "-" {
				continue
			}
			if idx := strings.Index(ft, ","); idx == -1 {
				key = ke(ft)
			} else {
				key = ft[:idx]
				option := ft[idx+1:]
				isOmitEmpty = strings.Contains(option, "omitempty")
			}
		} else {
			key = f.Name
		}

		field := elem.Field(i)
		if field.IsZero() && isOmitEmpty {
			continue
		}
		r[key] = field.Interface()
	}

	return r, nil
}
