module github.com/wanghao75/robot-invoke-jenkins

go 1.20

require (
	github.com/bndr/gojenkins v1.1.0
	github.com/go-co-op/gocron v1.31.0
	github.com/go-xorm/xorm v0.7.9
	github.com/lib/pq v1.0.0
	github.com/opensourceways/community-robot-lib v0.0.0-20230111083119-2d2c0df320bb
	github.com/opensourceways/go-gitee v0.0.0-20220714075315-cb246f1dfb96
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/antihax/optional v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	golang.org/x/sys v0.10.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apimachinery v0.24.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
	xorm.io/builder v0.3.6 // indirect
	xorm.io/core v0.7.2-0.20190928055935-90aeac8d08eb // indirect
)

replace github.com/bndr/gojenkins v1.1.0 => github.com/wanghao75/gojenkins v0.0.0-20230727033818-3f8509137068
