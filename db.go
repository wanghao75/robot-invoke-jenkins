package main

import (
	"fmt"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	"strings"
)

func initDB(address, user, passwd, name string) (*xorm.Engine, error) {
	host, port := strings.Split(address, ":")[0], strings.Split(address, ":")[1]
	postgres := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, passwd, name)
	engine, err := xorm.NewEngine("postgres", postgres)
	if err != nil {
		return nil, err
	}

	err = engine.Ping()
	if err != nil {
		return nil, err
	}

	if exist, err := engine.IsTableExist("webhooks"); err != nil {
		return nil, err
	} else {
		if exist {
			return engine, nil
		} else {

			err = engine.Sync2(&Webhooks{})

			if err != nil {
				return nil, err
			}
		}
	}

	return engine, nil
}
