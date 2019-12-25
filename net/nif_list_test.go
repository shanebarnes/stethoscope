package net

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNifList_FindAddr_Uninitialized(t *testing.T) {
	var nl NifList
	lif := GetLoopbackIf()

	_, err := nl.FindAddr(GetNetworkAddr(&lif))
	assert.NotNil(t, err)
}

func TestNifList_FindAddr_Loopback(t *testing.T) {
	var nl NifList
	nl.Refresh()
	lif := GetLoopbackIf()

	nif, err := nl.FindAddr(GetNetworkAddr(&lif))
	assert.Nil(t, err)
	assert.Equal(t, lif.Name, nif.Name)
}

func TestNifList_FindIf_Uninitialized(t *testing.T) {
	var nl NifList
	lif := GetLoopbackIf()

	_, err := nl.FindIf(lif.Name)
	assert.NotNil(t, err)
}

func TestNifList_FindIf_Loopback(t *testing.T) {
	var nl NifList
	nl.Refresh()
	lif := GetLoopbackIf()

	nif, err := nl.FindIf(lif.Name)
	assert.Nil(t, err)
	assert.Equal(t, lif.Name, nif.Name)
}

func TestNifList_Len_Uninitialized(t *testing.T) {
	var nl NifList
	assert.Equal(t, 0, nl.Len())
}

func TestNifList_Len_Initialized(t *testing.T) {
	var nl NifList
	nl.Refresh()
	assert.Greater(t, nl.Len(), 0)
}

func TestNifList_Len_Refresh(t *testing.T) {
	var nl NifList

	assert.Equal(t, 0, nl.Len())

	nl.Refresh()
	assert.Greater(t, nl.Len(), 0)

	if ifs, err := net.Interfaces(); err == nil {
		for _, tif := range ifs {
			nif, err := nl.FindIf(tif.Name)
			assert.Nil(t, err)
			assert.Equal(t, tif.Name, nif.Name)
		}
	}
}

func GetLoopbackIf() net.Interface {
	var nif net.Interface

	if ifs, err := net.Interfaces(); err == nil {
		for _, nif = range ifs {
			if nif.Flags & net.FlagLoopback == 1 {
				return nif
			}
		}
	}

	return nif
}

func GetNetworkAddr(nif *net.Interface) string {
	if addrs, err := nif.Addrs(); err == nil {
		for _, addr := range addrs {
			ip, _, _ := net.ParseCIDR(addr.String())
			return ip.String()
		}
	}

	return ""
}
