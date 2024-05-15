package serialport

import "errors"

const (
	SLIP_SPECIAL_BYTE_END = 0xC0
	SLIP_SPECIAL_BYTE_ESC = 0xDB
	SLIP_ESCAPED_BYTE_END = 0xDC
	SLIP_ESCAPED_BYTE_ESC = 0xDD
)

const (
	SLIP_NO_ERROR                   = 0
	SLIP_ERROR_BUFFER_OVERFLOW      = 1
	SLIP_ERROR_UNKNOWN_ESCAPED_BYTE = 2
	SLIP_ERROR_CRC_MISMATCH         = 3
)

const (
	SLIP_STATE_NORMAL  = 0
	SLIP_STATE_ESCAPED = 1
)

var ErrBufferOverflow = errors.New("SLIP_ERROR_BUFFER_OVERFLOW")
var ErrUnknownEscapedByte = errors.New("SLIP_ERROR_UNKNOWN_ESCAPED_BYTE")

type SlipReadByte struct {
	buf          []byte
	size         int
	state        int
	recv_message func(buf []byte, size int)
}

func NewSlipReadByte(size_buf int, recv_msg func(buf []byte, size int)) *SlipReadByte {
	return &SlipReadByte{
		buf:          make([]byte, size_buf),
		recv_message: recv_msg,
	}
}

func (s *SlipReadByte) reset_rx() {
	s.state = SLIP_STATE_NORMAL
	s.size = 0
}

func (s *SlipReadByte) put_byte_to_buffer(value byte) error {
	var err error
	if s.size >= len(s.buf) {
		err = ErrBufferOverflow
		s.reset_rx()
	} else {
		s.buf[s.size] = value
		s.state = SLIP_STATE_NORMAL
		s.size++
	}
	return err
}

func (s *SlipReadByte) Readbyte(value byte) error {
	var err error
	switch s.state {
	case SLIP_STATE_NORMAL:
		switch value {
		case SLIP_SPECIAL_BYTE_END:
			if s.size >= 2 {
				s.recv_message(s.buf, s.size)
			}
			s.reset_rx()
		case SLIP_SPECIAL_BYTE_ESC:
			s.state = SLIP_STATE_ESCAPED
		default:
			err = s.put_byte_to_buffer(value)
		}
	case SLIP_STATE_ESCAPED:
		switch value {
		case SLIP_ESCAPED_BYTE_END:
			value = SLIP_SPECIAL_BYTE_END
		case SLIP_ESCAPED_BYTE_ESC:
			value = SLIP_SPECIAL_BYTE_ESC
		default:
			err = ErrUnknownEscapedByte
			s.reset_rx()
		}

		if err != nil {
			break
		}

		err = s.put_byte_to_buffer(value)
	}
	return err
}

type SlipWriteByte struct {
	buf          []byte
	size         int
	state        int
	send_message func(buf []byte, size int)
}

func NewSlipWriteByte(size_buf int, send_message func(buf []byte, size int)) *SlipWriteByte {
	return &SlipWriteByte{
		buf:          make([]byte, size_buf),
		send_message: send_message,
	}
}

func (s *SlipWriteByte) Reset() {
	s.size = 0
}

func (s *SlipWriteByte) put_byte_to_buffer(value byte) error {
	var err error
	if s.size >= len(s.buf) {
		err = ErrBufferOverflow
		s.Reset()
	} else {
		s.buf[s.size] = value
		s.state = SLIP_STATE_NORMAL
		s.size++
	}
	return err
}

func (s *SlipWriteByte) WriteByte(value byte) (err error) {
	escape := false

	switch value {
	case SLIP_SPECIAL_BYTE_END:
		value = SLIP_ESCAPED_BYTE_END
		escape = true
	case SLIP_SPECIAL_BYTE_ESC:
		value = SLIP_ESCAPED_BYTE_ESC
		escape = true
	}

	if escape {
		if err = s.put_byte_to_buffer(SLIP_SPECIAL_BYTE_ESC); err != nil {
			return err
		}
	}
	if err = s.put_byte_to_buffer(value); err != nil {
		return nil
	}
	return
}

func (s *SlipWriteByte) WriteEnd(value byte) (err error) {
	if err = s.put_byte_to_buffer(SLIP_SPECIAL_BYTE_ESC); err != nil {
		return err
	}
	s.send_message(s.buf, s.size)
	s.Reset()
	return
}
