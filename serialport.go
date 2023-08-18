package serialport

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

var (
	// распечатка пол  данных
	LogPrintData = false
)

type InterfaceSerial interface {
	Read(b []byte, estimated_byte int) (int, error)
	Write(b []byte) (int, error)
	Connect() error
	Close() error
	Reconnect() error
	Is_connect() bool
}

// интерфайсе управления приемником/передатчиком
type ICtrlTxRxEn interface {
	TxEn(value bool) error
	RxEn(value bool) error
}

const (
	_ = iota
	type_serial_stty
	type_serial_udp
)

type SerialPort struct {
	type_serial int // 1 -type_source_stty,  type_source_udp
	stty        *Port
	config_stty struct {
		device            string
		baud              int
		wait              time.Duration
		typeRS            int // RS 233/422/485
		oneSymbolDuration int // длительность одного символа в микросекундах
	}
	ctrlEn ICtrlTxRxEn

	udp_wg          sync.WaitGroup
	udp_con         *net.UDPConn
	udp_listen_addr *net.UDPAddr
	udp_dest_addr   *net.UDPAddr
	ch_udp_recv     chan []byte
	config_udp      struct {
		host        string
		listen_port uint16
		dest_port   uint16
		wait        time.Duration
	}
	udp_read_timeout bool
}

func NewSerialPortUdp(host string, listen_port uint16, dest_port uint16, wait time.Duration) (*SerialPort, error) {
	serial := SerialPort{type_serial: type_serial_udp}
	serial.config_udp.dest_port = dest_port
	serial.config_udp.listen_port = listen_port
	serial.config_udp.host = host
	serial.config_udp.wait = wait
	var err error
	serial.udp_listen_addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", listen_port))
	if err != nil {
		return nil, err
	}
	serial.udp_dest_addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, dest_port))
	if err != nil {
		return nil, err
	}
	return &serial, nil
}

func NewSerialPortStty(device string, baud int, wait time.Duration, typeRS int, ctrlEn ICtrlTxRxEn) (*SerialPort, error) {
	serial := SerialPort{type_serial: type_serial_stty}
	serial.config_stty.device = device
	serial.config_stty.baud = baud
	serial.config_stty.wait = wait
	serial.config_stty.typeRS = typeRS
	serial.config_stty.oneSymbolDuration = 10000000 / baud
	serial.ctrlEn = ctrlEn
	return &serial, nil
}

func (s *SerialPort) Write(buf []byte) (int, error) {
	if LogPrintData {
		fmt.Printf("Write:%x\n", buf)
	}
	switch s.type_serial {
	case type_serial_stty:
		s.stty.Flush()
		if LogPrintData {
			fmt.Printf("Stty Write:%s %x\n", s.udp_dest_addr, buf)
		}
		if s.config_stty.typeRS == 485 {
			s.ctrlEn.TxEn(true)
			//time.Sleep(time.Microsecond * 50)
			s.ctrlEn.RxEn(true)
		}

		len_write, err := s.stty.Write(buf)
		if err != nil {
			return 0, err
		}

		time.Sleep(time.Microsecond * time.Duration(s.config_stty.oneSymbolDuration*len_write))

		if s.config_stty.typeRS == 485 {
			s.ctrlEn.TxEn(false)
			//time.Sleep(time.Microsecond * 50)
			s.ctrlEn.RxEn(false)
		}
		return len_write, nil
	case type_serial_udp:
		//print_time(time.Now().UnixNano())
		if LogPrintData {
			fmt.Printf("Udp Write:%s %x\n", s.udp_dest_addr, buf)
		}
		len_write, err := s.udp_con.WriteToUDP(buf, s.udp_dest_addr)
		if err != nil {
			return 0, err
		}
		return len_write, nil
	}
	return 0, fmt.Errorf("error type_source")
}

func print_time(unix_nano int64) {
	time_msec := unix_nano / int64(time.Millisecond)
	time_sec := unix_nano / int64(time.Second)
	fmt.Printf("time:%d.%d\n", time_sec, time_msec-time_sec*1000)
}

func (s *SerialPort) Read(buf []byte, estimated_byte int) (int, error) {
	switch s.type_serial {
	case type_serial_stty:
		if estimated_byte > 0 {
			//fmt.Println("sleep", time.Microsecond*time.Duration(s.config_stty.oneSymbolDuration*estimated_byte), estimated_byte)
			time.Sleep(time.Microsecond * time.Duration(s.config_stty.oneSymbolDuration*estimated_byte))
		}
		if s.stty.Wait(s.config_stty.wait.Milliseconds()) == 0 {
			return 0, fmt.Errorf("timeout")
		}
		l, err := s.stty.Read(buf)
		if LogPrintData {
			log.Printf("Stty Read:%x\n", buf[0:l])
		}
		if err != nil {
			return 0, err
		}
		return l, nil
	case type_serial_udp:

		s.udp_con.SetReadDeadline(time.Now().Add(s.config_udp.wait))
		read_len := 0
		for {
			n, err := s.udp_con.Read(buf[read_len:])
			if LogPrintData {
				fmt.Printf("s.udp_con.Read:%d, %s\n", n, err)
			}
			if err != nil {
				if e, ok := err.(net.Error); !ok || !e.Timeout() {
					// handle error, it's not a timeout
				}
				break
			}
			if n == 0 {
				break
			}
			read_len += n
			// do something with packet here
		}
		return read_len, nil

		/*		select {
				case <-time.After(s.config_udp.wait):
					return 0, fmt.Errorf("timeout")
					break
				case recv_buff := <-s.ch_udp_recv:
					s.udp_con.SetReadDeadline((time.))
					copy(buf, recv_buff)
					if LogPrintData {
						log.Printf("Udp Read:%x\n", buf[0:len(recv_buff)])
					}
					return len(recv_buff), nil
				}*/
	}
	return 0, fmt.Errorf("error type_source")
}

func (s *SerialPort) recv_udp(wg *sync.WaitGroup) {
	defer s.udp_wg.Done()
	buf_in := make([]byte, 1024)

	for {
		len_msg, _, err := s.udp_con.ReadFromUDP(buf_in)
		if err != nil {
			break
		}
		s.ch_udp_recv <- buf_in[:len_msg]
	}
}

func (s *SerialPort) Connect() error {
	switch s.type_serial {
	case type_serial_stty:
		if s.stty != nil {
			s.Close()
		}
		c := &Config{
			Name:        s.config_stty.device,
			Baud:        s.config_stty.baud,
			ReadTimeout: s.config_stty.wait,
		}
		stty, err := OpenPort(c)
		if err != nil {
			return err
		}
		s.stty = stty
	case type_serial_udp:
		var err error
		if s.udp_con != nil {
			s.udp_con.Close()
		}
		s.udp_con, err = net.ListenUDP("udp", s.udp_listen_addr)
		//s.udp_con.SetReadDeadline(time.Now().Add(time.Millisecond * 1000))
		s.ch_udp_recv = make(chan []byte)
		if err != nil {
			return err
		}
		//s.udp_wg.Add(1)
		//go s.recv_udp(&s.udp_wg)
	default:
		return fmt.Errorf("error type_source %d", s.type_serial)
	}
	return nil
}

func (s *SerialPort) Close() error {
	switch s.type_serial {
	case type_serial_stty:
		if s.stty != nil {
			return s.stty.Close()
		}
	case type_serial_udp:
		if s.udp_con != nil {
			s.udp_con.Close()
			close(s.ch_udp_recv)
			s.udp_wg.Wait()
		}
	}
	return nil
}

func (s *SerialPort) Reconnect() error {
	s.Close()
	s.Connect()
	return nil
}

func (s *SerialPort) Is_connect() bool {
	if s.type_serial == type_serial_stty {
		if s.stty != nil {
			return true
		}
	}
	return false
}
