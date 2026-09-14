package waitgroup

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func finished(f func()) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		f()
		close(ch)
	}()
	return ch
}

func TestWaitOnZeroReturnsAtOnce(t *testing.T) {
	var wg WaitGroup
	select {
	case <-finished(wg.Wait):
	case <-time.After(time.Second):
		t.Fatal("Wait на нулевом счётчике не должен блокировать")
	}
}

func TestWaitsForEveryone(t *testing.T) {
	var wg WaitGroup
	var done int32

	const workers = 20
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			time.Sleep(20 * time.Millisecond)
			atomic.AddInt32(&done, 1)
			wg.Done()
		}()
	}

	select {
	case <-finished(wg.Wait):
	case <-time.After(5 * time.Second):
		t.Fatal("Wait не дождался")
	}

	if got := atomic.LoadInt32(&done); got != workers {
		t.Fatalf("Wait вернулся, когда закончили только %d из %d", got, workers)
	}
}

func TestManyWaiters(t *testing.T) {
	var wg WaitGroup
	wg.Add(1)

	var waiters sync.WaitGroup
	for i := 0; i < 30; i++ {
		waiters.Add(1)
		go func() {
			defer waiters.Done()
			wg.Wait()
		}()
	}

	time.Sleep(50 * time.Millisecond)
	wg.Done()

	select {
	case <-finished(waiters.Wait):
	case <-time.After(5 * time.Second):
		t.Fatal("не все ожидающие проснулись")
	}
}

func TestWaitBlocksUntilDone(t *testing.T) {
	var wg WaitGroup
	wg.Add(1)

	waiting := finished(wg.Wait)
	select {
	case <-waiting:
		t.Fatal("Wait вернулся раньше времени")
	case <-time.After(50 * time.Millisecond):
	}

	wg.Done()

	select {
	case <-waiting:
	case <-time.After(2 * time.Second):
		t.Fatal("Wait не проснулся после Done")
	}
}

func TestReuse(t *testing.T) {
	var wg WaitGroup

	for round := 0; round < 5; round++ {
		wg.Add(4)
		for i := 0; i < 4; i++ {
			go wg.Done()
		}

		select {
		case <-finished(wg.Wait):
		case <-time.After(2 * time.Second):
			t.Fatalf("раунд %d: Wait завис", round)
		}
	}
}

func TestNegativeCounterPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("уход счётчика в минус должен паниковать")
		}
	}()
	var wg WaitGroup
	wg.Done()
}

func TestStress(t *testing.T) {
	for round := 0; round < 200; round++ {
		var wg WaitGroup
		var counter int32

		wg.Add(8)
		for i := 0; i < 8; i++ {
			go func() {
				atomic.AddInt32(&counter, 1)
				wg.Done()
			}()
		}

		select {
		case <-finished(wg.Wait):
		case <-time.After(5 * time.Second):
			t.Fatalf("раунд %d завис", round)
		}

		if counter != 8 {
			t.Fatalf("раунд %d: счётчик %d", round, counter)
		}
	}
}
