package mutex

import (
	"sync"
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

func TestLockUnlock(t *testing.T) {
	var m Mutex
	m.Lock()
	m.Unlock()
	m.Lock()
	m.Unlock()
}

func TestMutualExclusion(t *testing.T) {
	const goroutines, iterations = 16, 20_000
	var m Mutex
	counter := 0

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				m.Lock()
				counter++
				m.Unlock()
			}
		}()
	}
	wg.Wait()

	if want := goroutines * iterations; counter != want {
		t.Fatalf("счётчик %d, ожидалось %d", counter, want)
	}
}

func TestOnlyOneInside(t *testing.T) {
	var m Mutex
	inside := 0
	bad := false

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 3000; j++ {
				m.Lock()
				inside++
				if inside != 1 {
					bad = true
				}
				inside--
				m.Unlock()
			}
		}()
	}
	wg.Wait()

	if bad {
		t.Fatal("в критической секции оказалось больше одной горутины")
	}
}

func TestLockWaitsForUnlock(t *testing.T) {
	var m Mutex
	m.Lock()

	second := finished(func() {
		m.Lock()
		m.Unlock()
	})

	select {
	case <-second:
		t.Fatal("второй Lock прошёл, пока мьютекс занят")
	case <-time.After(50 * time.Millisecond):
	}

	m.Unlock()

	select {
	case <-second:
	case <-time.After(2 * time.Second):
		t.Fatal("второй Lock не проснулся после Unlock")
	}
}

func TestEveryWaiterWakesUp(t *testing.T) {
	var m Mutex
	const waiters = 50

	m.Lock()

	var wg sync.WaitGroup
	for i := 0; i < waiters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Lock()
			time.Sleep(time.Millisecond)
			m.Unlock()
		}()
	}

	time.Sleep(50 * time.Millisecond)
	m.Unlock()

	select {
	case <-finished(wg.Wait):
	case <-time.After(10 * time.Second):
		t.Fatal("часть горутин осталась спать навсегда")
	}
}

func TestTryLock(t *testing.T) {
	var m Mutex

	if !m.TryLock() {
		t.Fatal("TryLock на свободном мьютексе должен получиться")
	}

	got := make(chan bool, 1)
	go func() { got <- m.TryLock() }()

	select {
	case ok := <-got:
		if ok {
			t.Fatal("TryLock на занятом мьютексе должен вернуть false")
		}
	case <-time.After(time.Second):
		t.Fatal("TryLock заблокировался, а не должен")
	}

	m.Unlock()
	if !m.TryLock() {
		t.Fatal("после Unlock мьютекс свободен")
	}
	m.Unlock()
}

func TestUnlockWithoutLockPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Unlock без Lock должен паниковать")
		}
	}()
	var m Mutex
	m.Unlock()
}

func TestHandoffUnderLoad(t *testing.T) {
	var m Mutex
	deadline := time.Now().Add(300 * time.Millisecond)
	done := make(chan int, 8)

	for i := 0; i < 8; i++ {
		go func() {
			passes := 0
			for time.Now().Before(deadline) {
				m.Lock()
				passes++
				m.Unlock()
			}
			done <- passes
		}()
	}

	for i := 0; i < 8; i++ {
		select {
		case passes := <-done:
			if passes == 0 {
				t.Fatal("горутина не получила мьютекс ни разу")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("горутина зависла")
		}
	}
}

func BenchmarkUncontended(b *testing.B) {
	var m Mutex
	for i := 0; i < b.N; i++ {
		m.Lock()
		m.Unlock()
	}
}

func BenchmarkContended(b *testing.B) {
	var m Mutex
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Lock()
			m.Unlock()
		}
	})
}

func BenchmarkStdlibContended(b *testing.B) {
	var m sync.Mutex
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Lock()
			m.Unlock()
		}
	})
}
