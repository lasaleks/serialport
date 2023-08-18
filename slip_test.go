package serialport

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

type Mock struct {
	MockWrite func(b []byte) (int, error)
	MockRead  func(b []byte, e int) (int, error)
	write_buf []byte
	idx_read  int
}

func (m *Mock) Write(b []byte) (int, error) {
	m.write_buf = b
	return m.MockWrite(b)
}

func (m *Mock) Read(b []byte, e int) (int, error) {
	return m.MockRead(b, e)
}

type field_ret struct {
	len int
	err error
}

func TestWrite(t *testing.T) {
	var test_os_1 [BUF_SIZE + 1]byte
	var test_os_2 [BUF_SIZE - 4]byte
	test_os_2[0] = COD_END
	test_os_2[1] = COD_ESC
	var test_os_3 [BUF_SIZE - 2]byte
	test_os_3[BUF_SIZE-3] = COD_END
	var test_os_4 [BUF_SIZE - 2]byte
	test_os_4[BUF_SIZE-3] = COD_ESC

	tt := []struct {
		caseName   string
		write_data []byte
		mock       Mock
		mock_cmp   []byte
		expected   field_ret
	}{
		{"Проверка SlipWrite 1", []byte{0x0A, 0x01, 0x00, 0xCE, 0xE4}, Mock{
			MockWrite: func(b []byte) (int, error) {
				return len(b), nil
			},
		}, []byte{0xff, 0xC0, 0x0A, 0x01, 0x00, 0xCE, 0xE4, 0xC0}, field_ret{8, nil}},
		{"Проверка SlipWrite 2 в данных символы END, ESC", []byte{0x0A, 0x01, COD_END, 0x00, 0xCE, COD_ESC, 0xE4}, Mock{
			MockWrite: func(b []byte) (int, error) {
				return len(b), nil
			},
		}, []byte{0xff, 0xC0, 0x0A, 0x01, COD_ESC, ESC_END, 0x00, 0xCE, COD_ESC, ESC_ESC, 0xE4, 0xC0}, field_ret{12, nil}},
		{"Check OverSize 1", test_os_1[:], Mock{
			MockWrite: func(b []byte) (int, error) {
				return len(b), nil
			},
		}, []byte{}, field_ret{0, fmt.Errorf("---")}},
		{"Check OverSize 2", test_os_2[:], Mock{
			MockWrite: func(b []byte) (int, error) {
				return len(b), nil
			},
		}, []byte{}, field_ret{0, fmt.Errorf("---")}},
		{"Check OverSize 3", test_os_3[:], Mock{
			MockWrite: func(b []byte) (int, error) {
				return len(b), nil
			},
		}, []byte{}, field_ret{0, fmt.Errorf("---")}},
		{"Check OverSize 4", test_os_4[:], Mock{
			MockWrite: func(b []byte) (int, error) {
				return len(b), nil
			},
		}, []byte{}, field_ret{0, fmt.Errorf("---")}},
	}

	for _, tc := range tt { //Прогоняем набор тестов, tc - сокращённо от test case
		t.Run(tc.caseName, func(t *testing.T) { //В каждой итерации цикла вызываем сабтест - t.Run
			nwrite, err := SlipWrite(&tc.mock, tc.write_data)
			if err == nil && tc.expected.err != nil {
				t.Error("Ожидается ошибка err:", tc.expected.err, " в ответе нет ошибки err:", err)
			} else if err != nil && tc.expected.err == nil {
				t.Error("Ошибка не должна быть", tc.expected.err, " в ответе получена ошибка err:", err)
			}

			if nwrite != tc.expected.len {
				t.Error()
			}
			if bytes.Equal(tc.mock.write_buf, tc.mock_cmp) == false {
				t.Error(
					"Записаны байты:", tc.mock.write_buf,
					"не соответствует", tc.mock_cmp,
				)
			}
		})
	}
}

func TestRead(t *testing.T) {
	tt := []struct {
		caseName     string
		buff_read    []byte
		compare_data []byte
		mock         Mock
		expected     field_ret
	}{
		{

			"Check SlipRead",
			make([]byte, 128),
			[]byte{0x0A, 0x01, 0x22, 0x07, 0x07, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x7F, 0xFE, 0x0A, 0x01, 0xAE, 0xB0},
			Mock{
				MockRead: func(buf []byte, e int) (int, error) {
					ret := []byte{0xC0, 0x0A, 0x01, 0x22, 0x07, 0x07, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x7F, 0xFE, 0x0A, 0x01, 0xAE, 0xB0, 0xC0}
					copy(buf, ret)
					return len(ret), nil
				},
			}, field_ret{17, nil},
		},
	}

	for _, tc := range tt { //Прогоняем набор тестов, tc - сокращённо от test case
		t.Run(tc.caseName, func(t *testing.T) { //В каждой итерации цикла вызываем сабтест - t.Run
			// t.Skip()
			slip := Slip{}
			is_read := false
			for {
				nread, left_read, err := slip.SlipRead(&tc.mock, tc.buff_read, 0)
				if err == nil && tc.expected.err != nil {
					t.Error("Ожидается ошибка err:", tc.expected.err, " в ответе нет ошибки err:", err)
				} else if err != nil && tc.expected.err == nil {
					t.Error("Ошибка не должна быть", tc.expected.err, " в ответе получена ошибка err:", err)
				}
				if nread > 0 {
					is_read = true
					if nread != tc.expected.len {
						t.Error("length excepted len", tc.expected.len, " len read:", nread)
					}
					if bytes.Equal(tc.buff_read[:nread], tc.compare_data) == false {
						t.Error("Read:", tc.buff_read[:nread], " not cmp buff:", tc.compare_data)
					}
				}
				if left_read == 0 {
					break
				}
			}
			if !is_read {
				t.Errorf("ERROR")
			}
		})
	}
}

func TestSlipPack(t *testing.T) {
	pack := []byte{0x0A, 0x01, 0x00, 0xCE, 0xE4}
	buf := SlipPack(pack)
	buf_cmp := []byte{0xff, 0xC0, 0x0A, 0x01, 0x00, 0xCE, 0xE4, 0xC0}
	if !reflect.DeepEqual(buf, buf_cmp) {
		t.Errorf("\n%X\n%X", buf, buf_cmp)
		t.Errorf("ERROR SlipPack")
	}
}

func TestSlipUnPack(t *testing.T) {
	buf := []byte{0xff, 0xC0, 0x0A, 0x01, 0x00, 0xCE, 0xE4, 0xC0}
	pack_cmp := []byte{0x0A, 0x01, 0x00, 0xCE, 0xE4}
	unpacks := SlipUnpack(buf, 0)
	cmp := false
	for _, unpack := range unpacks {
		if reflect.DeepEqual(unpack, pack_cmp) {
			cmp = true
			break
		}
	}
	if !cmp {
		t.Errorf("\n%X\n%X", pack_cmp, unpacks)
		t.Errorf("ERROR SlipUnpack")
	}
}

func TestSlip(t *testing.T) {
	buf := []byte{0xff, 0xC0, 0x0A, 0x01, 0x00, 0xCE, 0xE4, 0xC0, 0xff}
	for i := 0; i < 1020; i++ {
		buf = append(buf, 0x0A)
	}
	buf = append(buf, 0xc0)
	len_buf := len(buf)
	idx := 0
	//fmt.Println(len_buf)
	mock := Mock{MockRead: func(b []byte, e int) (int, error) {
		if idx >= len_buf {
			return 0, nil
		}
		b[0] = buf[idx]
		idx += 1
		return 1, nil
	}}
	slip := Slip{}
	var buff_read [256]byte
	for {
		nread, left_read, _ := slip.SlipRead(&mock, buff_read[:], 0)
		//fmt.Println(nread, left_read, err)
		if nread == 0 && left_read == 0 {
			break
		}
	}
}
