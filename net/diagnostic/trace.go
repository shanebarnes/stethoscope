package diagnostic

import (
	//"fmt"
	"syscall"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/shanebarnes/goto/units"
)

type FlowInfo struct {
	Addr     string         `json:"address"`
	DupBytes uint64         `json:"-"` // Missed packets?
	OooBytes uint64         `json:"-"`
	HdrBytes uint64         `json:"-"`
	PayBytes uint64         `json:"-"`
	FlowHash uint64         `json:"-"`
	context layerContext
	first time.Time
	last time.Time
	window map[uint32]uint32
	WinSize uint32           `json:"winSize"`
	mws uint32                // max window size
	winHead uint32
	winTail uint32
	lastSeq uint32
	DupPackets uint64        `json:"-"`
	Packets uint64           `json:"packets"`
	Bps     uint64           `json:"-"`
	Per     uint64           `json:"-"`
	Pps     uint64           `json:"pps"`
	Rtt     time.Duration    `json:"-"`
	PrettyBps string         `json:"bps"`
	PrettyHdrBytes string    `json:"hdrBytes"`
	PrettyPayBytes string    `json:"appBytes"`
	PrettyRtt string         `json:"rtt"`
}

type LayerInfo struct {
	Dur time.Duration `json:"-"`
	Dst FlowInfo      `json:"dstFlow"` // Flow information for packets originating from destination
	Src FlowInfo      `json:"srcFlow"` // Flow information for packets originating from source
	PrettyDur string  `json:"duration"`
}

type trace struct {
	borrow bool
	layer map[gopacket.LayerType]*LayerInfo  // stack?
}

type TraceTable struct {
	     sync.RWMutex
	tbl map[uint64]*trace
}

func (t *TraceTable) GetLayerInfo(ctx *Context, lt gopacket.LayerType) (*LayerInfo, error) {
	var li LayerInfo

	t.RLock()

	trc, err := t.getTraceInfo(ctx)
	if err == nil {
		if trc.layer == nil {
			err = syscall.ENOENT
		} else {
			info, ok := trc.layer[lt]
			if ok {
				li = *info
				//idx := len(*arr) - 1
				//li = (*arr)[idx]
			}
		}
	}
//cons := len(t.tbl)
	t.RUnlock()
//fmt.Println("Table size: ", cons, ", retrieving layer info for hash: ", ctx.srcSum+ctx.dstSum, ", flow src hash: ", ctx.srcSum, "flow dst hash: ", ctx.dstSum)

	li.PrettyDur = units.ToTimeString(float64(li.Dur) / float64(time.Second))
	li.Src.PrettyHdrBytes = units.ToMetricString(float64(li.Src.HdrBytes), 3, "", "B")
	li.Dst.PrettyHdrBytes = units.ToMetricString(float64(li.Dst.HdrBytes), 3, "", "B")
	li.Src.PrettyPayBytes = units.ToMetricString(float64(li.Src.PayBytes), 3, "", "B")
	li.Dst.PrettyPayBytes = units.ToMetricString(float64(li.Dst.PayBytes), 3, "", "B")
	li.Src.PrettyRtt = units.ToMetricString(float64(li.Src.Rtt) / float64(time.Second), 3, "", "s")
	li.Dst.PrettyRtt = units.ToMetricString(float64(li.Dst.Rtt) / float64(time.Second), 3, "", "s")
	if li.Src.Packets > 0 {
		li.Src.Per = li.Src.DupPackets * 1000 / li.Src.Packets
	}

	if li.Dst.Packets > 0 {
		li.Dst.Per = li.Dst.DupPackets * 1000 / li.Dst.Packets
	}

	if li.Dur.Milliseconds() > 0 {
		li.Src.Pps = li.Src.Packets * 1000 / uint64(li.Dur.Milliseconds())
		li.Dst.Pps = li.Dst.Packets * 1000 / uint64(li.Dur.Milliseconds())
		li.Src.Bps = li.Src.PayBytes * 8000 / uint64(li.Dur.Milliseconds())
		li.Src.PrettyBps = units.ToMetricString(float64(li.Src.Bps), 3, "", "bps")
		li.Dst.Bps = li.Dst.PayBytes * 8000 / uint64(li.Dur.Milliseconds())
		li.Dst.PrettyBps = units.ToMetricString(float64(li.Dst.Bps), 3, "", "bps")
	}

	return &li, err
}

func (t *TraceTable) getTraceInfo(ctx *Context) (*trace, error) {
	var trc *trace
	var err error

	if t.tbl == nil {
		err = syscall.ENOENT
	} else {
		var ok bool
		if trc, ok = t.tbl[ctx.srcSum+ctx.dstSum]; !ok {
			err = syscall.ENOENT
		}
	}

	return trc, err
}

func (t *TraceTable) Update(ctx *Context) {
	//t.updateTable(&ctx.l4)
	t.updateTable(ctx)
}

// layer should know if the packet captured is acceptable to create/delete or update trace
//func (t *TraceTable) updateTable(ctx *layerContext) {
func (t *TraceTable) updateTable(ctx *Context) {
	t.Lock()

	if t.tbl == nil {
		t.tbl = make(map[uint64]*trace)
	}

	val, ok := t.tbl[ctx.srcSum+ctx.dstSum]
	if !ok {
		val = &trace{}
		t.tbl[ctx.srcSum+ctx.dstSum] = val
	}

	// todo: protocols need to be parsed for the correct layer (e.g., TCP protocol belongs in l4 layer context despite coming from layer 3)
	switch ctx.l3.proto {
	case uint16(layers.IPProtocolTCP):
		t.updateLayer(&val, ctx, layers.LayerTypeTCP)
	case uint16(layers.IPProtocolUDP):
		t.updateLayer(&val, ctx, layers.LayerTypeUDP)
	}

	defer t.Unlock()
}

func (t *TraceTable) addToWindow(f *FlowInfo, /*window *map[uint32]uint32,*/ seq, size uint32) uint32  {
	var ret uint32

	//fmt.Println("add: lwe=", f.winHead, ", uwe=", f.winTail, ", seq=", seq)

	// Sanity check: also check that size is <= mss?
	// If sequence number is in the window
	if t.between(seq, f.winHead, f.winTail) {
		winSize := seq - f.winHead
		if winSize > f.WinSize {
			f.WinSize = winSize
			//fmt.Println("+win: ", f.WinSize)
		}
	}
	return ret


	// Zero window updates?
	//if size > 0 {
	//	if val, ok := f.window[seq]; ok {
	//		if val != size {
	//			// Update
	//			fmt.Println("Seq len changed from", val, "to", size)
	//		}

	//		if size > 0 {
	//			//fmt.Println("seq ", seq, " size ", size, " is a DUP!!!")
	//			f.DupPackets++
	//			f.DupBytes += uint64(size)
	//		}
	//	} else {
	//		//fmt.Println("seq ", seq, " size ", size, " is ADDED!!!")
	//		f.window[seq] = size
	//		ret = size
	//	}
	//}

	//return ret
}

// taken from net/tcp.h
func (t *TraceTable) after(s1, s2 uint32) bool {
	return t.before(s2, s1)
}

func (t *TraceTable) before(s1, s2 uint32) bool {
	return int32(s1 - s2) < 0
}

// s2<=s1<=s3
func (t *TraceTable) between(s1, s2, s3 uint32) bool {
	return s3-s2 >= s1-s2
}

func (t *TraceTable) delFrWindow(f *FlowInfo,/*window *map[uint32]uint32,*/ ack uint32) uint32 {
	var ret uint32

	//fmt.Println("del: lwe=", f.winHead, ", uwe=", f.winTail, ", ack=", ack)
	// Sanity check: also check that size is <= mss?
	// If sequence number is in the window
	if t.between(ack, f.winHead, f.winTail) {
		winSize := ack - f.winHead
		if winSize < f.WinSize {
			f.WinSize = winSize
			f.winHead = ack
			f.winTail = ack + f.mws
			//fmt.Println("-win: ", f.WinSize)
		}
	}
	return ret

	//for seq, size := range f.window {
	//	if seq <= ack {
	//		ret += size
	//		delete(f.window, seq)
	//	}
	//}

	//return ret
}

// This layer could be defined as an interface that can be swapped by anyone's own definition
func (t *TraceTable) updateLayer(tr **trace, ctx *Context, lyr gopacket.LayerType) {
	if (*tr).layer == nil {
		(*tr).layer = make(map[gopacket.LayerType]*LayerInfo)
	}

	val, ok := (*tr).layer[lyr]
	if !ok {
		val = &LayerInfo{}
		val.Src.window = make(map[uint32]uint32)
		val.Dst.window = make(map[uint32]uint32)
		//*val = append(*val, LayerInfo{})
		(*tr).layer[lyr] = val
		val.Src.FlowHash = ctx.srcSum
	}

	if ctx.srcSum == ctx.dstSum {
		// This is a problem. Another piece of information must be used
		// to distinguish
	}

	if ctx.srcSum == val.Src.FlowHash {
		t.updateFlow(&ctx.l4, &val.Src)
		// On first packet, create address strings based on all layers
		switch val.Src.Packets {
		case 1:
			if lyr == layers.LayerTypeTCP { // better to have an interface processing function than if/else
				val.Src.mws = ctx.l4.mws
				val.Src.winHead = ctx.l4.seq
				val.Src.winTail = ctx.l4.seq + ctx.l4.mws
			}
			val.Src.first = ctx.capInfo.Timestamp
			val.Src.Addr = t.createSrcAddr(&ctx.l4)
			val.Dst.Addr = t.createDstAddr(&ctx.l4)
		case 2: // Verify that matches first packet
			val.Src.Rtt = ctx.capInfo.Timestamp.Sub(val.Dst.first) // What if no packet yet for Dst? Check if capture timestamp is zero
		}

		if lyr == layers.LayerTypeTCP { // better to have an interface processing function than if/else
			val.Src.WinSize += t.addToWindow(&val.Src, ctx.l4.seq, uint32(ctx.l4.payLen))
			val.Dst.WinSize -= t.delFrWindow(&val.Dst, ctx.l4.ack)
		}
	} else {
		t.updateFlow(&ctx.l4, &val.Dst)
		// On first packet, create address strings based on all layers
		switch val.Dst.Packets {
		case 1:
			if lyr == layers.LayerTypeTCP { // better to have an interface processing function than if/else
				val.Dst.mws = ctx.l4.mws
				val.Dst.winHead = ctx.l4.seq
				val.Dst.winTail = ctx.l4.seq + ctx.l4.mws
			}
			val.Dst.first = ctx.capInfo.Timestamp
			//val.Dst.Addr = t.createSrcAddr(&ctx.l4)
		//case 1:
			val.Dst.Rtt = ctx.capInfo.Timestamp.Sub(val.Src.first) // The dst RTT will typically be larger than the src RTT if capturing on the src endpoint
		}

		if lyr == layers.LayerTypeTCP { // better to have an interface processing function than if/else
			val.Dst.WinSize += t.addToWindow(&val.Dst, ctx.l4.seq, uint32(ctx.l4.payLen))
			val.Src.WinSize -= t.delFrWindow(&val.Src, ctx.l4.ack)
		}
	}

	val.Dur = ctx.capInfo.Timestamp.Sub(val.Src.first)
	// Get current flow (i.e., last element)
	//if idx := len(*val) - 1; idx >= 0 {
	//	t.updateFlow(&((*val)[idx].Src))
	//}
}

func (t *TraceTable) createSrcAddr(ctx *layerContext) string {
	if ctx.lowerLayer == nil {
		return "[" + ctx.src + "]"
	} else {
		return t.createSrcAddr(ctx.lowerLayer) + ":[" + ctx.src + "]"
	}
}

func (t *TraceTable) createDstAddr(ctx *layerContext) string {
	if ctx.lowerLayer == nil {
		return "[" + ctx.dst + "]"
	} else {
		return t.createDstAddr(ctx.lowerLayer) + ":[" + ctx.dst + "]"
	}
}

func (t *TraceTable) updateFlow(ctx *layerContext, f *FlowInfo) {
	f.HdrBytes += uint64(ctx.hdrLen)
	f.PayBytes += uint64(ctx.payLen)
	f.Packets++
}
