/*
 * @author Qi Fenglong (qifenglong@artron.net) @time: 2019-06-24
 * Copyright (c) 2000-2019. http://www.artron.net. All rights reserved.
 * Use of this source code is governed a license that can be found in the LICENSE file.
 */
package fizz

import (
	"errors"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

var (
	master = make(chan bool,1)
	slave = make(chan bool,1)

	Event = make(chan zk.Event,1)
)


type Zookeeper struct {
	Client *zk.Conn

	Addr []string
	MasterPath string
	ServerId string
	SessionTimeout time.Duration
	ConnectTimeout time.Duration
}

func (z *Zookeeper) Connect() error {
	var err error
	sign := make(<-chan zk.Event)

	z.Client, sign, err = zk.Connect(z.Addr, z.SessionTimeout)
	if err != nil {
		return err
	}
	connected := make(chan bool)
	go func() {
		for connEvent := range sign {
			if connEvent.State == zk.StateConnected {
				connected <- true
			}
		}
	}()

	select {
	case <-connected:
		return nil
	case <-time.After(z.ConnectTimeout):
		return errors.New("connect timeout")
	}
}

func (z *Zookeeper) MonitorMasterSlave(masterFunc func(), slaveFunc func()) {
	go func() {
		for {
			select {
			case <-master: masterFunc()
			case <-slave: slaveFunc()
			}
		}
	}()
}

func (z *Zookeeper) MonitorRemoteEvent() {
	go func() {
		_, _, ch, err := z.Client.GetW(z.MasterPath)
		if err != nil {
			panic(err)
		}
		Event <- <-ch
	}()
}

func (z *Zookeeper) ElectionMaster() {
	existed, _, _ := z.Client.Exists(z.MasterPath)
	if !existed {
		_, err := z.Client.Create(z.MasterPath, []byte(z.ServerId), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
		if err != nil {
			slave <- true
		} else {
			master <- true
		}
	}else {
		slave <- true
	}
	z.MonitorRemoteEvent()
}

func (z *Zookeeper) MonitorLocalEvent() {
	go func() {
		for {
			select {
			case <- Event:
				z.ElectionMaster()
			}
		}
	}()
}