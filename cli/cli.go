package cli

import (
	"d7024e/kademlia"
	"fmt"
	"os"
)

func Init() {
	switch os.Args[1] {
	case "ping":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ping <ip-address>")
			return
		}
		ip := os.Args[2]
		kademlia.SendPingMessageByIP(ip)
		break

	case "help":
		fmt.Println("Commands:")
		fmt.Println("ping <ip-address>")
		break
	default:
		fmt.Println("Unknown command:", os.Args[1])
	}

}
