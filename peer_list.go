package swim

import (
	"net"
	"sync"
)

type PeerList struct {
	sync.Mutex
	list map[string]*net.UDPAddr
}

func NewPeerList() *PeerList {
	list := make(map[string]*net.UDPAddr)
	return &PeerList{
		list: list,
	}
}

func (l *PeerList) Add(peer *net.UDPAddr) {
	l.Lock()
	l.list[peer.String()] = peer
	l.Unlock()
}

func (l *PeerList) Remove(peer *net.UDPAddr) {
	l.Lock()
	delete(l.list, peer.String())
	l.Unlock()
}

func (l *PeerList) Size() int {
	return len(l.list)
}

func (l *PeerList) Sample(ignored *net.UDPAddr) *net.UDPAddr {
	var sample *net.UDPAddr
	for _, value := range l.list {
		if value != ignored {
			sample = value
			break
		} else {
			continue
		}
	}
	return sample
}

func (l *PeerList) Include(address *net.UDPAddr) bool {
	_, included := l.list[address.String()]
	return included
}
