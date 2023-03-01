package socketcan

import "golang.org/x/sys/unix"

const (
	CAN_EFF_MASK uint32 = unix.CAN_EFF_MASK
	CAN_SFF_MASK uint32 = unix.CAN_SFF_MASK
	CAN_RTR_FLAG uint32 = unix.CAN_RTR_FLAG
	CAN_EFF_FLAG uint32 = unix.CAN_EFF_FLAG
)
