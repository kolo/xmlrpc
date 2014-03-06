package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const iso8601 = "20060102T15:04:05"

var (
	invalidXmlError   = errors.New("invalid xml")
	typeMismatchError = errors.New("type mismatch")
)

type decoder struct {
	*xml.Decoder
}

func unmarshal(data []byte, v interface{}) (err error) {
	dec := &decoder{xml.NewDecoder(bytes.NewBuffer(data))}

	var tok xml.Token
	for {
		if tok, err = dec.Token(); err != nil {
			return err
		}

		if t, ok := tok.(xml.StartElement); ok {
			if t.Name.Local == "value" {
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

	var typeName string
	for {
		if tok, err = dec.Token(); err != nil {
			return err
		}

		if t, ok := tok.(xml.StartElement); ok {
			typeName = t.Name.Local
			break
		}

		// Treat value data without type identifier as string
		if t, ok := tok.(xml.CharData); ok {
			if value := strings.TrimSpace(string(t)); value != "" {
				typeName = "string"
				val.SetString(value)
				return nil
			}
		}
	}

	switch typeName {
	case "struct":
		fields := make(map[string]reflect.Value)

		valType := val.Type()
		for i := 0; i < valType.NumField(); i++ {
			field := valType.Field(i)
			if fn := field.Tag.Get("xmlrpc"); fn != "" {
				fields[fn] = val.FieldByName(field.Name)
			} else {
				fields[field.Name] = val.FieldByName(field.Name)
			}
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
					return invalidXmlError
				}

				tagName, fieldName, err := dec.readTag()
				if err != nil {
					return err
				}
				if tagName != "name" {
					return invalidXmlError
				}

				if fv, ok := fields[string(fieldName)]; ok {
					for {
						if tok, err = dec.Token(); err != nil {
							return err
						}
						if t, ok := tok.(xml.StartElement); ok && t.Name.Local == "value" {
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
			case xml.EndElement:
				break StructLoop
			}
		}
	case "array":
	ArrayLoop:
		for {
			if tok, err = dec.Token(); err != nil {
				return err
			}
			switch t := tok.(type) {
			case xml.StartElement:
				if t.Name.Local != "data" {
					return invalidXmlError
				}

				slice := reflect.MakeSlice(val.Type(), 0, 0)

			DataLoop:
				for {
					if tok, err = dec.Token(); err != nil {
						return err
					}

					switch tt := tok.(type) {
					case xml.StartElement:
						if tt.Name.Local != "value" {
							return invalidXmlError
						}

						v := reflect.New(val.Type().Elem())
						if err = dec.decodeValue(v); err != nil {
							return err
						}

						slice = reflect.Append(slice, v.Elem())

						// </value>
						if err = dec.Skip(); err != nil {
							return err
						}
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
		data, err := dec.readCharData()
		if err != nil {
			return err
		}

		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				val.Set(reflect.New(val.Type().Elem()))
			}
			val = val.Elem()
		}

		switch typeName {
		case "int", "i4":
			i, err := strconv.ParseInt(string(data), 10, val.Type().Bits())
			if err != nil {
				return err
			}

			val.SetInt(i)
		case "string":
			val.SetString(string(data))
		case "dateTime.iso8601":
			t, err := time.Parse(iso8601, string(data))
			if err != nil {
				return err
			}

			val.Set(reflect.ValueOf(t))
		case "boolean":
			v, err := strconv.ParseBool(string(data))
			if err != nil {
				return err
			}

			val.SetBool(v)
		case "double":
			i, err := strconv.ParseFloat(string(data), val.Type().Bits())
			if err != nil {
				return err
			}

			val.SetFloat(i)
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
	} else {
		return nil, invalidXmlError
	}
}
