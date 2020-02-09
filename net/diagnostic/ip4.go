package diagnostic

import (
	//"fmt"
	"syscall"

	"github.com/google/gopacket/layers"
)

func ProcessIPv4(ctx *Context, ip4 *layers.IPv4) error {
	flow := ip4.NetworkFlow()
	src, dst := flow.Endpoints()
	ctx.dstHash.Write(dst.Raw())
	ctx.srcHash.Write(src.Raw())

	proto := []byte{byte(layers.IPProtocolIPv4)}
	ctx.dstHash.Write(proto)
	ctx.srcHash.Write(proto)

	lctx, err := ctx.GetLayerContext(ip4.LayerType())
	if err == nil {
		processIPv4(lctx, ip4)
	}
	return err
}

func processIPv4(ctx *layerContext, ip4 *layers.IPv4) error {
	var err error

	switch ip4.Protocol {
	case layers.IPProtocolTCP, layers.IPProtocolUDP:
		ctx.pkt = ip4
		ctx.flow = ip4.NetworkFlow()
	        // perf: We may only want to do this on the first packet
		ctx.src = ip4.SrcIP.String()
		ctx.dst = ip4.DstIP.String()
                // perf
		ctx.proto = uint16(ip4.Protocol)
		ctx.hdrLen = uint16(ip4.IHL)*4
		ctx.payLen = ip4.Length-ctx.hdrLen
		//fmt.Println("ip4 hdrLen: ", ctx.hdrLen, ", payLen: ", ctx.payLen)
	default:
		err = syscall.EPROTONOSUPPORT
	}

	return err
}
