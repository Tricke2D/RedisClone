package protocol

import (
	"fmt"
	"io"

	"github.com/company/redis-clone/pkg/resp"
)

// RespEncoder handles encoding of RESP protocol responses
// Converts Go values ke RESP format untuk send ke clients
type RespEncoder struct {
	writer io.Writer
}

// NewRespEncoder creates new RESP encoder untuk given connection
func NewRespEncoder(w io.Writer) *RespEncoder {
	return &RespEncoder{writer: w}
}

// Encode writes RESP value ke writer dalam format yang benar
func (e *RespEncoder) Encode(val *resp.Value) error {
	if val == nil {
		return fmt.Errorf("cannot encode nil value")
	}

	switch val.Type {
	case resp.TypeSimpleString:
		return e.encodeSimpleString(val.Str)
	case resp.TypeError:
		return e.encodeError(val.Str)
	case resp.TypeInteger:
		return e.encodeInteger(val.Num)
	case resp.TypeBulkString:
		return e.encodeBulkString(val.Str, val.Num)
	case resp.TypeArray:
		return e.encodeArray(val.Arr)
	default:
		return fmt.Errorf("unknown RESP type: %c", val.Type)
	}
}

// encodeSimpleString writes "+string\r\n"
func (e *RespEncoder) encodeSimpleString(s string) error {
	_, err := fmt.Fprintf(e.writer, "+%s\r\n", s)
	return err
}

// encodeError writes "-error message\r\n"
func (e *RespEncoder) encodeError(msg string) error {
	_, err := fmt.Fprintf(e.writer, "-%s\r\n", msg)
	return err
}

// encodeInteger writes ":<number>\r\n"
func (e *RespEncoder) encodeInteger(n int64) error {
	_, err := fmt.Fprintf(e.writer, ":%d\r\n", n)
	return err
}

// encodeBulkString writes "$<length>\r\n<data>\r\n"
// Jika specialNum = -1, tulis null: "$-1\r\n"
func (e *RespEncoder) encodeBulkString(s string, specialNum int64) error {
	if specialNum == -1 {
		_, err := e.writer.Write([]byte("$-1\r\n"))
		return err
	}
	_, err := fmt.Fprintf(e.writer, "$%d\r\n%s\r\n", len(s), s)
	return err
}

// encodeArray writes "*<count>\r\n<element1>...<elementN>"
func (e *RespEncoder) encodeArray(elements []*resp.Value) error {
	if _, err := fmt.Fprintf(e.writer, "*%d\r\n", len(elements)); err != nil {
		return err
	}
	for _, elem := range elements {
		if err := e.Encode(elem); err != nil {
			return fmt.Errorf("failed to encode array element: %w", err)
		}
	}
	return nil
}

// EncodeOK adalah convenience untuk response "OK"
func EncodeOK(w io.Writer) error {
	_, err := w.Write([]byte("+OK\r\n"))
	return err
}

// EncodeNil adalah convenience untuk response null
func EncodeNil(w io.Writer) error {
	_, err := w.Write([]byte("$-1\r\n"))
	return err
}