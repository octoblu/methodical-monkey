package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mgo "gopkg.in/mgo.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/coreos/go-semver/semver"
	"github.com/octoblu/methodical-monkey/monkey"
	"github.com/octoblu/methodical-monkey/servers"
	"github.com/urfave/cli"
	De "github.com/visionmedia/go-debug"
)

var debug = De.Debug("methodical-monkey:main")

func main() {
	app := cli.NewApp()
	app.Name = "methodical-monkey"
	app.Version = version()
	app.Action = run
	app.Flags = []cli.Flag{}
	app.Run(os.Args)
}

func run(context *cli.Context) {
	svc := connectEC2()
	db := connectMongo("mongodb://localhost:27017", "methodical-monkey")
	monkeyClient := monkey.NewClient(db)
	sigTerm := make(chan os.Signal)
	signal.Notify(sigTerm, syscall.SIGTERM)

	sigTermReceived := false

	go func() {
		<-sigTerm
		fmt.Println("SIGTERM received, waiting to exit")
		sigTermReceived = true
	}()

	for {
		if sigTermReceived {
			fmt.Println("I'll be back.")
			os.Exit(0)
		}

		debug("methodical-monkey.loop")
		list, err := servers.List(svc)
		if err != nil {
			panic(err)
		}
		err = monkeyClient.Process(list)
		if err != nil {
			panic(err)
		}
		// time.Sleep(60 * time.Second)
		os.Exit(0)
	}
}

func connectEC2() *ec2.EC2 {
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}
	svc := ec2.New(sess, &aws.Config{Region: aws.String("us-west-2")})
	return svc
}

func connectMongo(url, db string) *mgo.Database {
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	return session.DB(db)
}

func version() string {
	version, err := semver.NewVersion(VERSION)
	if err != nil {
		errorMessage := fmt.Sprintf("Error with version number: %v", VERSION)
		log.Panicln(errorMessage, err.Error())
	}
	return version.String()
}
