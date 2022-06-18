package ecode

import (
	"fmt"
	"sync"
	"testing"
)

// 这里仅做错误消息map并发读写
func TestCode(t *testing.T) {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer func() {
			wg.Done()
		}()
		testWrite(t)
	}()

	go func() {
		defer func() {
			wg.Done()
		}()
		testRead(t)
	}()


	wg.Wait()
}

func testWrite(t *testing.T) {
	for i := 0; i < 1000; i++ {
		OK.Message()
	}
}

func testRead(t *testing.T) {
	for i := 0; i < 1000; i++ {
		_ = NewWithMessage(i, fmt.Sprintf("%v_msg", i))
	}
}
