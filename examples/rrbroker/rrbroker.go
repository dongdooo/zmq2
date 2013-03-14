//
//  Simple request-reply broker.
//
package main

import (
	zmq "github.com/pebbe/zmq2"
)

func main() {
	//  Prepare our sockets
	frontend, _ := zmq.NewSocket(zmq.ROUTER)
	defer frontend.Close()
	backend, _ := zmq.NewSocket(zmq.DEALER)
	defer backend.Close()
	frontend.Bind("tcp://*:5559")
	backend.Bind("tcp://*:5560")

	//  Initialize poll set
	poller := zmq.NewPoller()
	poller.Add(frontend, zmq.POLLIN)
	poller.Add(backend, zmq.POLLIN)

	//  Switch messages between sockets
	for {
		sockets, _ := poller.Poll(-1)
		for socket := range sockets {
			switch socket {
			case frontend:
				for {
					msg, _ := frontend.Recv(0)
					if more, _ := frontend.GetRcvmore(); more {
						backend.Send(msg, zmq.SNDMORE)
					} else {
						backend.Send(msg, 0)
						break
					}
				}
			case backend:
				for {
					msg, _ := backend.Recv(0)
					if more, _ := backend.GetRcvmore(); more {
						frontend.Send(msg, zmq.SNDMORE)
					} else {
						frontend.Send(msg, 0)
						break
					}
				}
			}
		}
	}
}
