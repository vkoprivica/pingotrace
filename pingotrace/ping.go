package pingotrace

import (
	"net"
	"sync/atomic"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

func Ping(ipAddr string) (time.Duration, *net.IPAddr) {
	const (
		protocolICMPv4 = 1
		protocolICMPv6 = 58
	)

	var ipAddress *net.IPAddr
	var protocolICMP int
	var pktConn *icmp.PacketConn
	var err error
	pingSeqNum := newPingICMPSeq()

	ipAddr4, err := net.ResolveIPAddr("ip4", ipAddr)
	if err == nil {
		pktConn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			return 0, ipAddr4
		}
		ipAddress = ipAddr4
		protocolICMP = protocolICMPv4
	} else {
		ipAddr6, err := net.ResolveIPAddr("ip6", ipAddr)
		if err == nil {
			pktConn, err = icmp.ListenPacket("ip6:icmp", "::")
			if err != nil {
				return 0, ipAddr6
			}
			ipAddress = ipAddr6
			protocolICMP = protocolICMPv6
		} else {
			return 0, nil
		}
	}
	defer pktConn.Close()

	var message icmp.Message
	if protocolICMP == protocolICMPv4 {
		message = icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID:   newPingICMPID(),
				Seq:  pingSeqNum,
				Data: []byte("PinGoTrace"),
			},
		}
	} else {
		message = icmp.Message{
			Type: ipv6.ICMPTypeEchoRequest, Code: 0,
			Body: &icmp.Echo{
				ID:   newPingICMPID(),
				Seq:  pingSeqNum,
				Data: []byte("PinGoTrace"),
			},
		}
	}

	b, err := message.Marshal(nil)
	if err != nil {
		return 0, ipAddress
	}

	startTime := time.Now()
	if _, err := pktConn.WriteTo(b, ipAddress); err != nil {
		return 0, ipAddress
	}

	reply := make([]byte, 1500)
	if err := pktConn.SetReadDeadline(time.Now().Add(4 * time.Second)); err != nil {
		return 0, ipAddress
	}

	for {
		n, peer, err := pktConn.ReadFrom(reply)
		if err != nil {
			return 0, ipAddress
		}

		if peer.String() != ipAddress.String() {
			continue
		}

		rm, err := icmp.ParseMessage(protocolICMP, reply[:n])
		if err != nil {
			return 0, ipAddress
		}

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
			echoReply, ok := rm.Body.(*icmp.Echo)
			if ok && echoReply.Seq == pingSeqNum {
				// Adding nanosecond to duration as some Ping tests were returning 0ms RTT
				duration := (time.Since(startTime) + time.Nanosecond)
				return duration, ipAddress
			}
			continue

		case ipv4.ICMPTypeDestinationUnreachable:
			if _, ok := rm.Body.(*icmp.DstUnreach); !ok {
				return 0, ipAddress
			}

			switch rm.Code {
			case 0:
				return 0, ipAddress //, fmt.Errorf("net unreachable")
			case 1:
				return 0, ipAddress //, fmt.Errorf("host unreachable")
			case 2:
				return 0, ipAddress //, fmt.Errorf("protocol unreachable")
			case 3:
				return 0, ipAddress //, fmt.Errorf("port unreachable")
			case 4:
				return 0, ipAddress //, fmt.Errorf("fragmentation needed and DF set")
			case 5:
				return 0, ipAddress //, fmt.Errorf("source route failed")
			default:
				return 0, ipAddress //, fmt.Errorf("destination unreachable with code %d", rm.Code)
			}
		default:
			return 0, ipAddress //, fmt.Errorf("Received unexpected ICMP message type")
		}
	}
}

// ICMP sequence generator to support concurency
var (
	localPingRand  int32
	seqPingCounter int32 // Note that sync/atomic requires int32 or int64
)

// ICMP sequence generator to support concurrency
func newPingICMPSeq() int {
	return int(atomic.AddInt32(&seqPingCounter, 1))
}

// ICMP ID generator
func newPingICMPID() int {
	return int(atomic.AddInt32(&seqPingCounter, 1))
}
