package pingotrace

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func PinGoTrace(destIP string, maxHops int, timeout time.Duration, ctx context.Context, traceOutputChan chan []string) {
	// defer close(traceOutputChan)
	// Resolve the destination IP address
	ipAddr, err := net.ResolveIPAddr("ip4", destIP)
	if err != nil {
		traceOutputChan <- []string{fmt.Sprintf("Unable to resolve destination IP address: %s", err)}
		return
	}

	// Create an ICMP packet connection
	pktConn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		traceOutputChan <- []string{fmt.Sprintf("Unable to open ICMP connection: %s", err)}
		return
	}
	defer pktConn.Close()

	// Flag to indicate if destination has been reached
	destinationReached := false

	// Loop through the hop counts up to maxHops
	var hop int
	for hop = 1; hop <= maxHops; hop++ {
		if destinationReached {
			break
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			// defer close(traceOutputChan)
			return
		default:
			// Set the TTL (Time To Live) for the current hop
			pktConn.IPv4PacketConn().SetTTL(hop)

			// Initialize slice to store round-trip times for each probe
			responseTimes := make([]string, 3)
			currentPeer := ""

			// Send 3 probes per hop
			for probe := 0; probe < 3; probe++ {
				// Record the start time for this probe
				startTime := time.Now()

				// Set a read deadline for the connection
				err = pktConn.SetReadDeadline(startTime.Add(timeout))
				if err != nil {
					traceOutputChan <- []string{fmt.Sprintf("Unable to set read deadline: %s", err)}
					return
				}

				traceSeqNum := newTraceICMPSeq()
				// Create an ICMP Echo Request message
				var message icmp.Message
				message = icmp.Message{
					Type: ipv4.ICMPTypeEcho, Code: 0,
					Body: &icmp.Echo{
						ID:   newTraceICMPID(),
						Seq:  traceSeqNum,
						Data: []byte("PinGoTrace"),
					},
				}

				// Marshal the ICMP message into bytes
				b, err := message.Marshal(nil)
				if err != nil {
					traceOutputChan <- []string{fmt.Sprintf("Unable to marshal ICMP message: %s", err)}
					return
				}

				// Send the ICMP message
				_, err = pktConn.WriteTo(b, ipAddr)
				if err != nil {
					traceOutputChan <- []string{fmt.Sprintf("Unable to send ICMP message: %s", err)}
					return
				}

				// Receive the ICMP response
				buf := make([]byte, 1500)
				n, peer, err := pktConn.ReadFrom(buf)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						// Handle timeouts
						responseTimes[probe] = "*"
						continue
					} else {
						// Handle other errors
						traceOutputChan <- []string{fmt.Sprintf("Unable to read ICMP message: %s", err)}
						return
					}
				}

				// Record the peer address
				currentPeer = peer.(*net.IPAddr).String()

				// Parse the received ICMP message
				messagePtr, err := icmp.ParseMessage(1, buf[:n])
				if err != nil {
					traceOutputChan <- []string{fmt.Sprintf("Unable to parse ICMP message: %s", err)}
					return
				}
				message = *messagePtr

				// Calculate and record the round-trip time
				rtt := time.Since(startTime).Round(time.Millisecond)
				switch message.Type {
				case ipv4.ICMPTypeTimeExceeded:
					// Time exceeded, usually means still in transit
					responseTimes[probe] = fmt.Sprintf("RTT: %v", rtt)
				case ipv4.ICMPTypeEchoReply:
					// Echo reply received, destination reached
					responseTimes[probe] = fmt.Sprintf("RTT: %v", rtt)
					destinationReached = true
				default:
					// Unexpected ICMP message received
					responseTimes[probe] = "Unexpected ICMP message"
				}
			}

			// If no reply is received, mark as timed out
			if currentPeer == "" {
				currentPeer = "Request timed out"
			} else {
				// Try to resolve the IP address to a hostname
				dnsName, err := net.LookupAddr(currentPeer)
				if err == nil && len(dnsName) > 0 {
					currentPeer = fmt.Sprintf("%s [%s]", dnsName[0][:len(dnsName[0])-1], currentPeer)
				}
			}

			// Send the output line to the channel for further processing or output
			outputLine := append([]string{fmt.Sprintf("%2d", hop), currentPeer}, responseTimes...)
			traceOutputChan <- outputLine
		}
	}
}
