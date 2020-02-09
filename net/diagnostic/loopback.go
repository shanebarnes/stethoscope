package diagnostic

import (
	"syscall"

	"github.com/google/gopacket/layers"
)

func ProcessLoopback(ctx *Context, loop *layers.Loopback) error {
	ctx.dstHash.Reset()
	ctx.srcHash.Reset()

	ltx, err := ctx.GetLayerContext(loop.LayerType())
	if err != nil {
		err = processLoopback(ltx, loop)
	}
	return err
}

func processLoopback(ctx *layerContext, loop *layers.Loopback) error {
	var err error

	ctx.pkt = loop

	switch loop.Family {
	case layers.ProtocolFamilyIPv4:
		ctx.proto = uint16(layers.EthernetTypeIPv4)
	case layers.ProtocolFamilyIPv6BSD, layers.ProtocolFamilyIPv6FreeBSD, layers.ProtocolFamilyIPv6Darwin, layers.ProtocolFamilyIPv6Linux:
		ctx.proto = uint16(layers.EthernetTypeIPv6)
	default:
		err = syscall.EPROTONOSUPPORT
	}

	return err
}
