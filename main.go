package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		monkey.ProcessServers(list)
		time.Sleep(60 * time.Second)
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

func version() string {
	version, err := semver.NewVersion(VERSION)
	if err != nil {
		errorMessage := fmt.Sprintf("Error with version number: %v", VERSION)
		log.Panicln(errorMessage, err.Error())
	}
	return version.String()
}
