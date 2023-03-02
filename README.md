# socketcan_go

forked from https://github.com/atuleu/golang-socketcan
1) add filter support
2) only working on linux
3) need cgo support

## Install

- go get github.com/yoursmengle/socketcan_go

## Usage

```go
package main

import (
	"fmt"
	"time"

	socketcan "github.com/yoursmengle/socketcan_go"
)

func main() {
	fmt.Println("can1 start...")
	sock_can, err := socketcan.NewCanItf("can1")
	if err != nil {
		g_str_disp.Add(fmt.Sprintln("task_can1 init error:", err))
		return
	}
	defer sock_can.Close()

	// step1, set filter for can1
	err = sock_can.AddfilterPass(recv_ids1, uint32(len(recv_ids1)))
	if err != nil {
		return
	}

	//step2, create receive thread
	go can1_recv(sock_can)

	//step3, send data periodically
	ticker := time.NewTicker(10 * time.Millisecond)
	for {
		<-ticker.C // tick 10ms

		frame, need_send := fill_can1_frame()
		if !need_send {
			continue
		}

        // send can frame
		err := sock_can.Send(frame)
		if err != nil {
            fmt.Println("send error:", err)
		}
	}
}

// can1 receive thread
func can1_recv(sock_can *socketcan.RawInterface) {
	for {
		// wait for receiving one can frame
		frame, err := sock_can.Receive()
		if err != nil {
            fmt.Println("recv error:", err)
            continue
		}

		proc_can1_frame(&frame)
	}
}

// receive can frame id list 
var recv_ids1 []uint32 = []uint32{
	0x0cf00400 | socketcan.CAN_EFF_FLAG,  // ext id 
    0x0cf00401 | socketcan.CAN_EFF_FLAG,  // ext id
	0x67b,     // std id
    0x67c,     // std id
}

func fill_can1_frame() (socketcan.CanFrame, bool) {
	frame_send := socketcan.CanFrame{}
	frame_send.ID = 0x1cf14717
	frame_send.Data = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	frame_send.Dlc = 8
	frame_send.Extended = true

	return frame_send, true
}

func proc_can1_frame(pFrame *socketcan.CanFrame) {
	if pFrame.ID == recv_ids1[0]&0x7fffffff {
		rpm := (uint16(pFrame.Data[3])<<8 | uint16(pFrame.Data[2])) / 8
        fmt.Println("rpm:", rpm)
	} else if pFrame.ID == recv_ids1[1]&0x7fffffff {
		fmt.Println("proc_can1:", pFrame)
	}
}