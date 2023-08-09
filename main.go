package main

import (
	"context"
	"flag"
	jenkins "github.com/bndr/gojenkins"
	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/opensourceways/community-robot-lib/logrusutil"
	liboptions "github.com/opensourceways/community-robot-lib/options"
	framework "github.com/opensourceways/community-robot-lib/robot-gitee-framework"
	"github.com/opensourceways/community-robot-lib/secret"
	"github.com/sirupsen/logrus"
	"os"
)

var Ctx = context.Background()

type options struct {
	service         liboptions.ServiceOptions
	gitee           liboptions.GiteeOptions
	jenkinsEndpoint string
	jenkinsUser     string
	jenkinsPass     string
	dbAddress       string
	dbUser          string
	dbPasswd        string
	dbName          string
}

func (o *options) Validate() error {
	if err := o.service.Validate(); err != nil {
		return err
	}

	return o.gitee.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.gitee.AddFlags(fs)
	o.service.AddFlags(fs)
	fs.StringVar(&o.jenkinsEndpoint, "jenkins-endpoint", "", "The endpoint of jenkins service")
	fs.StringVar(&o.jenkinsUser, "jenkins-user", "", "The admin username of jenkins")
	fs.StringVar(&o.jenkinsPass, "jenkins-pass", "", "The admin password of jenkins")
	fs.StringVar(&o.dbAddress, "address", "", "The db address contains ip and port")
	fs.StringVar(&o.dbUser, "user", "", "The db user")
	fs.StringVar(&o.dbPasswd, "password", "", "The db user's password")
	fs.StringVar(&o.dbName, "db", "", "The db name")

	_ = fs.Parse(args)

	return o
}

func main() {
	logrusutil.ComponentInit(botName)

	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	secretAgent := new(secret.Agent)
	if err := secretAgent.Start([]string{o.gitee.TokenPath}); err != nil {
		logrus.WithError(err).Fatal("Error starting secret agent.")
	}

	defer secretAgent.Stop()

	c := giteeclient.NewClient(secretAgent.GetTokenGenerator(o.gitee.TokenPath))

	j := jenkins.CreateJenkins(nil, o.jenkinsEndpoint, o.jenkinsUser, o.jenkinsPass)

	jen, err := j.Init(Ctx)
	if err != nil {
		logrus.WithError(err).Fatal("Error starting jenkins agent.")
	}

	dbEngine, err := initDB(o.dbAddress, o.dbUser, o.dbPasswd, o.dbName)
	if err != nil {
		logrus.WithError(err).Fatal("Error connect to db")
	}

	p := newRobot(c, jen, dbEngine)

	framework.Run(p, o.service)
}
