package diagnostic

import (
	"encoding/binary"
	"fmt"
	//"hash/fnv"
	//"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type tcpOp func(*layerContext, *layers.TCP) error
var tcpOps = []tcpOp {
	processTCP,
	processTcpConnect,
	processTcpClose,
}

// Flow
func ProcessTCPFlow(ctx *Context) error {
	return nil
}

// Process Layer
// Process Flow

type TcpDiagnostic struct {

}

// Packet
//func ProcessTCP(ctx *Context, tcp *layers.TCP) error {
func (t *TcpDiagnostic) ProcessLayer(ctx *Context, lyr gopacket.DecodingLayer) error {
	tcp, _ := lyr.(*layers.TCP)
	flow := tcp.TransportFlow()
	src, dst := flow.Endpoints()
	ctx.dstHash.Write(dst.Raw())
	ctx.srcHash.Write(src.Raw())

	proto := []byte{byte(layers.IPProtocolTCP)}
	ctx.dstHash.Write(proto)
	ctx.srcHash.Write(proto)

	lctx, err := ctx.GetLayerContext(tcp.LayerType())
	// perf: We may only want to do this on the first packet
	lctx.src = src.String()
	lctx.dst = dst.String()
        // perf:
	lctx.seq = tcp.Seq
	lctx.ack = tcp.Ack
	if err == nil {
		for _, op := range tcpOps {
			err = op(lctx, tcp)
		}
	}
	return err
}

func processTCP(ctx *layerContext, tcp *layers.TCP) error {
	ctx.pkt = tcp
	ctx.flow = tcp.TransportFlow()
	ctx.hdrLen = uint16(tcp.DataOffset)*4
	if ctx.lowerLayer != nil {
		ctx.payLen = ctx.lowerLayer.payLen - ctx.hdrLen
	}
	getTcpOpts(ctx, tcp)
	//fmt.Println("seq: ", tcp.Seq, ", ack: ", tcp.Ack, "hdrLen: ", ctx.hdrLen, ", payLen: ", ctx.payLen)
	//fmt.Println("tcp: ", tcp.TransportFlow())
	//ctx.l4.src = strconv.FormatUint(uint64(lyr.SrcPort), 10)
	//ctx.l4.dst = strconv.FormatUint(uint64(lyr.DstPort), 10)

	/*if _, ok := layers.TCPPortNames[lyr.DstPort]; ok {
		ctx.l4.proto = uint16(lyr.DstPort)
	} else if _, ok := layers.TCPPortNames[lyr.SrcPort]; ok {
		ctx.l4.proto = uint16(lyr.SrcPort)
	} else {
		ctx.l4.proto = 0
	}*/

	//calcHash(ctx, tcp)

	return nil
}

func getTcpOpts(ctx *layerContext, tcp *layers.TCP) {
	for _, opt := range tcp.Options {
		switch opt.OptionType {
		case layers.TCPOptionKindMSS:
			ctx.mss = uint32(binary.BigEndian.Uint16(opt.OptionData[:2]))
		case layers.TCPOptionKindWindowScale:
			ctx.ws = uint32(opt.OptionData[0])
			ctx.mws = (uint32)(tcp.Window) << ctx.ws
		}
	}
}

func processTcpConnect(ctx *layerContext, tcp *layers.TCP) error {
	var err error
// Could all be done later
	if tcp.SYN && !tcp.ACK {
		// Src flow started ... populate address strings
		// Set context open flag
		ctx.openPkt = true
		//fmt.Println("Src flow started: mss=", ctx.mss, ", win=", ctx.mws)
	} else if tcp.SYN && tcp.ACK {
		ctx.openPkt = true
		//fmt.Println("Dst flow started: mss=", ctx.mss, ", win=", ctx.mws)
	}

	return err
}

func processTcpClose(ctx *layerContext, tcp *layers.TCP) error {
	var err error

	if tcp.FIN {
		ctx.closePkt = true
		fmt.Println("Flow finished")
	} else if tcp.RST {
		ctx.closePkt = true
		fmt.Println("Flow reset")
	}

	return err
}

// src address of first TCP SYN
func IsSrc(ctx *layerContext, tcp *layers.TCP) bool {
	var isSrc bool = false

	// Check if context already knows about TCP flow
	//if ctx.l4.src == tcp.SrcPort {
	//} else if tcp.SYN && !tcp.ACK { // Check if beginning of stream
	//}

	// Ignore
	return isSrc
}

func calcHash(ctx *layerContext, tcp *layers.TCP) error {
	/*h1 := fnv.New64a()
	h2 := fnv.New64a()
	h3 := fnv.New64a()
	h4 := fnv.New64a()

	saddr, daddr := ctx.l3.flow.Endpoints()
	sport, dport := ctx.l4.flow.Endpoints()

	h1.Write(saddr.Raw())
	h2.Write(daddr.Raw())
	h3.Write(sport.Raw())
	h4.Write(dport.Raw())
fmt.Println("type: ", ctx.l4.flow.EndpointType())
	// Should most mutable be added first? port instead of address?
	hash := fnv.New64a()
	hash.Write(saddr.Raw())
	hash.Write(sport.Raw())
	//hash.Write(ctx.l3.flow.EndpointType())
	//hash.Write(ctx.l4.flow.EndpointType())
	srcSum := hash.Sum64()
	ctx.l4.flowHash = srcSum

	hash.Reset()
	hash.Write(daddr.Raw())
	hash.Write(dport.Raw())
	//hash.Write(uint64(ctx.l3.flow.EndpointType()))
	//hash.Write(uint64(ctx.l4.flow.EndpointType()))
	ctx.l4.hash = srcSum + hash.Sum64() // commutative

	//ctx.l4.flowHash = h3.Sum64() + h4.Sum64()
fmt.Printf("%v > %v: key %v flow %v\n", sport, dport, ctx.l4.hash, ctx.l4.flowHash)

	//sum := h1.Sum64() + h2.Sum64()
	//sum ^= uint64(ctx.l3.flow.EndpointType())
        //sum *= 1099511628211 //fnvPrime

	//sum += h3.Sum64() + h4.Sum64()

	// Variant of Fowler-Noll-Vo hashing, and is guaranteed to collide
	// with its reverse flow. (see github.com/google/gopacket/flows.go)
	//sum ^= uint64(ctx.l4.flow.EndpointType())
        //sum *= 1099511628211 //fnvPrime

	//ctx.l4.hash = sum
	//fmt.Println("Hash: ", sum)
*/
	return nil
}
