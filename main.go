package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
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
	db := connectMongo()
	delay := time.Second * 30
	monkeyClient := monkey.NewClient(db, delay)
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

		list, err := servers.List(svc)
		if err != nil {
			panic(err)
		}
		err = monkeyClient.Process(list)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}
}

func connectEC2() *ec2.EC2 {
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = "us-west-2"
	}
	svc := ec2.New(sess, &aws.Config{
		Credentials: credentials.NewEnvCredentials(),
		Region:      aws.String(region),
	})
	return svc
}

func connectMongo() *mgo.Database {
	mongoDbURI := os.Getenv("MONGODB_URI")
	debug("got MONGODB_URI %v", mongoDbURI)
	if mongoDbURI == "" {
		log.Panicln("Missing required env MONGODB_URI")
	}
	session, err := mgo.Dial(mongoDbURI)
	if err != nil {
		panic(err)
	}
	return session.DB("")
}

func version() string {
	version, err := semver.NewVersion(VERSION)
	if err != nil {
		errorMessage := fmt.Sprintf("Error with version number: %v", VERSION)
		log.Panicln(errorMessage, err.Error())
	}
	return version.String()
}
