package net

import (
	"sync"
	"syscall"

	"github.com/google/gopacket/pcap"
)

type NifList struct {
	mp  map[string]pcap.Interface
	mtx sync.RWMutex
}

func (n *NifList) FindAddr(name string) (pcap.Interface, error) {
	var pif pcap.Interface
	var err error =  syscall.ENXIO

	n.mtx.RLock()
	defer n.mtx.RUnlock()
	for _, val := range n.mp {
		for _, addr := range val.Addresses {
			if addr.IP.String() == name {
				return val, nil
			}
		}
	}

	return pif, err
}

func (n *NifList) FindIf(name string) (pcap.Interface, error) {
	n.mtx.RLock()
	defer n.mtx.RUnlock()
	var err error
	pif, ok := n.mp[name]
	if !ok {
		err = syscall.ENXIO
	}

	return pif, err
}

func (n *NifList) Len() int {
	n.mtx.RLock()
	defer n.mtx.RUnlock()
	return len(n.mp)
}

func (n *NifList) Refresh() (error) {
	ifs, err := pcap.FindAllDevs()
	if err == nil {
		n.mtx.Lock()
		defer n.mtx.Unlock()

		if n.mp == nil {
			n.mp = make(map[string]pcap.Interface)
		}

		for _, pif := range ifs {
			n.mp[pif.Name] = pif
		}
	}

	return err
}
