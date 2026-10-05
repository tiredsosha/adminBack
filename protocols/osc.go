package protocols

import (
	"github.com/hypebeast/go-osc/osc"
	"github.com/tiredsosha/admin/tools/logger"
)

func SendOsc(ip string, port int, address, data string) {
	client := osc.NewClient(ip, port)
	msg := osc.NewMessage(address)
	msg.Append(data)
	client.Send(msg)
	logger.Debug.Printf("osc msg - %q tis sent to %q:%d, %q\n", data, ip, port, address)
}

func SendOscInt(ip string, port int, address string, data int) {
	client := osc.NewClient(ip, port)
	msg := osc.NewMessage(address)
	msg.Append(int32(data))
	client.Send(msg)
	logger.Debug.Printf("osc msg - %q tis sent to %q:%d, %q\n", data, ip, port, address)
}
