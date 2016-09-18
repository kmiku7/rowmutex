package rowmutex

import (
	"errors"
	"sync"
	"testing"
	"time"
	"fmt"
)

func Test_Do(t *testing.T) {
	expectedErr := errors.New("error info for test.")

	var table Table
	err := table.Do("test-key", func() error {
		t.Log("function called.")
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("Do expect:%v, actual:%v", expectedErr, err)
	}
}

func Test_Concurrency(t *testing.T) {

	counterA := 0
	counterB := 10000
	var table Table
	var wg sync.WaitGroup
	wg.Add(203)

	rowKey := "test-key-a"
	rowKey2 := "test-key-b"
	concurrencyCount := 100

	cycle_dep := make(chan interface{})
	finish_dep := make(chan interface{})

	go func() {
		_ = <- cycle_dep
		fmt.Println("all done")
		finish_dep <- nil
	}()

	go func() {
		c := time.Tick(time.Second * 2)
		for now := range c {
			fmt.Printf("tick %v %s\n", now, time.Now().String())
		}
	}()

	for i := 0; i < concurrencyCount; i++ {
		idx := i
		go table.Do(rowKey, func() error {
			time.Sleep(5 * time.Microsecond)
			counterA++
			t.Logf("a-routine\tidx:%d\tcounter:%d\twaitCount:%d", idx, counterA, table.m[rowKey].waitCount)
			wg.Done()
			return nil
		})
		go table.Do(rowKey2, func() error {
			time.Sleep(5 * time.Microsecond)
			counterB++
			t.Logf("b-routine\tidx:%d\tcounter:%d\twaitCount:%d", idx, counterB, table.m[rowKey2].waitCount)
			wg.Done()
			return nil
		})
	}

	wg.Wait()

	cycle_dep <- nil
	_ = <- finish_dep

	if len(table.m) != 0 {
		t.Errorf("invalid map len, expected:0, actual:%d", len(table.m))
	}

}
