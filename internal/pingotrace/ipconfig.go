package pingotrace

import (
	"fmt"
	"net"
	"strings"
)

func IPConfig() string {
	var infoBuilder strings.Builder

	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	for _, i := range interfaces {
		addresses, err := i.Addrs()
		if err != nil {
			return fmt.Sprintf("Error: %s", err)
		}

		for _, addr := range addresses {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					infoBuilder.WriteString(fmt.Sprintf("Interface: %v\nIP Address: %v\n", i.Name, ipnet.IP.String()))
				}
			}
		}
	}

	return infoBuilder.String()
}
