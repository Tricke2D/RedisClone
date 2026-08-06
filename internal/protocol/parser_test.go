package protocol

import (
	"bytes"
	"testing"

	"github.com/company/redis-clone/pkg/resp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseSimpleString verifies parsing dari "+OK\r\n" format
func TestParseSimpleString(t *testing.T) {
	input := "+OK\r\n"
	parser := NewRespParser(bytes.NewReader([]byte(input)))

	val, err := parser.Parse()
	require.NoError(t, err)
	assert.Equal(t, resp.TypeSimpleString, val.Type)
	assert.Equal(t, "OK", val.Str)
}

// TestParseError verifies parsing dari "-Error message\r\n" format
func TestParseError(t *testing.T) {
	input := "-ERR unknown command\r\n"
	parser := NewRespParser(bytes.NewReader([]byte(input)))

	val, err := parser.Parse()
	require.NoError(t, err)
	assert.Equal(t, resp.TypeError, val.Type)
	assert.Equal(t, "ERR unknown command", val.Str)
}

// TestParseInteger verifies parsing dari ":1000\r\n" format
func TestParseInteger(t *testing.T) {
	input := ":1000\r\n"
	parser := NewRespParser(bytes.NewReader([]byte(input)))

	val, err := parser.Parse()
	require.NoError(t, err)
	assert.Equal(t, resp.TypeInteger, val.Type)
	assert.Equal(t, int64(1000), val.Num)
}

// TestParseBulkString verifies parsing dari "$6\r\nfoobar\r\n" format
func TestParseBulkString(t *testing.T) {
	input := "$6\r\nfoobar\r\n"
	parser := NewRespParser(bytes.NewReader([]byte(input)))

	val, err := parser.Parse()
	require.NoError(t, err)
	assert.Equal(t, resp.TypeBulkString, val.Type)
	assert.Equal(t, "foobar", val.Str)
}

// TestParseBulkStringNull verifies null bulk string: "$-1\r\n"
func TestParseBulkStringNull(t *testing.T) {
	input := "$-1\r\n"
	parser := NewRespParser(bytes.NewReader([]byte(input)))

	val, err := parser.Parse()
	require.NoError(t, err)
	assert.True(t, val.IsNull())
}

// TestParseArray verifies parsing dari "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n" format
func TestParseArray(t *testing.T) {
	input := "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	parser := NewRespParser(bytes.NewReader([]byte(input)))

	val, err := parser.Parse()
	require.NoError(t, err)
	assert.Equal(t, resp.TypeArray, val.Type)
	assert.Equal(t, 2, len(val.Arr))
	assert.Equal(t, "foo", val.Arr[0].Str)
	assert.Equal(t, "bar", val.Arr[1].Str)
}

// TestParseCommand verifies conversion dari array ke command strings
func TestParseCommand(t *testing.T) {
	input := "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	parser := NewRespParser(bytes.NewReader([]byte(input)))

	val, err := parser.Parse()
	require.NoError(t, err)

	cmd, err := ParseCommand(val)
	require.NoError(t, err)
	assert.Equal(t, []string{"SET", "foo", "bar"}, cmd)
}

// TestParseInvalidFormat verifies error handling untuk malformed RESP
func TestParseInvalidFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Invalid type byte", "x\r\n", true},
		{"Integer invalid format", ":abc\r\n", true},
		{"Bulk string truncated", "$10\r\nshort", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewRespParser(bytes.NewReader([]byte(tt.input)))
			_, err := parser.Parse()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}