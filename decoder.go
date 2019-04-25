package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	iso8601        = "20060102T15:04:05"
	iso8601Z       = "20060102T15:04:05Z07:00"
	iso8601Hyphen  = "2006-01-02T15:04:05"
	iso8601HyphenZ = "2006-01-02T15:04:05Z07:00"
	tagValue       = "value"
)

var (
	// CharsetReader is a function to generate reader which converts a non UTF-8
	// charset into UTF-8.
	CharsetReader func(string, io.Reader) (io.Reader, error)

	timeLayouts   = []string{iso8601, iso8601Z, iso8601Hyphen, iso8601HyphenZ}
	errInvalidXML = errors.New("invalid xml")
)

type TypeMismatchError string

func (e TypeMismatchError) Error() string { return string(e) }

type decoder struct {
	*xml.Decoder
}

func unmarshal(data []byte, v interface{}) (err error) {
	dec := &decoder{xml.NewDecoder(bytes.NewBuffer(data))}

	if CharsetReader != nil {
		dec.CharsetReader = CharsetReader
	}

	var tok xml.Token
	for {
		if tok, err = dec.Token(); err != nil {
			return err
		}

		if t, ok := tok.(xml.StartElement); ok {
			if t.Name.Local == tagValue {
				val := reflect.ValueOf(v)
				if val.Kind() != reflect.Ptr {
					return errors.New("non-pointer value passed to unmarshal")
				}
				if err = dec.decodeValue(val.Elem()); err != nil {
					return err
				}

				break
			}
		}
	}

	// read until end of document
	err = dec.Skip()
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (dec *decoder) decodeValue(val reflect.Value) error {
	var tok xml.Token
	var err error

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}
		val = val.Elem()
	}

	var typeName string
	for {
		if tok, err = dec.Token(); err != nil {
			return err
		}

		if t, ok := tok.(xml.EndElement); ok {
			if t.Name.Local == tagValue {
				return nil
			}

			return errInvalidXML
		}

		if t, ok := tok.(xml.StartElement); ok {
			typeName = t.Name.Local
			break
		}

		// Treat value data without type identifier as string
		if t, ok := tok.(xml.CharData); ok {
			if value := strings.TrimSpace(string(t)); value != "" {
				if err = checkType(val, reflect.String); err != nil {
					return err
				}

				val.SetString(value)
				return nil
			}
		}
	}

	switch typeName {
	case "struct":
		ismap := false
		pmap := val
		valType := val.Type()

		if err = checkType(val, reflect.Struct); err != nil {
			if checkType(val, reflect.Map) == nil {
				if valType.Key().Kind() != reflect.String {
					return fmt.Errorf("only maps with string key type can be unmarshalled")
				}
				ismap = true
			} else if checkType(val, reflect.Interface) == nil && val.IsNil() {
				var dummy map[string]interface{}
				pmap = reflect.New(reflect.TypeOf(dummy)).Elem()
				valType = pmap.Type()
				ismap = true
			} else {
				return err
			}
		}

		var fields map[string]reflect.Value

		if !ismap {
			fields = make(map[string]reflect.Value)

			for i := 0; i < valType.NumField(); i++ {
				field := valType.Field(i)
				fErr := handleStructField(fields, field, val.FieldByName(field.Name))
				if fErr != nil {
					return fErr
				}
			}
		} else {
			// Create initial empty map
			pmap.Set(reflect.MakeMap(valType))
		}

		// Process struct members.
	StructLoop:
		for {
			if tok, err = dec.Token(); err != nil {
				return err
			}
			switch t := tok.(type) {
			case xml.StartElement:
				if t.Name.Local != "member" {
					return errInvalidXML
				}

				tagName, fieldName, err := dec.readTag()
				if err != nil {
					return err
				}
				if tagName != "name" {
					return errInvalidXML
				}

				var fv reflect.Value
				ok := true

				if !ismap {
					fv, ok = fields[string(fieldName)]
				} else {
					fv = reflect.New(valType.Elem())
				}

				if ok {
					for {
						if tok, err = dec.Token(); err != nil {
							return err
						}
						if t, ok := tok.(xml.StartElement); ok && t.Name.Local == tagValue {
							if err = dec.decodeValue(fv); err != nil {
								return err
							}

							// </value>
							if err = dec.Skip(); err != nil {
								return err
							}

							break
						}
					}
				}

				// </member>
				if err = dec.Skip(); err != nil {
					return err
				}

				if ismap {
					pmap.SetMapIndex(reflect.ValueOf(string(fieldName)), reflect.Indirect(fv))
					val.Set(pmap)
				}
			case xml.EndElement:
				break StructLoop
			}
		}
	case "array":
		slice := val
		if checkType(val, reflect.Interface) == nil && val.IsNil() {
			slice = reflect.ValueOf([]interface{}{})
		} else if err = checkType(val, reflect.Slice); err != nil {
			return err
		}

	ArrayLoop:
		for {
			if tok, err = dec.Token(); err != nil {
				return err
			}

			switch t := tok.(type) {
			case xml.StartElement:
				var index int
				if t.Name.Local != "data" {
					return errInvalidXML
				}
			DataLoop:
				for {
					if tok, err = dec.Token(); err != nil {
						return err
					}

					switch tt := tok.(type) {
					case xml.StartElement:
						if tt.Name.Local != tagValue {
							return errInvalidXML
						}

						if index < slice.Len() {
							v := slice.Index(index)
							if v.Kind() == reflect.Interface {
								v = v.Elem()
							}
							if v.Kind() != reflect.Ptr {
								return errors.New("error: cannot write to non-pointer array element")
							}
							if err = dec.decodeValue(v); err != nil {
								return err
							}
						} else {
							v := reflect.New(slice.Type().Elem())
							if err = dec.decodeValue(v); err != nil {
								return err
							}
							slice = reflect.Append(slice, v.Elem())
						}

						// </value>
						if err = dec.Skip(); err != nil {
							return err
						}
						index++
					case xml.EndElement:
						val.Set(slice)
						break DataLoop
					}
				}
			case xml.EndElement:
				break ArrayLoop
			}
		}
	default:
		if tok, err = dec.Token(); err != nil {
			return err
		}

		var data []byte

		switch t := tok.(type) {
		case xml.EndElement:
			return nil
		case xml.CharData:
			data = []byte(t.Copy())
		default:
			return errInvalidXML
		}

		switch typeName {
		case "int", "i4", "i8":
			if checkType(val, reflect.Interface) == nil && val.IsNil() {
				i, err := strconv.ParseInt(string(data), 10, 64)
				if err != nil {
					return err
				}

				pi := reflect.New(reflect.TypeOf(i)).Elem()
				pi.SetInt(i)
				val.Set(pi)
			} else if err = checkType(val, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64); err != nil {
				return err
			} else {
				i, err := strconv.ParseInt(string(data), 10, val.Type().Bits())
				if err != nil {
					return err
				}

				val.SetInt(i)
			}
		case "string", "base64":
			str := string(data)
			if checkType(val, reflect.Interface) == nil && val.IsNil() {
				pstr := reflect.New(reflect.TypeOf(str)).Elem()
				pstr.SetString(str)
				val.Set(pstr)
			} else if err = checkType(val, reflect.String); err != nil {
				return err
			} else {
				val.SetString(str)
			}
		case "dateTime.iso8601":
			var t time.Time
			var err error

			for _, layout := range timeLayouts {
				t, err = time.Parse(layout, string(data))
				if err == nil {
					break
				}
			}
			if err != nil {
				return err
			}

			if checkType(val, reflect.Interface) == nil && val.IsNil() {
				ptime := reflect.New(reflect.TypeOf(t)).Elem()
				ptime.Set(reflect.ValueOf(t))
				val.Set(ptime)
			} else if _, ok := val.Interface().(time.Time); !ok {
				return TypeMismatchError(fmt.Sprintf("error: type mismatch error - can't decode %v to time", val.Kind()))
			} else {
				val.Set(reflect.ValueOf(t))
			}
		case "boolean":
			v, err := strconv.ParseBool(string(data))
			if err != nil {
				return err
			}

			if checkType(val, reflect.Interface) == nil && val.IsNil() {
				pv := reflect.New(reflect.TypeOf(v)).Elem()
				pv.SetBool(v)
				val.Set(pv)
			} else if err = checkType(val, reflect.Bool); err != nil {
				return err
			} else {
				val.SetBool(v)
			}
		case "double":
			if checkType(val, reflect.Interface) == nil && val.IsNil() {
				i, err := strconv.ParseFloat(string(data), 64)
				if err != nil {
					return err
				}

				pdouble := reflect.New(reflect.TypeOf(i)).Elem()
				pdouble.SetFloat(i)
				val.Set(pdouble)
			} else if err = checkType(val, reflect.Float32, reflect.Float64); err != nil {
				return err
			} else {
				i, err := strconv.ParseFloat(string(data), val.Type().Bits())
				if err != nil {
					return err
				}

				val.SetFloat(i)
			}
		default:
			return errors.New("unsupported type")
		}

		// </type>
		if err = dec.Skip(); err != nil {
			return err
		}
	}

	return nil
}

func (dec *decoder) readTag() (string, []byte, error) {
	var tok xml.Token
	var err error

	var name string
	for {
		if tok, err = dec.Token(); err != nil {
			return "", nil, err
		}

		if t, ok := tok.(xml.StartElement); ok {
			name = t.Name.Local
			break
		}
	}

	value, err := dec.readCharData()
	if err != nil {
		return "", nil, err
	}

	return name, value, dec.Skip()
}

func (dec *decoder) readCharData() ([]byte, error) {
	var tok xml.Token
	var err error

	if tok, err = dec.Token(); err != nil {
		return nil, err
	}

	if t, ok := tok.(xml.CharData); ok {
		return []byte(t.Copy()), nil
	}

	return nil, errInvalidXML
}

func checkType(val reflect.Value, kinds ...reflect.Kind) error {
	if len(kinds) == 0 {
		return nil
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	match := false

	for _, kind := range kinds {
		if val.Kind() == kind {
			match = true
			break
		}
	}

	if !match {
		return TypeMismatchError(fmt.Sprintf("error: type mismatch - can't unmarshal %v to %v",
			val.Kind(), kinds[0]))
	}

	return nil
}

// handleStructField will add struct values to the fields map indexed by the "xmlrpc" tag, defaulting to the
// struct field name if no tag is found. This function will recurse into embedded structs. Conflicting
// tag/field names will result in an error.
func handleStructField(fields map[string]reflect.Value, field reflect.StructField, val reflect.Value) error {
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}

		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			var sField reflect.StructField
			if field.Type.Kind() == reflect.Ptr {
				sField = field.Type.Elem().Field(i)
			} else {
				sField = field.Type.Field(i)
			}

			fErr := handleStructField(fields, sField, val.FieldByName(sField.Name))
			if fErr != nil {
				return fErr
			}
		}
	}

	if val.CanSet() {
		fn := field.Tag.Get("xmlrpc")
		if fn == "" {
			fn = field.Name
		}

		if _, exists := fields[fn]; exists {
			return fmt.Errorf("duplicate field %q in struct", fn)
		}

		fields[fn] = val
	}

	return nil
}
