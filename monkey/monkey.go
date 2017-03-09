package monkey

import (
	"log"

	"github.com/octoblu/methodical-monkey/servers"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MonkeyClient struct {
	db *mgo.Database
}

// NewMonkeyClient constructs a new instance Methodical Monkey
func NewMonkeyClient(db *mgo.Database) *MonkeyClient {
	return &MonkeyClient{db: db}
}

// Process finds servers to shutdown
func (client *MonkeyClient) Process(list []*servers.Server) error {
	for _, server := range list {
		log.Println(server.GetName())
		err := client.storeMachine(server.GetName())
		if err != nil {
			return err
		}
	}
	log.Println("it has been done")
	return nil
}

func (client *MonkeyClient) storeMachine(machineName string) error {
	c := client.db.C("machines")
	query := bson.M{"name": machineName}
	update := bson.M{
		"$set": bson.M{
			"name":      machineName,
			"updatedAt": bson.Now(),
		},
	}
	_, err := c.Upsert(query, update)
	return err
}
