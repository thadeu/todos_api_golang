package sqlite

import (
	"database/sql"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) ScanRowToStruct(rows *sql.Rows, dest interface{}) error {
	destValue := reflect.ValueOf(dest)

	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	destElem := destValue.Elem()
	destType := destElem.Type()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	if !rows.Next() {
		return sql.ErrNoRows
	}

	scanArgs := make([]interface{}, len(columns))
	for i := range scanArgs {
		scanArgs[i] = new(interface{})
	}

	err = rows.Scan(scanArgs...)

	if err != nil {
		return err
	}

	for i, colName := range columns {
		val := *(scanArgs[i].(*interface{}))

		field := s.findStructField(destType, colName)

		if field.Name == "" || field.Type == nil {
			continue
		}

		if val != nil && field.Type.Kind() == reflect.Int {
			s.setFieldValue(destElem.FieldByIndex(field.Index), int(val.(int64)), field)
			continue
		}

		if err := s.setFieldValue(destElem.FieldByIndex(field.Index), val, field); err != nil {
			slog.Warn("Failed to set field", "field", field.Name, "error", err)
		}
	}

	return nil
}

func (s *Scanner) ScanRowsToSlice(rows *sql.Rows, dest interface{}) error {
	destValue := reflect.ValueOf(dest)

	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	sliceValue := destValue.Elem()
	elemType := sliceValue.Type().Elem()

	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("slice elements must be structs or pointers to structs")
	}

	for rows.Next() {
		elemValue := reflect.New(elemType)
		elem := elemValue.Interface()

		if err := s.ScanRowToStruct(rows, elem); err != nil {
			return err
		}

		if elemType.Kind() == reflect.Ptr {
			sliceValue.Set(reflect.Append(sliceValue, elemValue))
		} else {
			sliceValue.Set(reflect.Append(sliceValue, elemValue.Elem()))
		}
	}

	return nil
}

func (s *Scanner) findStructField(structType reflect.Type, colName string) reflect.StructField {
	colNameLower := strings.ToLower(colName)

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if strings.ToLower(field.Name) == colNameLower {
			return field
		}
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if tag := field.Tag.Get("db"); tag != "" && strings.ToLower(tag) == colNameLower {
			return field
		}
	}

	camelCaseName := s.snakeToCamel(colName)
	if field, found := structType.FieldByName(camelCaseName); found {
		return field
	}

	snakeCaseName := s.camelToSnake(colName)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if strings.ToLower(field.Name) == snakeCaseName {
			return field
		}
	}

	return reflect.StructField{}
}

func (s *Scanner) snakeToCamel(snake string) string {
	parts := strings.Split(snake, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
		}
	}
	return strings.Join(parts, "")
}

func (s *Scanner) camelToSnake(camel string) string {
	var result []rune
	for i, r := range camel {
		if i > 0 && unicode.IsUpper(r) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func (s *Scanner) shouldSkipField(field reflect.StructField) bool {
	tag := field.Tag.Get("scan")

	return tag == "skip"
}

func (s *Scanner) setFieldValue(field reflect.Value, val interface{}, structField reflect.StructField) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	if val == nil {
		return nil
	}

	fieldType := field.Type()

	if s.shouldSkipField(structField) {
		return nil
	}

	valValue := reflect.ValueOf(val)

	if valValue.IsValid() && valValue.Type().AssignableTo(fieldType) {
		field.Set(valValue)
		return nil
	}

	switch fieldType.Kind() {
	case reflect.String:
		if str, ok := val.(string); ok {
			field.SetString(str)
		}
	case reflect.Int, reflect.Int64:
		switch v := val.(type) {
		case int64:
			field.SetInt(v)
		case int:
			field.SetInt(int64(v))

			field.SetInt(int64(v))
		}
	case reflect.Bool:
		if b, ok := val.(bool); ok {
			field.SetBool(b)
		}
	case reflect.Float64, reflect.Float32:
		if f, ok := val.(float64); ok {
			field.SetFloat(f)
		}
	}

	switch fieldType.String() {
	case "uuid.UUID":
		if str, ok := val.(string); ok {
			if parsedUUID, err := uuid.Parse(str); err == nil {
				field.Set(reflect.ValueOf(parsedUUID))
			} else {
				slog.Warn("Failed to parse UUID", "value", str, "error", err)
			}
		}
	case "time.Time":
		if str, ok := val.(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, str); err == nil {
				field.Set(reflect.ValueOf(parsedTime))
			} else if parsedTime, err := time.Parse("2006-01-02 15:04:05", str); err == nil {
				field.Set(reflect.ValueOf(parsedTime))
			} else {
				slog.Warn("Failed to parse time", "value", str, "error", err)
			}
		}
	}

	return nil
}
