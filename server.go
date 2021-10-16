package swim

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

type Server struct {
	address            *net.UDPAddr
	initialPeerAddress *net.UDPAddr
	alivePeers         *PeerList
	pinged             chan *net.UDPAddr
	acknowledged       chan *net.UDPAddr
	suspected          chan *net.UDPAddr
	timeout            time.Duration
}

func NewServer(port, initialPeerPort string, timeout time.Duration) (*Server, error) {
	address, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		return nil, err
	}

	alive := NewPeerList()
	pinged := make(chan *net.UDPAddr)

	initialPeerAddress, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("0.0.0.0:%s", initialPeerPort))
	if err == nil {
		alive.Add(initialPeerAddress)
	}

	return &Server{
		address:            address,
		alivePeers:         alive,
		initialPeerAddress: initialPeerAddress,
		pinged:             pinged,
	}, nil
}

func (s *Server) Run() {
	conn, err := net.ListenUDP("udp4", s.address)
	if err != nil {
		panic(err)
	}

	log.Printf("Listening at %s", s.address.String())

	defer conn.Close()

	if s.initialPeerAddress != nil {
		s.Alive(s.address.String(), s.initialPeerAddress)
	}

	go func(server *Server) {
		for {
			pinged := <-server.pinged
			select {
			case <-server.acknowledged:
				continue
			case <-time.After(server.timeout):
				server.suspected <- pinged
			}
		}
	}(s)

	go func(server *Server) {
		for {
			suspected := <-s.suspected
			server.Suspect(suspected.String(), server.alivePeers.Sample(suspected))
		}
	}(s)

	for {
		message := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(message)
		if err != nil {
			log.Println(err)
		}

		if n > 0 {
			msg, err := NewMessage(message[0:n])
			if err != nil {
				log.Println(err)
			}
			if err := s.Handle(msg); err != nil {
				log.Println(err)
			}
		}
	}
}

func (s *Server) Ping(address *net.UDPAddr) error {
	log.Printf("I'll ping %s", address.String())

	if err := s.sendMessage(Ping, "", address); err != nil {
		return err
	}

	s.pinged <- address
	return nil
}

func (s *Server) Ack(address *net.UDPAddr) error {
	log.Printf("I'll ack %s", address.String())

	if err := s.sendMessage(Ack, "", address); err != nil {
		return err
	}

	s.acknowledged <- address
	return nil
}

func (s *Server) Alive(who string, address *net.UDPAddr) error {
	log.Printf("I'll say to %s that %s is alive", address.String(), who)
	return s.sendMessage(Alive, who, address)
}

func (s *Server) Suspect(who string, address *net.UDPAddr) error {
	return s.sendMessage(Suspect, who, address)
}

func (s *Server) Handle(message *Message) error {
	switch message.Type {
	case Ping:
		return s.HandlePing(message)
	case Ack:
		return s.HandleAck(message)
	case Alive:
		return s.HandleAlive(message)
	default:
		return InvalidMessage
	}
}

func (s *Server) HandlePing(message *Message) error {
	if address, err := message.FromAddress(); err != nil {
		return InvalidMessage
	} else {
		log.Printf("%s pinged me", message.From)
		return s.Ack(address)
	}
}

func (s *Server) HandleAlive(message *Message) error {
	if who, err := message.WhoAddress(); err != nil {
		return InvalidMessage
	} else {
		log.Printf("%s said %s is alive", message.From, message.Who)
		if !s.alivePeers.Include(who) {
			s.alivePeers.Add(who)
			from, _ := message.FromAddress()
			s.Alive(who.String(), s.alivePeers.Sample(from))
		}
		return nil
	}
}

func (s *Server) HandleAck(message *Message) error {
	log.Println("Got Ack from:", message.From)
	return nil
}

func (s *Server) sendMessage(msgType, who string, target *net.UDPAddr) error {
	clientConn, err := net.DialUDP("udp4", nil, target)
	if err != nil {
		return err
	}

	msg := Message{
		Type: msgType,
		Who:  who,
		From: s.address.String(),
	}

	if jsonMsg, err := json.Marshal(msg); err != nil {
		return err
	} else {
		clientConn.Write(jsonMsg)
		return nil
	}
}
