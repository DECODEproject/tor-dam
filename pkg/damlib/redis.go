package damlib

/*
 * Copyright (c) 2017-2018 Dyne.org Foundation
 * tor-dam is written and maintained by Ivan Jelincic <parazyd@dyne.org>
 *
 * This file is part of tor-dam
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/go-redis/redis"
)

// RedisAddress points us to our Redis instance.
const RedisAddress = "127.0.0.1:6379"

// RedisCli is our global Redis client
var RedisCli = redis.NewClient(&redis.Options{
	Addr:     RedisAddress,
	Password: "",
	DB:       0,
})

// StartRedis is the function that will start up the Redis server. Takes the
// path to a configuration file as an argument and returns error upon failure.
func StartRedis(conf string) (*exec.Cmd, error) {
	log.Println("Starting up redis-server...")
	cmd := exec.Command("redis-server", conf)
	err := cmd.Start()
	if err != nil {
		return cmd, err
	}

	time.Sleep(1000 * time.Millisecond)
	if _, err := RedisCli.Ping().Result(); err != nil {
		return cmd, err
	}

	PubSub := RedisCli.Subscribe(PubSubChan)
	if _, err := PubSub.Receive(); err != nil {
		return cmd, err
	}

	log.Printf("Created \"%s\" channel in Redis.\n", PubSubChan)
	return cmd, nil
}

// PublishToRedis is a function that publishes a node's status to Redis.
// This is used for Gource visualization.
func PublishToRedis(mt, address string) {
	var timestamp, username, modtype, onion, pubstr string

	nodedata, err := RedisCli.HGetAll(address).Result()
	CheckError(err)

	timestamp = nodedata["lastseen"]
	if timestamp == nodedata["firstseen"] {
		modtype = "A"
	} else if mt == "d" {
		modtype = "D"
	} else {
		modtype = "M"
	}
	username = address
	onion = address

	pubstr = fmt.Sprintf("%s|%s|%s|%s", timestamp, username, modtype, onion)

	RedisCli.Publish(PubSubChan, pubstr)
}
