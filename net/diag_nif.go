package net

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	//"github.com/google/gopacket/tcpassembly"
	diag "github.com/shanebarnes/stethoscope/net/diagnostic"
)


type NifDiag struct {}

func (n *NifDiag) Capture(ifName string, filters ...string) error {
	var handle *pcap.Handle
	_, err := os.Stat(ifName)
	if !os.IsNotExist(err) {
		handle, err = pcap.OpenOffline(ifName)
		if err != nil {
			return err
		}
	} else {
		//handle, err := pcap.OpenLive(ifName, 128, true, pcap.BlockForever)
		ihandle, err0 := pcap.NewInactiveHandle(ifName)
		if err0 != nil {
			return err0
		}
		ihandle.SetBufferSize(32*1024*1024)
		ihandle.SetPromisc(false)
		// Ethernet: frame header: 18B, IPv4: max header: 16x4=60B, IPv6: max header: 40B, TCP: max header: 16x4=60B
		// Ergo: 18+60+60=138 ... round up to 256 
		ihandle.SetSnapLen(164)
		ihandle.SetTimeout(pcap.BlockForever)

		handle, err = ihandle.Activate()
		if err != nil {
			return err
		}
	}
	defer handle.Close()

	filter := ""
	for i, val := range filters {
		if i > 0 {
			filter += " or "
		}
		filter += val
	}

	if err = handle.SetBPFFilter(filter); err != nil {
		return err
	}

	log.Printf("capture: interface %v with filter '%v'\n", ifName, filter)

	// TODO: Zero Copy
	/*pktSource := gopacket.NewPacketSource(handle, handle.LinkType())

	workers := 1
	ch := make(chan gopacket.Packet, workers)
	for i := 0; i < workers; i++ {
		go n.handler(i+1, ch)
	}

	for packet := range pktSource.Packets() {
		ch <- packet
		//n.process(packet)
	}*/

	//var stats *pcap.Stats
	// See parser.go for why this is faster than packet decoding
	var loop layers.Loopback
	var eth layers.Ethernet
	var ip4 layers.IPv4
	var ip6 layers.IPv6
	var tcp layers.TCP
	var udp layers.UDP
	var pay gopacket.Payload

	var parser *gopacket.DecodingLayerParser
	switch handle.LinkType() {
	case layers.LinkTypeLoop, layers.LinkTypeNull:
		parser = gopacket.NewDecodingLayerParser(layers.LayerTypeLoopback, &loop, &eth, &ip4, &ip6, &tcp, &udp, &pay)
	case layers.LinkTypeEthernet:
		parser = gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &udp, &pay)
	default:
		fmt.Println("Detected unsupported: ", handle.LinkType())
		err = syscall.ENOTSUP
	}
	decodedLayers := make([]gopacket.LayerType, 0, 10)

	last := time.Now()
	table := diag.TraceTable{}


	tcpDiag := diag.TcpDiagnostic{}

	for err == nil {
		if data, ci, err1 := handle.ZeroCopyReadPacketData(); err1 == nil {
			l4proto := layers.LayerTypeTCP
			err = parser.DecodeLayers(data, &decodedLayers)
			ctx := diag.NewContext(&ci)
			for _, typ := range decodedLayers {
				switch typ {
				case layers.LayerTypeLoopback:
					err = diag.ProcessLoopback(ctx, &loop)
				case layers.LayerTypeEthernet:
					err = diag.ProcessEthernet(ctx, &eth)
				case layers.LayerTypeIPv4:
					err = diag.ProcessIPv4(ctx, &ip4)
				case layers.LayerTypeIPv6:
					err = diag.ProcessIPv6(ctx, &ip6)
				case layers.LayerTypeTCP:
					//err = diag.ProcessTCP(ctx, &tcp)
					err = tcpDiag.ProcessLayer(ctx, &tcp)
				case layers.LayerTypeUDP:
					l4proto = layers.LayerTypeUDP
					err = diag.ProcessUDP(ctx, &udp)
				default:
					err = syscall.EPROTONOSUPPORT
				}

				if err != nil {
					//fmt.Println(typ)
					//fmt.Println(err)
					err = nil
					break
				}
			}

			ctx.Finalize()
			table.Update(ctx)
			now := time.Now()
			diff := now.Sub(last)

			isOpen, _ := ctx.IsOpenPkt(l4proto)
			isClose, _ := ctx.IsClosePkt(l4proto)

			if diff >= time.Second || isOpen || isClose {
				last = now
				if li, err := table.GetLayerInfo(ctx, l4proto/*layers.LayerTypeTCP*/); err == nil {
					if b, err := json.Marshal(li); err == nil {
						fmt.Println(string(b))
						fmt.Println("")
					}

					/*fmt.Println(" Src Address: ", li.Src.Addr, "Dst Address: ", li.Dst.Addr, "Elapsed: ", li.Dur)
					fmt.Println(" Src Rtt: ", li.Src.Rtt, "Dst Rtt: ", li.Dst.Rtt)
					fmt.Println(" Src Wind: ", li.Src.WinSize, "Dst Wind: ", li.Dst.WinSize)
					if li.Src.Packets > 0 {
						fmt.Print(" Src PL: ", fmt.Sprintf("%.6f", float64(li.Src.DupPackets) * 100 / float64(li.Src.Packets)))
					}
					if li.Dst.Packets > 0 {
						fmt.Print(" Dst PL: ", fmt.Sprintf("%.6f", float64(li.Dst.DupPackets) * 100 / float64(li.Dst.Packets)))
					}
					fmt.Println("")
					if li.Dur.Milliseconds() > 0 {
						fmt.Println(" Src PPS: ", li.Src.Packets * 1000 / uint64(li.Dur.Milliseconds()), "Dst PPS: ", li.Dst.Packets * 1000 / uint64(li.Dur.Milliseconds()))
						fmt.Println(" Src bps: ", li.Src.PayBytes * 8000 / uint64(li.Dur.Milliseconds()), "Dst bps: ", li.Dst.PayBytes * 8000 / uint64(li.Dur.Milliseconds()))
					}
					fmt.Println(" Src Packets: ", li.Src.Packets, "(Dup: ", li.Src.DupPackets, " Dst Packets: ", li.Dst.Packets)
					fmt.Println(" Src Hdr Bytes: ", li.Src.HdrBytes, ", Dst Hdr Bytes: ", li.Dst.HdrBytes)
					fmt.Println(" Src Pay Bytes: ", li.Src.PayBytes, ", Dst Pay Bytes: ", li.Dst.PayBytes)*/
					//if stats, err := handle.Stats(); err == nil {
					//	fmt.Println("Packets: received ", stats.PacketsReceived, ", dropped ", stats.PacketsDropped, ", ifDropped ", stats.PacketsIfDropped)
					//}
				}
			}
		} else {
			err = err1
		}
		//if err1 == nil {
		//		n.process(handle, data)
		//}
//		select {
//		case /*pkt :=*/ <-pkts:
//			//log.Println(ifName, ": captured packet ", pkt)
//			stats, err = handle.Stats()
//			log.Println(ifName, ": stats: ", stats)
//		//default:
//		//	log.Println("Default capture")
//		}
	}

	return nil
}

func (n *NifDiag) handler(id int, ch chan gopacket.Packet) error {
	var err error

	counter := int64(0)
	marker := time.Now()
	for /*packet :=*/ _ = range ch {
		counter++
		//n.process(packet)

		now := time.Now()
		diff := now.Sub(marker)
		if diff > time.Second {
			fmt.Printf("Handler %d has processed %v packets in %v\n", id, counter, diff)
			marker = now
			counter = 0
		}
	}
	fmt.Println("Done handling")
	//for err == nil {
		//if data, _, err1 := handle./*ZeroCopy*/ReadPacketData(); err1 == nil {
			//n.process(handle, data)
		//} else {
		//	fmt.Println("capture worker failed: ", err1)
		//	err = err1
		//}
	//}

	return err
}
func (n *NifDiag) process(/*packet gopacket.Packet*/handle *pcap.Handle, data []byte) {
	packet := gopacket.NewPacket(data, handle.LinkType(), gopacket.NoCopy)

	var err error
	//ctx := diag.Context{}
	for _, /*layer*/ _= range packet.Layers() {
		/*switch layer.LayerType() {
		case layers.LayerTypeLoopback:
			err = diag.ProcessLoopback(&ctx, layer.(*layers.Loopback))
		case layers.LayerTypeEthernet:
			err = diag.ProcessEthernet(&ctx, layer.(*layers.Ethernet))
		case layers.LayerTypeIPv4:
			err = diag.ProcessIPv4(&ctx, layer.(*layers.IPv4))
		case layers.LayerTypeIPv6:
			err = diag.ProcessIPv6(&ctx, layer.(*layers.IPv6))
		case layers.LayerTypeTCP:
			err = diag.ProcessTCP(&ctx, layer.(*layers.TCP))
		case layers.LayerTypeUDP:
			err = diag.ProcessUDP(&ctx, layer.(*layers.UDP))
		default:
			err = syscall.EPROTONOSUPPORT
		}*/

		if err != nil {
			break
		}

	}
	//fmt.Println("Diagnostic context: ", ctx)
	// Set up assembly
	//streamFactory := &(StatsStreamFactory{})
	//streamPool := tcpassembly.NewStreamPool(streamFactory)
	//assembler := tcpassembly.NewAssembler(streamPool)
	//assembler.MaxBufferedPagesPerConnection = 0 //*bufferedPerConnection
	//assembler.MaxBufferedPagesTotal = 0 //*bufferedTotal

	// Walk layers to build hash
	//if packet.NetworkLayer() == nil {
	//	log.Println("No network layer packet")
	//} else if packet.TransportLayer() == nil {
	//	log.Println("No transport layer packet")
	//} else if packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
	//	log.Println("Unusable transport layer packet")
	//} else {
	//	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	//	if tcpLayer != nil {
	//		tcp := tcpLayer.(*layers.TCP)
	//		//var ip4 layers.IPv4
	//		//var netFlow gopacket.Flow = ip4.NetworkFlow()
	//		//assembler.AssembleWithTimestamp(netFlow, tcp, ci.Timestamp)
	//		fmt.Printf("%v\n", tcp)
	//		//if tcp.SYN {
	//		//	fmt.Println("TCP socket: ", tcp.SrcPort, " > ", tcp.DstPort)
	//		//}
	//	}
	//}
}
