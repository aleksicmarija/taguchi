package taguchi

import (
	"fmt"
	"reflect"
)

// factorsFrom extracts a []Factor from the exported []float64 fields of a
// struct value. Each field becomes a factor with Name = field name and
// Levels = the slice value.
func factorsFrom[T any](v T) ([]ControlFactor, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("factorsFrom requires a struct value, got %s", rv.Kind())
	}
	t := rv.Type()

	var factors []ControlFactor
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		if field.Type != reflect.TypeOf([]float64{}) {
			continue
		}
		levels := rv.Field(i).Interface().([]float64)
		if len(levels) < 2 {
			return nil, fmt.Errorf("field %s: at least 2 levels required, got %d", field.Name, len(levels))
		}
		factors = append(factors, ControlFactor{Name: field.Name, Levels: levels})
	}

	if len(factors) == 0 {
		return nil, fmt.Errorf("no exported []float64 fields found in %s", t.Name())
	}
	return factors, nil
}

// buildControlAs pre-computes field indices for type P and returns a closure
// that converts a Trial's Control map into a value of P.
func buildControlAs[P any]() func(Trial) P {
	var zero P
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return func(trial Trial) P {
			return zero
		}
	}

	type fieldInfo struct {
		index int
		name  string
	}
	var fields []fieldInfo
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() || field.Type.Kind() != reflect.Float64 {
			continue
		}
		fields = append(fields, fieldInfo{index: i, name: field.Name})
	}

	return func(trial Trial) P {
		var result P
		v := reflect.ValueOf(&result).Elem()
		for _, f := range fields {
			if val, ok := trial.Control[f.name]; ok {
				v.Field(f.index).SetFloat(val)
			}
		}
		return result
	}
}
