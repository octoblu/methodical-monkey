package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
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
	dbsess := connectMongo()
	debug("connected")
	defer dbsess.Close()

	delay := (time.Minute * 5)
	monkeyClient := monkey.NewClient(dbsess.DB(""), delay)
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
		debug("sleeping for 30m")
		time.Sleep(time.Minute * 30)
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

func connectMongo() *mgo.Session {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Panicln("Missing required env MONGODB_URI")
	}
	if strings.Contains(uri, "ssl=true") {
		return connectSecureMongo(uri)
	}
	return connectInsecureMongo(uri)
}

func connectSecureMongo(uri string) *mgo.Session {
	uri = strings.TrimSuffix(uri, "ssl=true")
	uri = strings.TrimSuffix(uri, "?")
	debug("connecting to secure mongo %v", uri)
	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = true

	dialInfo, err := mgo.ParseURL(uri)
	if err != nil {
		panic(err)
	}
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, dialerr := tls.Dial("tcp", addr.String(), tlsConfig)
		return conn, dialerr
	}
	sess, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		panic(err)
	}
	return sess
}

func connectInsecureMongo(uri string) *mgo.Session {
	debug("connecting to insecure mongo %v", uri)
	sess, err := mgo.Dial(uri)
	if err != nil {
		panic(err)
	}
	return sess
}

func version() string {
	version, err := semver.NewVersion(VERSION)
	if err != nil {
		errorMessage := fmt.Sprintf("Error with version number: %v", VERSION)
		log.Panicln(errorMessage, err.Error())
	}
	return version.String()
}
