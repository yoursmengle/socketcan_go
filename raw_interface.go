package socketcan

/*
#include "filter.c"
*/
import "C"

import (
	"errors"
	"fmt"
	"os/exec"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	BAUD_1M   = 1000000
	BAUD_500K = 500000
	BAUD_250K = 250000
	BAUD_125K = 125000
	BAUD_100K = 100000
	BAUD_50K  = 50000
)

type RawInterface struct {
	fd   int
	name string
}

func (itf *RawInterface) getIfIndex(ifName string) (int, error) {
	ifNameRaw, err := unix.ByteSliceFromString(ifName)
	if err != nil {
		return 0, err
	}
	if len(ifNameRaw) > unix.IFNAMSIZ {
		return 0, fmt.Errorf("Maximum ifname length is %d characters", unix.IFNAMSIZ)
	}

	type ifreq struct {
		Name  [unix.IFNAMSIZ]byte
		Index int
	}
	var ifReq ifreq
	copy(ifReq.Name[:], ifNameRaw)
	_, _, errno := unix.Syscall(unix.SYS_IOCTL,
		uintptr(itf.fd),
		unix.SIOCGIFINDEX,
		uintptr(unsafe.Pointer(&ifReq)))
	if errno != 0 {
		return 0, fmt.Errorf("ioctl: %v", errno)
	}

	return ifReq.Index, nil
}

func NewCanItf(ifName string) (*RawInterface, error) {
	itf := &RawInterface{name: ifName}
	var err error
	itf.fd, err = unix.Socket(unix.AF_CAN, unix.SOCK_RAW, unix.CAN_RAW)
	if err != nil {
		return nil, err
	}

	ifIndex, err := itf.getIfIndex(ifName)
	if err != nil {
		return itf, err
	}

	addr := &unix.SockaddrCAN{Ifindex: ifIndex}
	err = unix.Bind(itf.fd, addr)

	return itf, err
}

func (itf *RawInterface) Close() error {
	return unix.Close(itf.fd)
}

func (itf *RawInterface) Send(f *CanFrame) error {
	frameBytes := make([]byte, 16)
	f.putID(frameBytes)
	frameBytes[4] = f.Dlc
	copy(frameBytes[8:], f.Data)
	_, err := unix.Write(itf.fd, frameBytes)
	return err
}

func (itf *RawInterface) Receive() (*CanFrame, error) {
	f := CanFrame{Data: make([]byte, 8)}
	frameBytes := make([]byte, 16)
	_, err := unix.Read(itf.fd, frameBytes)
	if err != nil {
		return &f, err
	}

	f.getID(frameBytes)
	f.Dlc = frameBytes[4]
	copy(f.Data, frameBytes[8:])

	return &f, nil
}

func (itf *RawInterface) up() error {
	return exec.Command("ifconfig", itf.name, "up").Run()
}

func (itf *RawInterface) down() error {
	return exec.Command("ifconfig", itf.name, "down").Run()
}

func (itf *RawInterface) AddfilterPass(recv_ids []uint32, len uint32) error {
	ptr := unsafe.Pointer(&recv_ids[0])
	succ := C.rcvFiltersSet(C.int(itf.fd), ptr, C.uint(len), C.CAN_FILTER_PASS)
	if succ == 0 {
		return nil
	}

	return errors.New("can filter failed")
}

func (itf *RawInterface) AddfilterId(id uint, len uint) error {
	succ := C.rcvFiltersID(C.int(itf.fd), C.uint(id), C.uint(len))
	if succ == 0 {
		return nil
	}
	return errors.New("can filter failed")
}

func (itf *RawInterface) SetBaud(baud uint32) error {
	var err error

	err = itf.down()
	if err != nil {
		return err
	}

	exec.Command("ip", "link", "set", itf.name, "type", "can", "bitrate", fmt.Sprintf("%d", baud)).Run()
	if err != nil {
		return err
	}

	return itf.up()
}

func (itf *RawInterface) SetTxQueueLen(size uint32) error {
	var err error

	err = itf.down()
	if err != nil {
		return err
	}

	exec.Command("ifconfig", itf.name, "txqueuelen", fmt.Sprintf("%d", size)).Run()
	if err != nil {
		return err
	}

	return itf.up()

}
