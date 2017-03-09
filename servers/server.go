package servers

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	De "github.com/visionmedia/go-debug"
)

var debug = De.Debug("methodical-monkey:server")

// Server represents our instance of a server
type Server struct {
	instance *ec2.Instance
	svc      *ec2.EC2
}

// NewServer constructs a new instance of a server
func NewServer(instance *ec2.Instance, svc *ec2.EC2) *Server {
	return &Server{instance: instance, svc: svc}
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

// Reboot will reboot the machine
func (server *Server) Reboot() error {
	var err error
	svc := server.svc
	instanceID := server.instance.InstanceId
	debug("reboot machine %v", instanceID)
	input := &ec2.RebootInstancesInput{
		DryRun:      aws.Bool(true),
		InstanceIds: []*string{instanceID},
	}
	_, err = svc.RebootInstances(input)
	awsErr, ok := err.(awserr.Error)
	if ok && awsErr.Code() == "DryRunOperation" {
		debug("we've got the power to reboot this machine")
		input.DryRun = aws.Bool(false)
		_, err = svc.RebootInstances(input)
		if err != nil {
			return err
		}
		debug("reboot machine success")
	}
	return err
}

// WaitForReboot will wait for the machine to be in a running state
func (server *Server) WaitForReboot() error {
	svc := server.svc
	instanceID := server.instance.InstanceId
	debug("waiting for the instance to be running %v", instanceID)
	return svc.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{instanceID},
	})
}
