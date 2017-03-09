package monkey

import (
	"log"

	"github.com/octoblu/methodical-monkey/servers"
)

// ProcessServers finds servers to shutdown
func ProcessServers(list []*servers.Server) error {
	for _, server := range list {
		log.Println(server.String())
	}
	return nil
}
