// Copyright 2018 The Harbor Authors. All rights reserved.

package pool

//Interface for worker pool
type Interface interface {
	//Start to server
	Start() error
}
