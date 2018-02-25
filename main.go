package main

import (
	"gopkg.in/mgo.v2"
	"github.com/garyburd/redigo/redis"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"strconv"
)

var jobs chan UserInfo
var done chan bool
var counter int64
var activeCounter int64
var inactiveCounter int64
//var rClient *redis.Client
var pool *redis.Pool

var pusgDataToRedisSaveActivity = `{"token":"${0}","msisdn":"${1}","uid":"${2}","platform_uid":"${3}",
"platform_token":"${4}"}`

func main() {

	jobs = make(chan UserInfo, 10000)
	done = make(chan bool, 1)

	session, err := mgo.Dial("10.15.0.145")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	c := session.DB("userlist").C("newuserdata")

	//rClient =GetRedisInstanceGCP()

	pool = &redis.Pool{
		MaxIdle: 500,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "redis-13511.c2.asia-southeast-1-1.gce.cloud.redislabs.com:13511")
		},
	}

	var result UserInfo
	for w := 1; w <= 500; w++ {
		go workerPool()
	}

	find := c.Find(bson.M{})
	items := find.Iter()
	for items.Next(&result) {
		jobs <- result

	}
	<-done
}

func workerPool() {
	for (true) {
		select {
		case msg,ok := <-jobs:
			if ok {
				conn := pool.Get()
				testdata :=(SubstParams([]string{msg.UserData.Token, msg.UserData.Msisdn, msg.UserData.UID,msg.UserData.PlatformUID,msg.UserData.PlatformToken},pusgDataToRedisSaveActivity))
				_,err :=conn.Do("SADD","newuserdata",testdata)
				if(err!=nil){
					fmt.Println("Active Record not updated ========>",err)
					}
				conn.Close()
				counter++
				fmt.Println("Migrated records till now --- >", counter)
			}
		case <-done:
			done<-true
		}
	}

}

func SubstParams(sessionMap []string, textData string) string {
	for i, value := range sessionMap {
		if strings.ContainsAny(textData, "${"+strconv.Itoa(i)) {
			textData = strings.Replace(textData, "${"+strconv.Itoa(i)+"}", value, -1)
		}
	}
	return textData
}



//func GetRedisInstanceGCP() *redis.Client {
//	var onceGCP sync.Once
//	var instanceGCP *redis.Client
//	onceGCP.Do(func() {
//		client := redis.NewClient(&redis.Options{
//			Addr:     "redis-13065.c2.asia-southeast-1-1.gce.cloud.redislabs.com:13065",
//			Password: "",
//			DB:       0,
//			PoolSize: 100,
//			/*
//				PoolTimeout:  10 * time.Minute,
//				IdleTimeout:  5 * time.Minute,
//				ReadTimeout:  2 * time.Second,
//				WriteTimeout: 10 * time.Second,
//			*/
//		})
//		instanceGCP = client
//	})
//	return instanceGCP
//}

type UserInfo struct {
	UserData UserData `json:"UserData"`
	Flag bool `json:"flag"`
	Active bool `json:"active"`
}

type UserData struct {
	Msisdn string `json:"msisdn"`
	Token  string `json:"token"`
	UID    string `json:"uid"`
	PlatformUID string `json:"platformuid"`
	PlatformToken string `json:"platformtoken"`
}



type UserRecord struct {
	_id           string        `json:"_id"`
	Addressbook   struct{}      `json:"addressbook"`
	BackupToken   string        `json:"backup_token"`
	Connect       int           `json:"connect"`
	Country       string        `json:"country"`
	Devices       []interface{} `json:"devices"`
	Gender        string        `json:"gender"`
	Icon          string        `json:"icon"`
	InvitedJoined []interface{} `json:"invited_joined"`
	Invitetoken   string        `json:"invitetoken"`
	Locale        string        `json:"locale"`
	Msisdn        []string      `json:"msisdn"`
	Name          string        `json:"name"`
	PaUID         string        `json:"pa_uid"`
	Referredby    []string      `json:"referredby"`
	RewardToken   string        `json:"reward_token"`
	Status        int           `json:"status"`
	Sus           int           `json:"sus"`
	Uls           int           `json:"uls"`
	Version       int           `json:"version"`
}
