package diagnostic

import (
	//"strconv"

	"github.com/google/gopacket/layers"
)

func ProcessUDP(ctx *Context, udp *layers.UDP) error {
	flow := udp.TransportFlow()
	src, dst := flow.Endpoints()
	ctx.dstHash.Write(dst.Raw())
	ctx.srcHash.Write(src.Raw())

	proto := []byte{byte(layers.IPProtocolUDP)}
	ctx.dstHash.Write(proto)
	ctx.srcHash.Write(proto)

	lctx, err := ctx.GetLayerContext(udp.LayerType())
	// perf: We may only want to do this on the first packet
	lctx.src = src.String()
	lctx.dst = dst.String()
        // perf:

	lctx.pkt = udp
	lctx.flow = udp.TransportFlow()
	lctx.hdrLen = 8
	lctx.payLen = uint16(udp.Length) - lctx.hdrLen

	return err
	//ctx.flow = lyr.TransportFlow()
	//ctx.l4.src = strconv.FormatUint(uint64(lyr.SrcPort), 10)
	//ctx.l4.dst = strconv.FormatUint(uint64(lyr.DstPort), 10)
	//ctx.proto = uint16(lyr.SrcPort)

	//return nil
}
