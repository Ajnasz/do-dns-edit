package main

import "reflect"
import "fmt"

func validateField(field reflect.StructField, v reflect.Value) error {
	if required, ok := field.Tag.Lookup("required"); ok {
		if required == "true" {
			value := v.FieldByName(field.Name)
			isOk := false

			switch field.Type.Name() {
			case "string":
				isOk = value.String() != ""
				break
			case "boolean":
				isOk = value.Bool()
				break
			}

			if !isOk {
				return fmt.Errorf("%s is required", field.Name)
			}
		}
	}

	return nil
}

func validateConfig(config Config) error {
	st := reflect.TypeOf(config)
	v := reflect.ValueOf(config)

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)

		err := validateField(field, v)

		if err != nil {
			return err
		}

	}

	return nil
}
