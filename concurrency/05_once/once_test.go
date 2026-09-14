package once

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

func TestRunsOnlyOnce(t *testing.T) {
	var o Once
	calls := 0

	for i := 0; i < 10; i++ {
		o.Do(func() { calls++ })
	}

	if calls != 1 {
		t.Fatalf("функция вызвана %d раз", calls)
	}
}

func TestDone(t *testing.T) {
	var o Once

	if o.Done() {
		t.Fatal("до первого Do работа не сделана")
	}

	o.Do(func() {})

	if !o.Done() {
		t.Fatal("после Do работа сделана")
	}
}

func TestConcurrentCallsRunItOnce(t *testing.T) {
	var o Once
	var calls int32

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			o.Do(func() {
				atomic.AddInt32(&calls, 1)
				time.Sleep(10 * time.Millisecond)
			})
		}()
	}

	select {
	case <-finished(wg.Wait):
	case <-time.After(10 * time.Second):
		t.Fatal("часть горутин зависла в Do")
	}

	if calls != 1 {
		t.Fatalf("функция вызвана %d раз", calls)
	}
}

func TestDoWaitsForTheWinner(t *testing.T) {
	var o Once
	var ready int32

	first := finished(func() {
		o.Do(func() {
			time.Sleep(100 * time.Millisecond)
			atomic.StoreInt32(&ready, 1)
		})
	})

	time.Sleep(10 * time.Millisecond)

	second := finished(func() {
		o.Do(func() {})
		if atomic.LoadInt32(&ready) != 1 {
			t.Error("Do вернулся до того, как работа была закончена")
		}
	})

	select {
	case <-first:
	case <-time.After(5 * time.Second):
		t.Fatal("первый Do завис")
	}
	select {
	case <-second:
	case <-time.After(5 * time.Second):
		t.Fatal("второй Do завис")
	}
}

func TestPanicCountsAsDone(t *testing.T) {
	var o Once
	calls := 0

	func() {
		defer func() { recover() }()
		o.Do(func() {
			calls++
			panic("упало")
		})
	}()

	o.Do(func() { calls++ })

	if calls != 1 {
		t.Fatalf("после паники функция вызвана ещё раз: всего %d", calls)
	}
}

func TestStress(t *testing.T) {
	for round := 0; round < 300; round++ {
		var o Once
		var calls int32

		var wg sync.WaitGroup
		for i := 0; i < 16; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				o.Do(func() { atomic.AddInt32(&calls, 1) })
			}()
		}
		wg.Wait()

		if calls != 1 {
			t.Fatalf("раунд %d: вызовов %d", round, calls)
		}
	}
}
