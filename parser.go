package gongular2

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
)

const (
	PlaceParameter = "URL Path Parameter"
	PlaceQuery     = "Query Parameter"
	PlaceBody      = "Body"
)

func checkIntRange(kind reflect.Kind, val int64) bool {
	// TODO: int8 -> -127+128 etc
	return false
}

func parseSimpleParam(s string, place string, field reflect.StructField, val *reflect.Value) error {
	switch field.Type.Kind() {
	case reflect.String:
		val.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return ParseError{
				FieldName: field.Name,
				Reason:    fmt.Sprintf("The '%s' is not parseable to int", s),
			}
		}
		val.SetInt(i)
	case reflect.Float32, reflect.Float64:
		i, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return ParseError{
				FieldName: field.Name,
				Reason:    fmt.Sprintf("The '%s' is not parseable to float/double", s),
			}
		}
		val.SetFloat(i)
	case reflect.Bool:
		switch strings.ToLower(s) {
		case "true", "1", "yes":
			val.SetBool(true)
		case "false", "0", "no":
			val.SetBool(false)
		default:
			return ParseError{
				FieldName: field.Name,
				Place:     place,
				Reason:    fmt.Sprintf("The '%s' is not a boolean", s),
			}
		}
	}
	return nil
}

func validateStruct(obj reflect.Value, place string) error {
	isValid, err := govalidator.ValidateStruct(obj.Interface())
	if !isValid {
		m := govalidator.ErrorsByField(err)
		return ValidationError{
			Place:  place,
			Fields: m,
		}
	}
	return nil
}

func (c *Context) parseParams(obj reflect.Value) error {
	param := obj.FieldByName("Param")
	paramType := param.Type()

	numFields := paramType.NumField()
	for i := 0; i < numFields; i++ {
		field := paramType.Field(i)

		// TODO: Parse it accordingly, int-string
		s := c.Params().ByName(field.Name)
		val := param.Field(i)
		err := parseSimpleParam(s, PlaceParameter, field, &val)
		if err != nil {
			return err
		}
	}

	return validateStruct(obj, PlaceParameter)
}

func (c *Context) parseBody(handlerObject reflect.Value) error {
	// Cache body if possible
	body := handlerObject.FieldByName("Body")
	b := body.Addr().Interface()

	err := json.NewDecoder(c.Request().Body).Decode(b)
	return err
}

func (c *Context) parseQuery(obj reflect.Value) error {
	query := obj.FieldByName("Query")
	queryType := obj.Type()

	numFields := queryType.NumField()
	queryValues := c.Request().URL.Query()
	for i := 0; i < numFields; i++ {
		field := queryType.Field(i)
		// TODO: Parse it accordingly, int-string
		s := queryValues.Get(field.Name)
		val := query.Field(i)
		err := parseSimpleParam(s, PlaceQuery, field, &val)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) parseForm(obj reflect.Value) error {
	// See if we parsed earlier
	// TODO: Cache the files in the context so that we do not re-read it unnecessarily
	contentType := c.Request().Header.Get("Content-Type")

	// TODO: Instead of header-check, check if we have files only
	if strings.Contains(contentType, "multipart/form-data") {
		err := c.Request().ParseMultipartForm(10 * 1024 * 1024) // 10 MB??
		if err != nil {
			return err
		}
	} else if contentType == "application/x-www-form-urlencoded" {
		err := c.Request().ParseForm()
		if err != nil {
			return err
		}
	} else {
		return errors.New("The request's content-type is neither multipart/form-data" +
			" or application/x-www-form-urlencoded, it is: " + contentType)
	}

	form := obj.FieldByName("Form")
	formType := form.Type()

	numFields := formType.NumField()

	for i := 0; i < numFields; i++ {
		field := formType.Field(i)
		// If it is a file, parse the form
		if field.Type == reflect.TypeOf(&UploadedFile{}) {
			file, header, err := c.Request().FormFile(field.Name)

			if err != nil {
				return err
			}

			// Pack them to a single structure
			uploadedFile := &UploadedFile{
				File:   file,
				Header: header,
			}

			form.Field(i).Set(reflect.ValueOf(uploadedFile))
		} else {
			s := c.Request().FormValue(field.Name)
			form.Field(i).SetString(s)
		}
	}
	return nil
}

func (c *Context) parseInjections(obj reflect.Value, injector *injector) error {
	numFields := obj.Type().NumField()

	for i := 0; i < numFields; i++ {
		field := obj.Type().Field(i)
		tip := field.Type
		name := field.Name

		// We can skip the field if it is a special one
		if name == "Body" || name == "Param" || name == "Query" || name == "Form" {
			continue
		}

		if !obj.Field(i).CanSet() {
			// It is an un-exported one
			continue
		}

		var key string
		tag, ok := field.Tag.Lookup("inject")
		if !ok {
			key = "default"
		} else {
			key = tag
		}

		val, directOk := injector.GetDirectValue(tip, key)
		fn, customOk := injector.GetCustomValue(tip, key)

		if directOk {
			obj.Field(i).Set(reflect.ValueOf(val))
		} else if customOk {
			// TODO: Cache the result so that in subsequent requests we do not have to execute the function
			val, err := fn(c)
			if err != nil {
				return errors.New("Could not")
			}
			obj.Field(i).Set(reflect.ValueOf(val))
		} else {
			return errors.New("WOW NO SUCH DEPendency")
		}
	}
	return nil
}
