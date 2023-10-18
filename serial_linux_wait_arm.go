package serialport

import "golang.org/x/sys/unix"

func (p *Port) Wait(timeout int64) int {
	var tv unix.Timeval
	tv.Sec = int32(timeout / 1000)
	tv.Usec = (int32(timeout) - tv.Sec*1000) * 1000

	var rfds unix.FdSet
	rfds.Zero()
	rfds.Set(int(p.f.Fd()))
	ret, err := unix.Select(int(p.f.Fd())+1, &rfds, nil, nil, &tv)
	//fmt.Printf("select %d %s\n", ret, err)
	if err != nil {
		return 0
	} else if ret <= 0 {
		return 0
	}

	return 1
}
