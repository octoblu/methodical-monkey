package servers

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// Server represents our instance of a server
type Server struct {
	instance *ec2.Instance
}

// NewServer constructs a new instance of a server
func NewServer(instance *ec2.Instance) *Server {
	return &Server{instance: instance}
}

// String should return the JSON stringified version of the instance
func (server *Server) String() string {
	return fmt.Sprintf("%v", server.instance.GoString())
}

// GetName should return the instance name
func (server *Server) GetName() string {
	name := ""
	for _, tag := range server.instance.Tags {
		key := *tag.Key
		if key == "Name" {
			name = *tag.Value
		}
	}
	return fmt.Sprint(name)
}
