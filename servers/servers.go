package servers

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

// List fetches all the ec2 AWS Servers
func List(svc *ec2.EC2) ([]*Server, error) {
	instances := []*Server{}
	result, err := svc.DescribeInstances(nil)
	if err != nil {
		return instances, err
	}
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, NewServer(instance))
		}
	}
	return instances, nil
}
