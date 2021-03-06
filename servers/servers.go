package servers

import (
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	De "github.com/visionmedia/go-debug"
)

var debug = De.Debug("methodical-monkey:servers")

// List fetches all the ec2 AWS Servers
func List(svc *ec2.EC2) ([]*Server, error) {
	instances := []*Server{}
	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running")},
		},
		&ec2.Filter{
			Name:   aws.String("tag:methodical-monkey:rebootable"),
			Values: []*string{aws.String("true")},
		},
	}
	result, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: filters,
	})
	if err != nil {
		return instances, err
	}
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if instance != nil {
				instances = append(instances, NewServer(instance, svc))
			}
		}
	}
	instances = shuffle(instances)
	debug("found %v instances", len(instances))
	return instances, nil
}

func shuffle(servers []*Server) []*Server {
	rand.Seed(int64(time.Now().Nanosecond()))
	newServers := make([]*Server, len(servers))
	for i, server := range servers {
		v := rand.Intn(i + 1)
		newServers[v] = server
	}
	return newServers
}
