package zmq2

import (
	"fmt"
	"syscall"
)

/*
Returns true if err is not nil, and it's the result of an interrupted signal call.

Example usage:

    for {
        client.Send("HELLO", 0)
        reply, err := client.Recv(0)
        if err != nil {
            if zmq.IsInterrupted(err) {
                break
            } else {
                log.Fatal(err)
            }
        }
        fmt.Println(reply)
    }

See also: examples/interrupt.go
*/
func IsInterrupted(err error) bool {
	errno, ok := err.(syscall.Errno)
	if ok && errno == syscall.EINTR {
		return true
	}
	return false
}

/*
Send multi-part message on socket.

Any `[]string' or `[][]byte' is split into separate `string's or `[]byte's

Any other part that isn't a `string' or `[]byte' is converted
to `string' with `fmt.Sprintf("%v", part)'.

Returns total bytes sent.
*/
func (soc *Socket) SendMessage(parts ...interface{}) (total int, err error) {
	// TODO: make this faster

	// flatten first, just in case the last part may be an empty []string or [][]byte
	pp := make([]interface{}, 0)
	for _, p := range parts {
		switch t := p.(type) {
		case []string:
			for _, s := range t {
				pp = append(pp, s)
			}
		case [][]byte:
			for _, b := range t {
				pp = append(pp, b)
			}
		default:
			pp = append(pp, t)
		}
	}

	n := len(pp)
	if n == 0 {
		return
	}
	opt := SNDMORE
	for i, p := range pp {
		if i == n-1 {
			opt = 0
		}
		switch t := p.(type) {
		case string:
			j, e := soc.Send(t, opt)
			if e == nil {
				total += j
			} else {
				return -1, e
			}
		case []byte:
			j, e := soc.SendBytes(t, opt)
			if e == nil {
				total += j
			} else {
				return -1, e
			}
		default:
			j, e := soc.Send(fmt.Sprintf("%v", t), opt)
			if e == nil {
				total += j
			} else {
				return -1, e
			}
		}
	}
	return
}

/*
Receive parts as message from socket.

Returns last non-nil error code.
*/
func (soc *Socket) RecvMessage(flags Flag) (msg []string, err error) {
	msg = make([]string, 0)
	for {
		s, e := soc.Recv(flags)
		if e == nil {
			msg = append(msg, s)
		} else {
			return msg[0:0], e
		}
		more, e := soc.GetRcvmore()
		if e == nil {
			if !more {
				break
			}
		} else {
			return msg[0:0], e
		}
	}
	return
}

/*
Receive parts as message from socket.

Returns last non-nil error code.
*/
func (soc *Socket) RecvMessageBytes(flags Flag) (msg [][]byte, err error) {
	msg = make([][]byte, 0)
	for {
		b, e := soc.RecvBytes(flags)
		if e == nil {
			msg = append(msg, b)
		} else {
			return msg[0:0], e
		}
		more, e := soc.GetRcvmore()
		if e == nil {
			if !more {
				break
			}
		} else {
			return msg[0:0], e
		}
	}
	return
}
