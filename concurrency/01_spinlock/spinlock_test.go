package spinlock

import (
	"sync"
	"testing"
	"time"
)

type locker interface {
	Lock()
	Unlock()
	TryLock() bool
}

func each(t *testing.T, f func(t *testing.T, l locker)) {
	t.Helper()
	t.Run("Spinlock", func(t *testing.T) { f(t, &Spinlock{}) })
	t.Run("TTAS", func(t *testing.T) { f(t, &TTAS{}) })
}

func finished(f func()) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		f()
		close(ch)
	}()
	return ch
}

func TestLockUnlock(t *testing.T) {
	each(t, func(t *testing.T, l locker) {
		l.Lock()
		l.Unlock()
		l.Lock()
		l.Unlock()
	})
}

func TestMutualExclusion(t *testing.T) {
	each(t, func(t *testing.T, l locker) {
		const goroutines, iterations = 8, 20_000
		counter := 0

		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					l.Lock()
					counter++
					l.Unlock()
				}
			}()
		}
		wg.Wait()

		if want := goroutines * iterations; counter != want {
			t.Fatalf("счётчик %d, ожидалось %d", counter, want)
		}
	})
}

func TestOnlyOneInside(t *testing.T) {
	each(t, func(t *testing.T, l locker) {
		inside := 0
		bad := false

		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 2000; j++ {
					l.Lock()
					inside++
					if inside != 1 {
						bad = true
					}
					inside--
					l.Unlock()
				}
			}()
		}
		wg.Wait()

		if bad {
			t.Fatal("в критической секции оказалось больше одной горутины")
		}
	})
}

func TestTryLock(t *testing.T) {
	each(t, func(t *testing.T, l locker) {
		if !l.TryLock() {
			t.Fatal("TryLock на свободном замке должен получиться")
		}

		got := make(chan bool, 1)
		go func() { got <- l.TryLock() }()

		select {
		case ok := <-got:
			if ok {
				t.Fatal("TryLock на занятом замке должен вернуть false")
			}
		case <-time.After(time.Second):
			t.Fatal("TryLock заблокировался, а не должен")
		}

		l.Unlock()
		if !l.TryLock() {
			t.Fatal("после Unlock замок снова свободен")
		}
		l.Unlock()
	})
}

func TestLockWaitsForUnlock(t *testing.T) {
	each(t, func(t *testing.T, l locker) {
		l.Lock()

		second := finished(func() {
			l.Lock()
			l.Unlock()
		})

		select {
		case <-second:
			t.Fatal("второй Lock прошёл, пока замок занят")
		case <-time.After(50 * time.Millisecond):
		}

		l.Unlock()

		select {
		case <-second:
		case <-time.After(2 * time.Second):
			t.Fatal("второй Lock не проснулся после Unlock")
		}
	})
}

func TestUnlockWithoutLockPanics(t *testing.T) {
	each(t, func(t *testing.T, l locker) {
		defer func() {
			if recover() == nil {
				t.Fatal("Unlock без Lock должен паниковать")
			}
		}()
		l.Unlock()
	})
}

func BenchmarkUncontended(b *testing.B) {
	var l Spinlock
	for i := 0; i < b.N; i++ {
		l.Lock()
		l.Unlock()
	}
}

func BenchmarkContended(b *testing.B) {
	var l Spinlock
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Lock()
			l.Unlock()
		}
	})
}
