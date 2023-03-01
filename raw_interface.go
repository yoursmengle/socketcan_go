package socketcan

/*
#include <stdio.h>
#include <stdlib.h>
#include <sys/ioctl.h>
#include <sys/socket.h>
#include <stdint.h>
#include <ctype.h>
#include <errno.h>
#include <libgen.h>
#include <signal.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <net/if.h>
#include <sys/epoll.h>
#include <sys/ioctl.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <linux/can/raw.h>

#define CAN_FILTER_PASS     0x01    //过滤方式-通过
#define CAN_FILTER_REJECT   0x02    //过滤方式-拒绝

int rcvFiltersSet(int canfd, const uint canId, const uint filterType)
{
    if(canfd <= 0)	//canfd就不用解释了…
        return -1;

    if(0 == canId){
        setsockopt(canfd, SOL_CAN_RAW, CAN_RAW_FILTER, NULL, 0);    //不需要接收任何报文
        return 0;
    }

    struct can_filter rfilter;

    if(filterType & CAN_FILTER_PASS){
        rfilter.can_id = canId;
    } else {
        rfilter.can_id = canId | CAN_INV_FILTER;
    }
	if(canId &0x80000000) {
        rfilter.can_mask = 0x1fffffff; 
	} else {
		rfilter.can_mask = 0x7ff;
	}

    if(filterType & CAN_FILTER_REJECT){
        int join_filter = 1;
        setsockopt(canfd, SOL_CAN_RAW, CAN_RAW_JOIN_FILTERS, &join_filter, sizeof(join_filter));
    }
    setsockopt(canfd, SOL_CAN_RAW, CAN_RAW_FILTER, &rfilter, sizeof(rfilter));
    return 0;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
	"errors"
	"golang.org/x/sys/unix"
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

func NewCanItf(ifName string) (RawInterface, error) {
	itf := RawInterface{name: ifName}
	var err error
	itf.fd, err = unix.Socket(unix.AF_CAN, unix.SOCK_RAW, unix.CAN_RAW)
	if err != nil {
		return nil, err
	}

	ifIndex, err := getIfIndex(itf, ifName)
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

func (itf *RawInterface) Send(f CanFrame) error {
	frameBytes := make([]byte, 16)
	f.putID(frameBytes)
	frameBytes[4] = f.Dlc
	copy(frameBytes[8:], f.Data)
	_, err := unix.Write(itf.fd, frameBytes)
	return err
}

func (itf *RawInterface) Receive() (CanFrame, error) {
	f := CanFrame{Data: make([]byte, 8)}
	frameBytes := make([]byte, 16)
	_, err := unix.Read(itf.fd, frameBytes)
	if err != nil {
		return f, err
	}

	f.getID(frameBytes)
	f.Dlc = frameBytes[4]
	copy(f.Data, frameBytes[8:])

	return f, nil
}

func (itf *RawInterface) AddfilterPass(canid_pass uint) error {
	succ := C.rcvFiltersSet(C.int(itf.fd), C.uint(canid_pass), C.CAN_FILTER_PASS)
	if succ == 0 {
		return nil
	}

	return errors.New("can filter failed")
}
