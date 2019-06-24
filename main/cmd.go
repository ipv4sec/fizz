/*
 * @author Qi Fenglong (qifenglong@artron.net) @time: 2019-06-24
 * Copyright (c) 2000-2019. http://www.artron.net. All rights reserved.
 * Use of this source code is governed a license that can be found in the LICENSE file.
 */
package main

import (
	"fizz"
	"fmt"
	"github.com/satori/go.uuid"
	"time"
)

func main() {
	serverId := uuid.NewV4().String()
	fmt.Println(serverId)

	client := &fizz.Zookeeper{}

	client.Addr = []string{"192.168.64.185:21810"}
	client.MasterPath = "/fizz"
	client.ServerId = serverId
	client.SessionTimeout = time.Second * 5
	client.ConnectTimeout = time.Second * 5

	err := client.Connect()
	fmt.Println(err)

	client.MonitorMasterSlave(func() {
		fmt.Println("我成为主啦")
	}, func() {
		fmt.Println("我成为仆啦")
	})

	client.ElectionMaster()
	client.MonitorLocalEvent()

	select {}
}
