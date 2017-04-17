package gongular2

import (
	"encoding/json"
	"errors"
	"reflect"
)

func (c *Context) parseParams(handlerObject reflect.Value) error {
	param := handlerObject.FieldByName("Param")
	paramType := param.Type()

	numFields := paramType.NumField()
	for i := 0; i < numFields; i++ {
		field := paramType.Field(i)
		// TODO: Parse it accordingly, int-string
		s := c.Params().ByName(field.Name)
		param.Field(i).SetString(s)
	}
	return nil
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
		query.Field(i).SetString(s)
	}
	return nil
}

func (c *Context) parseForm(obj reflect.Value) error {
	// See if we parsed earlier
	// TODO: Cache the files in the context so that we do not re-read it unneccessarliy
	contentType := c.Request().Header.Get("Content-Type")

	if contentType == "multipart...." {
		err := c.Request().ParseMultipartForm(1024 * 1024) // 1 MB??
		if err != nil {
			return err
		}
	} else if contentType == "url encoded etc" {
		err := c.Request().ParseForm()
		if err != nil {
			return err
		}
	} else {
		return errors.New("Not a multipart or url encoded request but it is")
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
