package diagnostic

import (
	"hash"
	"hash/fnv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Each layer needs packet context (parsed or partially parsed packet, capture info) and flow context
type layerContext struct {
	openPkt, closePkt bool
	pkt      gopacket.Layer
	info     *LayerInfo
	hasher   hash.Hash64
	seq, ack uint32
	dst, src string
	flow     gopacket.Flow
	hash     uint64
	flowHash uint64
	hashDst  uint64
	hashSrc  uint64
	proto    uint16 // Used for layers.EthernetType/ProtocolFamily, layers.IPProtocol/layers.NextHeader, and layers.TCPPort/UDPPort
	hdrLen   uint16
	payLen   uint16
	mss      uint32 // Maximum Segment Size
	ws       uint32 // Window Scale
	mws      uint32 // Window Scale
	lowerLayer   *layerContext // TCP payload calcuation requires IP header
}

type packetContext struct {

}

//flowDecoder
//flowEncoder

type Context struct {
	capInfo  *gopacket.CaptureInfo
	dstSum, srcSum uint64
	dstHash, srcHash hash.Hash64
	//f3 gopacket.Flow
	//f4 gopacket.Flow
	l2 layerContext
	l3 layerContext
	l4 layerContext
	l7 layerContext
}

func (c *Context) Finalize() {
	c.srcSum = c.srcHash.Sum64()
	c.dstSum = c.dstHash.Sum64()
}

func (c *Context) IsClosePkt(l gopacket.LayerType) (bool, error) {
	var isClose bool = false

	lctx, err := c.GetLayerContext(l)
	if err == nil {
		 isClose = lctx.closePkt
	}

	return isClose, err
}

func (c *Context) IsOpenPkt(l gopacket.LayerType) (bool, error) {
	var isOpen bool = false

	lctx, err := c.GetLayerContext(l)
	if err == nil {
		 isOpen = lctx.openPkt
	}

	return isOpen, err
}

func (c *Context) GetLayerContext(l gopacket.LayerType) (*layerContext, error) {
	var ctx *layerContext
	var err error

	switch l {
	case layers.LayerTypeEthernet, layers.LayerTypeLoopback:
		ctx = &c.l2
	case layers.LayerTypeIPv4, layers.LayerTypeIPv6:
		ctx = &c.l3
	case layers.LayerTypeTCP, layers.LayerTypeUDP:
		ctx = &c.l4
	default:
	}

	return ctx, err
}

func NewContext(ci *gopacket.CaptureInfo) *Context {
	var ctx Context

	ctx.srcHash = fnv.New64a()
	ctx.dstHash = fnv.New64a()

	ctx.l3.lowerLayer = &ctx.l2
	ctx.l4.lowerLayer = &ctx.l3

	ctx.capInfo = ci

	return &ctx
}
