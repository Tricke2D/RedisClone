package resp

// RESP (Redis Serialization Protocol) type definitions
// Spec: https://redis.io/docs/reference/protocol-spec/

// RespType represents the type of RESP value
type RespType byte

const (
	// Simple string reply (prefix +)
	TypeSimpleString RespType = '+'
	// Error reply (prefix -)
	TypeError RespType = '-'
	// Integer reply (prefix :)
	TypeInteger RespType = ':'
	// Bulk string reply (prefix $)
	TypeBulkString RespType = '$'
	// Array reply (prefix *)
	TypeArray RespType = '*'
)

// Value represents a RESP value yang dapat berupa berbagai tipe
type Value struct {
	Type RespType // Tipe dari value (simple string, error, integer, bulk string, array)
	Str  string   // Untuk SimpleString, Error, BulkString
	Num  int64    // Untuk Integer; juga dipakai sebagai penanda null pada BulkString (-1)
	Arr  []*Value // Untuk Array
}

// NewSimpleString creates a simple string RESP value
// Used untuk response seperti "OK" dari SET command
func NewSimpleString(s string) *Value {
	return &Value{Type: TypeSimpleString, Str: s}
}

// NewError creates an error RESP value
func NewError(msg string) *Value {
	return &Value{Type: TypeError, Str: msg}
}

// NewInteger creates an integer RESP value
// Used untuk INCR, DEL (mengembalikan count), dll
func NewInteger(n int64) *Value {
	return &Value{Type: TypeInteger, Num: n}
}

// NewBulkString creates a bulk string RESP value
// Used untuk string data dari GET command
func NewBulkString(s string) *Value {
	return &Value{Type: TypeBulkString, Str: s}
}

// NewNull creates a null bulk string (special case)
// Used ketika key tidak ditemukan
func NewNull() *Value {
	return &Value{Type: TypeBulkString, Str: "", Num: -1} // -1 sebagai marker null
}

// NewArray creates an array RESP value
func NewArray(values ...*Value) *Value {
	return &Value{Type: TypeArray, Arr: values}
}

// IsNull checks if this value is a null bulk string
func (v *Value) IsNull() bool {
	return v.Type == TypeBulkString && v.Num == -1
}