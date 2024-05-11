package serialport

import (
	"fmt"
)

type ChannelI interface {
	Read(b []byte, e int) (int, error)
	Write(b []byte) (int, error)
}

const (
	BUF_SIZE  = 1024 * 2
	COD_FLUSH = 0xFF // Flush byte
	COD_END   = 0xC0 // Indicates end of packet
	COD_ESC   = 0xDB // indicates byte stuffing
	ESC_END   = 0xDC // ESC ESC_END means END data byte
	ESC_ESC   = 0xDD // ESC ESC_ESC means ESC data byte
)

const (
	SLIP_SPECIAL_BYTE_END = 0xC0
	SLIP_SPECIAL_BYTE_ESC = 0xDB
	SLIP_ESCAPED_BYTE_END = 0xDC
	SLIP_ESCAPED_BYTE_ESC = 0xDD
)

/*
#define SLIP_END     0xC0
#define SLIP_ESC     0xDB
#define SLIP_ESC_END 0xDC
#define SLIP_ESC_ESC 0xDD
*/

func SlipWrite(ch_i ChannelI, buf []byte) (int, error) {
	var write_buf [BUF_SIZE]byte
	out_len := 2
	write_buf[0] = COD_FLUSH
	write_buf[1] = COD_END
	for _, c := range buf {
		switch c {
		case COD_END:
			if out_len+2 > BUF_SIZE {
				return 0, fmt.Errorf("write buf %d size exceeded %d", BUF_SIZE, out_len)
			}
			write_buf[out_len] = COD_ESC
			out_len++
			write_buf[out_len] = ESC_END
			out_len++
		case COD_ESC:
			if out_len+2 > BUF_SIZE {
				return 0, fmt.Errorf("write buf %d size exceeded %d", BUF_SIZE, out_len)
			}
			write_buf[out_len] = COD_ESC
			out_len++
			write_buf[out_len] = ESC_ESC
			out_len++
		default:
			if out_len+1 > BUF_SIZE {
				return 0, fmt.Errorf("write buf %d size exceeded %d", BUF_SIZE, out_len)
			}
			write_buf[out_len] = c
			out_len++
		}
	}
	if out_len+1 > BUF_SIZE {
		return 0, fmt.Errorf("write buf %d size exceeded %d", BUF_SIZE, out_len)
	}
	write_buf[out_len] = COD_END
	out_len++
	return ch_i.Write(write_buf[:out_len])
}

const (
	SLIP_NO_ERROR = 0
	SLIP_ERROR_BUFFER_OVERFLOW
	SLIP_ERROR_UNKNOWN_ESCAPED_BYTE
	SLIP_ERROR_CRC_MISMATCH
)

const {
	SLIP_STATE_NORMAL = 0,
	SLIP_STATE_ESCAPED
}

type Slip struct {
	read_cache_buf     [BUF_SIZE]byte
	len_read_cahce_buf int
	size int
	state              int
}

func (s *Slip) reset_rx(value byte) error {
}

func (s *Slip) put_byte_to_buffer(value byte) error {

}

func (s *Slip) ReadByte(value byte) error {
	var err error
	switch(s.state) {
	case SLIP_STATE_NORMAL:
			switch (value) {
			case SLIP_SPECIAL_BYTE_END:
					if (s.size >= 2) {
							//slip->descriptor->recv_message(slip->descriptor->buf,slip->size);
							fmt.Prntf("")
					}                        
					reset_rx(slip);
				break;
			case SLIP_SPECIAL_BYTE_ESC:
					slip.state = SLIP_STATE_ESCAPED;
				break;
			default:
					err = put_byte_to_buffer(slip, byte);
				break;
			}
		break
	case SLIP_STATE_ESCAPED:
			switch (byte) {
			case SLIP_ESCAPED_BYTE_END:
				byte = SLIP_SPECIAL_BYTE_END;
				break;
			case SLIP_ESCAPED_BYTE_ESC:
				byte = SLIP_SPECIAL_BYTE_ESC;
				break;
			default:
				error = SLIP_ERROR_UNKNOWN_ESCAPED_BYTE;
				reset_rx(slip);
				break;
			}

			if (error != SLIP_NO_ERROR) {
				break
			}

			err = put_byte_to_buffer(slip, byte);
		break
	}
	return err
}

func (s *Slip) SlipRead(ch_i ChannelI, buf []byte, e int) (int, int, error) {
	/*var read_cache_buf [BUF_SIZE]byte
	var len_read_cahce_buf int*/

	var buf_tmp [BUF_SIZE]byte
	var lastc byte = 0

	var len_read int = 0

	for {
		ret := 0
		if s.len_read_cahce_buf > 0 {
			ret = s.len_read_cahce_buf
			copy(buf_tmp[:ret], s.read_cache_buf[:ret])
			s.len_read_cahce_buf = 0
		} else {
			nread, err := ch_i.Read(buf_tmp[:], e)
			if err != nil {
				// ????????
				if len_read > 0 {
					s.len_read_cahce_buf = len(buf_tmp[:len_read])
					copy(s.read_cache_buf[:], buf_tmp[:len_read])
				}
				return 0, len_read, err
			} else {
				ret = nread
			}
		}
		if ret > 0 {
			for idx, c := range buf_tmp[:ret] {
				switch c {
				case COD_ESC:
					lastc = c
				case COD_END:
					lastc = 0
					left_byte := ret - (idx + 1)
					if left_byte > 0 {
						s.len_read_cahce_buf = left_byte
						copy(s.read_cache_buf[:], buf_tmp[idx+1:ret])
						return len_read, left_byte, nil
					}
					return len_read, 0, nil
				default:
					if lastc == COD_ESC {
						lastc = c
						switch c {
						case ESC_END:
							c = COD_END
						case ESC_ESC:
							c = COD_ESC
						}
					} else {
						lastc = c
					}
					if len_read > BUF_SIZE {
						len_read = 0
					}
					buf[len_read] = c
					len_read++
				}
			}
		} else {
			lastc = 0
			if len_read > 0 {
				s.len_read_cahce_buf = len(buf_tmp[:len_read])
				copy(s.read_cache_buf[:], buf_tmp[:len_read])
			}
			return 0, 0, nil
		}
	}
	// return len_read, nil
}

func SlipPack(buf []byte) []byte {
	var write_buf [BUF_SIZE]byte
	out_len := 2
	write_buf[0] = COD_FLUSH
	write_buf[1] = COD_END
	for _, c := range buf {
		switch c {
		case COD_END:
			if out_len+2 > BUF_SIZE {
				return []byte{}
			}
			write_buf[out_len] = COD_ESC
			out_len++
			write_buf[out_len] = ESC_END
			out_len++
		case COD_ESC:
			if out_len+2 > BUF_SIZE {
				return []byte{}
			}
			write_buf[out_len] = COD_ESC
			out_len++
			write_buf[out_len] = ESC_ESC
			out_len++
		default:
			if out_len+1 > BUF_SIZE {
				return []byte{}
			}
			write_buf[out_len] = c
			out_len++
		}
	}
	if out_len+1 > BUF_SIZE {
		return []byte{}
	}
	write_buf[out_len] = COD_END
	out_len++
	return write_buf[:out_len]
}

type MockSlip struct {
	MockWrite func(b []byte) (int, error)
	MockRead  func(b []byte, e int) (int, error)
}

func (m *MockSlip) Write(b []byte) (int, error) {
	return m.MockWrite(b)
}

func (m *MockSlip) Read(b []byte, e int) (int, error) {
	return m.MockRead(b, e)
}

func SlipUnpack(buf_in []byte, e int) [][]byte {
	cnt := 0
	mock := MockSlip{MockRead: func(b []byte, e int) (int, error) {
		if cnt > 0 {
			return 0, nil
		}
		copy(b, buf_in)
		cnt += 1
		return len(buf_in), nil
	}}

	ret := [][]byte{}

	slip := Slip{}

	for {
		var buff_read [256]byte
		nread, left_read, _ := slip.SlipRead(&mock, buff_read[:], e)
		// fmt.Println(nread, left_read, err)
		if nread == 0 && left_read == 0 {
			break
		}
		if nread > 0 {
			ret = append(ret, buff_read[:nread])
		}
	}
	return ret
}
