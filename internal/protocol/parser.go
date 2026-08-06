package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"github.com/company/redis-clone/pkg/resp"
)

// RespParser handles parsing of RESP protocol messages
// Format umum: prefix byte + data + CRLF (\r\n)
type RespParser struct {
	reader *bufio.Reader // Buffered reader untuk efficient TCP reading
}

// NewRespParser creates a new RESP parser untuk given connection/reader
func NewRespParser(r io.Reader) *RespParser {
	return &RespParser{reader: bufio.NewReader(r)}
}

// Parse reads dan parses satu RESP message dari koneksi
func (p *RespParser) Parse() (*resp.Value, error) {
	typeByte, err := p.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read type byte: %w", err)
	}

	switch typeByte {
	case byte(resp.TypeSimpleString):
		return p.parseSimpleString()
	case byte(resp.TypeError):
		return p.parseError()
	case byte(resp.TypeInteger):
		return p.parseInteger()
	case byte(resp.TypeBulkString):
		return p.parseBulkString()
	case byte(resp.TypeArray):
		return p.parseArray()
	default:
		return nil, fmt.Errorf("unknown RESP type: %c", typeByte)
	}
}

// parseSimpleString reads format: +string\r\n
func (p *RespParser) parseSimpleString() (*resp.Value, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, fmt.Errorf("failed to read simple string: %w", err)
	}
	return resp.NewSimpleString(line), nil
}

// parseError reads format: -error message\r\n
func (p *RespParser) parseError() (*resp.Value, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, fmt.Errorf("failed to read error: %w", err)
	}
	return resp.NewError(line), nil
}

// parseInteger reads format: :1000\r\n
func (p *RespParser) parseInteger() (*resp.Value, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, fmt.Errorf("failed to read integer: %w", err)
	}
	num, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer format: %s", line)
	}
	return resp.NewInteger(num), nil
}

// parseBulkString reads format: $<length>\r\n<data>\r\n
// Special case: $-1\r\n adalah null
func (p *RespParser) parseBulkString() (*resp.Value, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, fmt.Errorf("failed to read bulk string length: %w", err)
	}

	length, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid bulk string length: %s", line)
	}

	if length == -1 {
		return resp.NewNull(), nil
	}
	if length < 0 {
		return nil, fmt.Errorf("invalid bulk string length: %d", length)
	}

	data := make([]byte, length)
	n, err := io.ReadFull(p.reader, data)
	if err != nil {
		return nil, fmt.Errorf("failed to read bulk string data: %w", err)
	}
	if int64(n) != length {
		return nil, fmt.Errorf("read %d bytes, expected %d", n, length)
	}

	if err := p.readCRLF(); err != nil {
		return nil, fmt.Errorf("invalid bulk string terminator: %w", err)
	}

	return resp.NewBulkString(string(data)), nil
}

// parseArray reads format: *<count>\r\n<element1><element2>...<elementN>
func (p *RespParser) parseArray() (*resp.Value, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, fmt.Errorf("failed to read array count: %w", err)
	}

	count, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid array count: %s", line)
	}
	if count == -1 {
		return nil, nil // Null array
	}

	elements := make([]*resp.Value, 0, count)
	for i := int64(0); i < count; i++ {
		element, err := p.Parse() // Recursive call
		if err != nil {
			return nil, fmt.Errorf("failed to parse array element %d: %w", i, err)
		}
		elements = append(elements, element)
	}

	return resp.NewArray(elements...), nil
}

// readLine reads hingga \r\n dan return string tanpa terminator
func (p *RespParser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) < 2 {
		return "", fmt.Errorf("incomplete line")
	}
	return line[:len(line)-2], nil // Buang \r\n
}

// readCRLF reads dan validates trailing \r\n
func (p *RespParser) readCRLF() error {
	b1, err := p.reader.ReadByte()
	if err != nil {
		return err
	}
	b2, err := p.reader.ReadByte()
	if err != nil {
		return err
	}
	if b1 != '\r' || b2 != '\n' {
		return fmt.Errorf("expected CRLF, got: %c%c", b1, b2)
	}
	return nil
}

// ParseCommand adalah convenience function untuk parse array commands
// Converts RESP array ke []string (untuk easy command dispatch)
func ParseCommand(val *resp.Value) ([]string, error) {
	if val.Type != resp.TypeArray {
		return nil, fmt.Errorf("expected array, got %c", val.Type)
	}

	cmd := make([]string, 0, len(val.Arr))
	for _, elem := range val.Arr {
		if elem.Type != resp.TypeBulkString {
			return nil, fmt.Errorf("expected bulk string in command array, got %c", elem.Type)
		}
		cmd = append(cmd, elem.Str)
	}

	return cmd, nil
}