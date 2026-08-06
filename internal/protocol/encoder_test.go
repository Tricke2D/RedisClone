package protocol

import (
	"bytes"
	"testing"

	"github.com/company/redis-clone/pkg/resp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEncodeSimpleString verifies encoding ke "+OK\r\n" format
func TestEncodeSimpleString(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewRespEncoder(&buf)

	val := resp.NewSimpleString("OK")
	err := encoder.Encode(val)
	require.NoError(t, err)

	assert.Equal(t, "+OK\r\n", buf.String())
}

// TestEncodeError verifies encoding ke "-Error\r\n" format
func TestEncodeError(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewRespEncoder(&buf)

	val := resp.NewError("ERR unknown command")
	err := encoder.Encode(val)
	require.NoError(t, err)

	assert.Equal(t, "-ERR unknown command\r\n", buf.String())
}

// TestEncodeInteger verifies encoding ke ":1000\r\n" format
func TestEncodeInteger(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewRespEncoder(&buf)

	val := resp.NewInteger(1000)
	err := encoder.Encode(val)
	require.NoError(t, err)

	assert.Equal(t, ":1000\r\n", buf.String())
}

// TestEncodeBulkString verifies encoding ke "$6\r\nfoobar\r\n" format
func TestEncodeBulkString(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewRespEncoder(&buf)

	val := resp.NewBulkString("foobar")
	err := encoder.Encode(val)
	require.NoError(t, err)

	assert.Equal(t, "$6\r\nfoobar\r\n", buf.String())
}

// TestEncodeNull verifies encoding ke "$-1\r\n" format
func TestEncodeNull(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewRespEncoder(&buf)

	val := resp.NewNull()
	err := encoder.Encode(val)
	require.NoError(t, err)

	assert.Equal(t, "$-1\r\n", buf.String())
}

// TestEncodeArray verifies encoding ke "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n" format
func TestEncodeArray(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewRespEncoder(&buf)

	val := resp.NewArray(
		resp.NewBulkString("foo"),
		resp.NewBulkString("bar"),
	)
	err := encoder.Encode(val)
	require.NoError(t, err)

	assert.Equal(t, "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", buf.String())
}

// TestRoundTripParseEncode verifies parser dan encoder konsisten satu sama lain
func TestRoundTripParseEncode(t *testing.T) {
	testCases := []string{
		"+OK\r\n",
		"-ERR error\r\n",
		":42\r\n",
		"$5\r\nhello\r\n",
		"$-1\r\n",
		"*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
	}

	for _, testInput := range testCases {
		t.Run(testInput, func(t *testing.T) {
			parser := NewRespParser(bytes.NewReader([]byte(testInput)))
			val, err := parser.Parse()
			require.NoError(t, err)

			var buf bytes.Buffer
			encoder := NewRespEncoder(&buf)
			err = encoder.Encode(val)
			require.NoError(t, err)

			assert.Equal(t, testInput, buf.String())
		})
	}
}