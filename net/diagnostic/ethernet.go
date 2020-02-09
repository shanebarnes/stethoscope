package diagnostic

import (
	"syscall"

	"github.com/google/gopacket/layers"
)

func ProcessEthernet(ctx *Context, eth *layers.Ethernet) error {
	ctx.dstHash.Reset()
	ctx.srcHash.Reset()

	ltx, err := ctx.GetLayerContext(eth.LayerType())
	if err == nil {
		err = processEthernet(ltx, eth)
	}
	return err
}

func processEthernet(ctx *layerContext, eth *layers.Ethernet) error {
	var err error

	switch eth.EthernetType {
	case layers.EthernetTypeIPv4, layers.EthernetTypeIPv6:
		ctx.pkt = eth
	        // perf: We may only want to do this on the first packet
		ctx.src = eth.SrcMAC.String()
		ctx.dst = eth.DstMAC.String()
                // perf
		ctx.proto = uint16(eth.EthernetType)
	default:
		err = syscall.EPROTONOSUPPORT
	}

	return err
}
