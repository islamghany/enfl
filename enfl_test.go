package enfl

import (
	"reflect"
	"testing"
	"time"
)

func TestSetFieldValue(t *testing.T) {
	l := NewLoader()

	tests := []struct {
		name      string
		field     interface{}
		value     string
		fieldName string
		wantErr   bool
		expected  interface{}
	}{
		// String tests
		{
			name:      "String - Valid",
			field:     "",
			value:     "test",
			fieldName: "StringField",
			wantErr:   false,
			expected:  "test",
		},

		// Integer tests
		{
			name:      "Int - Valid",
			field:     int(0),
			value:     "42",
			fieldName: "IntField",
			wantErr:   false,
			expected:  int(42),
		},
		{
			name:      "Int - Invalid",
			field:     int(0),
			value:     "not-an-int",
			fieldName: "IntField",
			wantErr:   true,
			expected:  int(0),
		},
		{
			name:      "Int8 - Valid",
			field:     int8(0),
			value:     "42",
			fieldName: "Int8Field",
			wantErr:   false,
			expected:  int8(42),
		},
		{
			name:      "Int16 - Valid",
			field:     int16(0),
			value:     "42",
			fieldName: "Int16Field",
			wantErr:   false,
			expected:  int16(42),
		},
		{
			name:      "Int32 - Valid",
			field:     int32(0),
			value:     "42",
			fieldName: "Int32Field",
			wantErr:   false,
			expected:  int32(42),
		},
		{
			name:      "Int64 - Valid",
			field:     int64(0),
			value:     "42",
			fieldName: "Int64Field",
			wantErr:   false,
			expected:  int64(42),
		},

		// Unsigned integer tests
		{
			name:      "Uint - Valid",
			field:     uint(0),
			value:     "42",
			fieldName: "UintField",
			wantErr:   false,
			expected:  uint(42),
		},
		{
			name:      "Uint - Negative",
			field:     uint(0),
			value:     "-1",
			fieldName: "UintField",
			wantErr:   true,
			expected:  uint(0),
		},
		{
			name:      "Uint8 - Valid",
			field:     uint8(0),
			value:     "42",
			fieldName: "Uint8Field",
			wantErr:   false,
			expected:  uint8(42),
		},
		{
			name:      "Uint16 - Valid",
			field:     uint16(0),
			value:     "42",
			fieldName: "Uint16Field",
			wantErr:   false,
			expected:  uint16(42),
		},
		{
			name:      "Uint32 - Valid",
			field:     uint32(0),
			value:     "42",
			fieldName: "Uint32Field",
			wantErr:   false,
			expected:  uint32(42),
		},
		{
			name:      "Uint64 - Valid",
			field:     uint64(0),
			value:     "42",
			fieldName: "Uint64Field",
			wantErr:   false,
			expected:  uint64(42),
		},

		// Float tests
		{
			name:      "Float32 - Valid",
			field:     float32(0),
			value:     "3.14",
			fieldName: "Float32Field",
			wantErr:   false,
			expected:  float32(3.14),
		},
		{
			name:      "Float32 - Invalid",
			field:     float32(0),
			value:     "not-a-float",
			fieldName: "Float32Field",
			wantErr:   true,
			expected:  float32(0),
		},
		{
			name:      "Float64 - Valid",
			field:     float64(0),
			value:     "3.14159",
			fieldName: "Float64Field",
			wantErr:   false,
			expected:  float64(3.14159),
		},

		// Boolean tests
		{
			name:      "Bool - True",
			field:     false,
			value:     "true",
			fieldName: "BoolField",
			wantErr:   false,
			expected:  true,
		},
		{
			name:      "Bool - False",
			field:     true,
			value:     "false",
			fieldName: "BoolField",
			wantErr:   false,
			expected:  false,
		},
		{
			name:      "Bool - 1",
			field:     false,
			value:     "1",
			fieldName: "BoolField",
			wantErr:   false,
			expected:  true,
		},
		{
			name:      "Bool - 0",
			field:     true,
			value:     "0",
			fieldName: "BoolField",
			wantErr:   false,
			expected:  false,
		},
		{
			name:      "Bool - Invalid",
			field:     false,
			value:     "not-a-bool",
			fieldName: "BoolField",
			wantErr:   true,
			expected:  false,
		},

		// Duration tests
		{
			name:      "Duration - Valid",
			field:     time.Duration(0),
			value:     "5s",
			fieldName: "DurationField",
			wantErr:   false,
			expected:  5 * time.Second,
		},
		{
			name:      "Duration - Valid with Minutes",
			field:     time.Duration(0),
			value:     "2m30s",
			fieldName: "DurationField",
			wantErr:   false,
			expected:  2*time.Minute + 30*time.Second,
		},
		{
			name:      "Duration - Invalid",
			field:     time.Duration(0),
			value:     "not-a-duration",
			fieldName: "DurationField",
			wantErr:   true,
			expected:  time.Duration(0),
		},

		// Slice tests
		{
			name:      "String Slice - Valid",
			field:     []string{},
			value:     "a,b,c",
			fieldName: "StringSliceField",
			wantErr:   false,
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "Int Slice - Valid",
			field:     []int{},
			value:     "1,2,3",
			fieldName: "IntSliceField",
			wantErr:   false,
			expected:  []int{1, 2, 3},
		},
		{
			name:      "Int Slice - Invalid Element",
			field:     []int{},
			value:     "1,not-an-int,3",
			fieldName: "IntSliceField",
			wantErr:   true,
			expected:  []int{},
		},
		{
			name:      "Bool Slice - Valid",
			field:     []bool{},
			value:     "true,false,true",
			fieldName: "BoolSliceField",
			wantErr:   false,
			expected:  []bool{true, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a value to test with
			v := reflect.New(reflect.TypeOf(tt.field)).Elem()

			// Call setFieldValue
			err := l.setFieldValue(v, tt.value, tt.fieldName)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we don't expect an error, check the result
			if !tt.wantErr {
				got := v.Interface()
				if !reflect.DeepEqual(got, tt.expected) {
					t.Errorf("setFieldValue() = %v, want %v, name: %v", got, tt.expected, tt.name)
				}
			}
		})
	}
}

// TestSetSliceValue specifically tests the setSliceValue function
func TestSetSliceValue(t *testing.T) {
	l := NewLoader()

	tests := []struct {
		name      string
		sliceType interface{}
		value     string
		fieldName string
		wantErr   bool
		expected  interface{}
	}{
		{
			name:      "String Slice - Comma Separated",
			sliceType: []string{},
			value:     "a,b,c",
			fieldName: "StringSlice",
			wantErr:   false,
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "String Slice - With Spaces",
			sliceType: []string{},
			value:     "a, b, c",
			fieldName: "StringSlice",
			wantErr:   false,
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "Int Slice - Valid",
			sliceType: []int{},
			value:     "1,2,3",
			fieldName: "IntSlice",
			wantErr:   false,
			expected:  []int{1, 2, 3},
		},
		{
			name:      "Float Slice - Valid",
			sliceType: []float64{},
			value:     "1.1,2.2,3.3",
			fieldName: "FloatSlice",
			wantErr:   false,
			expected:  []float64{1.1, 2.2, 3.3},
		},
		{
			name:      "Empty Slice Value",
			sliceType: []string{},
			value:     "",
			fieldName: "EmptySlice",
			wantErr:   false,
			expected:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a slice value to test with
			v := reflect.New(reflect.TypeOf(tt.sliceType)).Elem()

			// Call setSliceValue
			err := l.setSliceValue(v, tt.value, tt.fieldName)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("setSliceValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error expected, check the result
			if !tt.wantErr && tt.value != "" {
				got := v.Interface()
				if !reflect.DeepEqual(got, tt.expected) {
					t.Errorf("setSliceValue() got = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}
