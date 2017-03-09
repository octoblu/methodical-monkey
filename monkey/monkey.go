package monkey

import (
	"time"

	"github.com/octoblu/methodical-monkey/servers"
	De "github.com/visionmedia/go-debug"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var debug = De.Debug("methodical-monkey:monkey")

// Client represents the Methodical Monkey client
type Client struct {
	db    *mgo.Database
	delay time.Duration
}

// Machine represents the machine database document
type Machine struct {
	Name       string    `bson:"name"`
	RebootedAt time.Time `bson:"rebootedAt"`
}

// NewClient constructs a new instance Methodical Monkey
func NewClient(db *mgo.Database, delay time.Duration) *Client {
	return &Client{db: db, delay: delay}
}

// Process finds servers to shutdown
func (client *Client) Process(list []*servers.Server) error {
	var err error
	for _, server := range list {
		err = client.ProcessMachine(server)
		if err != nil {
			return err
		}
		debug("sleep for %v", client.delay.String())
		time.Sleep(client.delay)
	}
	debug("it has been done")
	return err
}

// ProcessMachine will determine if a single machine needs to be rebooted
func (client *Client) ProcessMachine(server *servers.Server) error {
	var err error
	machineName := server.GetName()
	debug(machineName)
	err = client.storeMachine(server)
	if err != nil {
		return err
	}
	shouldReboot, err := client.shouldRebootMachine(server)
	if err != nil {
		return err
	}
	if !shouldReboot {
		debug("i should not reboot")
		return nil
	}
	debug("i am going to reboot")
	return client.rebootMachine(server)
}

func (client *Client) storeMachine(server *servers.Server) error {
	c := client.db.C("machines")
	query := bson.M{"name": server.GetName()}
	update := bson.M{
		"$set": bson.M{
			"name": server.GetName(),
		},
	}
	_, err := c.Upsert(query, update)
	return err
}

func (client *Client) shouldRebootMachine(server *servers.Server) (bool, error) {
	c := client.db.C("machines")
	query := bson.M{"name": server.GetName()}
	machine := &Machine{}
	err := c.Find(query).One(machine)
	if err != nil {
		return false, err
	}
	if time.Since(machine.RebootedAt) < time.Hour {
		debug("no need to reboot %v", server.GetName())
		return false, nil
	}
	if time.Since(server.GetLaunchTime()) < time.Hour {
		debug("I am newby server, skip me")
		return false, nil
	}
	return true, nil
}

func (client *Client) rebootMachine(server *servers.Server) error {
	var err error
	debug("rebooting machine %v", server.GetName())
	c := client.db.C("machines")
	err = server.Reboot()
	if err != nil {
		return err
	}
	err = server.WaitForReboot()
	if err != nil {
		return err
	}
	debug("updating rebootedAt")
	query := bson.M{"name": server.GetName()}
	update := bson.M{
		"$set": bson.M{
			"rebootedAt": bson.Now(),
		},
	}
	return c.Update(query, update)
}
