package semaphore

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

func TestAcquireRelease(t *testing.T) {
	s := New(2)

	if s.Available() != 2 {
		t.Fatalf("Available вернул %d, ожидалось 2", s.Available())
	}

	s.Acquire()
	s.Acquire()

	if s.Available() != 0 {
		t.Fatalf("Available вернул %d, ожидалось 0", s.Available())
	}

	s.Release()
	s.Release()

	if s.Available() != 2 {
		t.Fatalf("Available вернул %d, ожидалось 2", s.Available())
	}
}

func TestBlocksWhenEmpty(t *testing.T) {
	s := New(1)
	s.Acquire()

	second := finished(func() {
		s.Acquire()
		s.Release()
	})

	select {
	case <-second:
		t.Fatal("Acquire прошёл, хотя разрешений нет")
	case <-time.After(50 * time.Millisecond):
	}

	s.Release()

	select {
	case <-second:
	case <-time.After(2 * time.Second):
		t.Fatal("Acquire не проснулся после Release")
	}
}

func TestNeverMoreThanLimit(t *testing.T) {
	const limit = 3
	s := New(limit)

	var inside, peak int32
	var wg sync.WaitGroup

	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Acquire()

			now := atomic.AddInt32(&inside, 1)
			for {
				old := atomic.LoadInt32(&peak)
				if now <= old || atomic.CompareAndSwapInt32(&peak, old, now) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			atomic.AddInt32(&inside, -1)

			s.Release()
		}()
	}

	select {
	case <-finished(wg.Wait):
	case <-time.After(10 * time.Second):
		t.Fatal("горутины не дождались разрешений")
	}

	if peak > limit {
		t.Fatalf("одновременно внутри было %d, лимит %d", peak, limit)
	}
	if peak < 2 {
		t.Fatalf("параллелизма не случилось вовсе: пик %d", peak)
	}
}

func TestTryAcquire(t *testing.T) {
	s := New(1)

	if !s.TryAcquire() {
		t.Fatal("TryAcquire при свободном разрешении должен получиться")
	}
	if s.TryAcquire() {
		t.Fatal("TryAcquire без разрешений должен вернуть false")
	}

	s.Release()

	if !s.TryAcquire() {
		t.Fatal("после Release разрешение снова доступно")
	}
}

func TestZeroPermits(t *testing.T) {
	s := New(0)

	if s.TryAcquire() {
		t.Fatal("у семафора на ноль разрешений брать нечего")
	}

	waiting := finished(s.Acquire)
	select {
	case <-waiting:
		t.Fatal("Acquire прошёл на пустом семафоре")
	case <-time.After(50 * time.Millisecond):
	}

	s.Release()

	select {
	case <-waiting:
	case <-time.After(2 * time.Second):
		t.Fatal("Acquire не проснулся")
	}
}

func TestAllWaitersWakeUp(t *testing.T) {
	s := New(0)
	const waiters = 30

	var wg sync.WaitGroup
	for i := 0; i < waiters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Acquire()
		}()
	}

	time.Sleep(50 * time.Millisecond)
	for i := 0; i < waiters; i++ {
		s.Release()
	}

	select {
	case <-finished(wg.Wait):
	case <-time.After(10 * time.Second):
		t.Fatal("не все ждущие проснулись")
	}
}

func BenchmarkAcquireRelease(b *testing.B) {
	s := New(1)
	for i := 0; i < b.N; i++ {
		s.Acquire()
		s.Release()
	}
}
