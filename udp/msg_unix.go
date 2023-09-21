package udp

import (
	"golang.org/x/sys/unix"
)

type Msg struct {
	cmsgs []unix.SocketControlMessage
}

func parseMsg(buf []byte) (Msg, error) {
	cmsgs, err := unix.ParseSocketControlMessage(buf)
	if err != nil {
		return Msg{}, err
	}
	return Msg{cmsgs: cmsgs}, nil
}
