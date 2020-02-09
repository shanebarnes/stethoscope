package diagnostic

import (
	"syscall"

	"github.com/google/gopacket/layers"
)

func ProcessIPv6(ctx *Context, ip6 *layers.IPv6) error {
	lctx, err := ctx.GetLayerContext(ip6.LayerType())
	if err != nil {
		processIPv6(lctx, ip6)
	}
	return err
}

func processIPv6(ctx *layerContext, ip6 *layers.IPv6) error {
	var err error

	switch ip6.NextHeader {
	case layers.IPProtocolTCP, layers.IPProtocolUDP:
		ctx.pkt = ip6
		ctx.flow = ip6.NetworkFlow()
		//ctx.l3.src = lyr.SrcIP.String()
		//ctx.l3.dst = lyr.DstIP.String()
		ctx.proto = uint16(ip6.NextHeader)
	default:
		err = syscall.EPROTONOSUPPORT
	}

	return err
}
