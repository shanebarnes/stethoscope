package stethoscope

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

func (n *NifList) New() {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	n.mp = make(map[string]pcap.Interface)
}

func (n *NifList) Refresh() (error) {
	var err error
	if n.mp == nil {
		return syscall.ENOMEM
	}

	var ifs []pcap.Interface
	if ifs, err = pcap.FindAllDevs(); err == nil {
		n.mtx.Lock()
		defer n.mtx.Unlock()
		for _, pif := range ifs {
			n.mp[pif.Name] = pif
		}

	}

	return err
}

func (n *NifList) Len() int {
	n.mtx.RLock()
	defer n.mtx.RUnlock()
	return len(n.mp)
}
