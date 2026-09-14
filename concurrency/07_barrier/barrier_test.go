package barrier

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

func TestNobodyPassesEarly(t *testing.T) {
	b := New(3)
	var passed int32

	for i := 0; i < 2; i++ {
		go func() {
			b.Wait()
			atomic.AddInt32(&passed, 1)
		}()
	}

	time.Sleep(100 * time.Millisecond)
	if got := atomic.LoadInt32(&passed); got != 0 {
		t.Fatalf("%d участников прошли барьер до прихода всех", got)
	}

	last := finished(func() {
		b.Wait()
		atomic.AddInt32(&passed, 1)
	})

	select {
	case <-last:
	case <-time.After(2 * time.Second):
		t.Fatal("последний участник не прошёл барьер")
	}

	deadline := time.Now().Add(2 * time.Second)
	for atomic.LoadInt32(&passed) != 3 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := atomic.LoadInt32(&passed); got != 3 {
		t.Fatalf("барьер прошли %d из 3", got)
	}
}

func TestSingleParticipant(t *testing.T) {
	b := New(1)
	select {
	case <-finished(b.Wait):
	case <-time.After(time.Second):
		t.Fatal("барьер на одного не должен блокировать")
	}
}

func TestReusableAcrossRounds(t *testing.T) {
	const parties, rounds = 6, 50
	b := New(parties)

	var round int32
	var wg sync.WaitGroup

	for i := 0; i < parties; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				b.Wait()
				atomic.AddInt32(&round, 1)
			}
		}()
	}

	select {
	case <-finished(wg.Wait):
	case <-time.After(30 * time.Second):
		t.Fatal("барьер завис между раундами")
	}

	if want := int32(parties * rounds); round != want {
		t.Fatalf("проходов %d, ожидалось %d", round, want)
	}
}

func TestRoundsDoNotOverlap(t *testing.T) {
	const parties, rounds = 4, 100
	b := New(parties)

	var inRound int32
	bad := int32(0)
	var wg sync.WaitGroup

	for i := 0; i < parties; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				b.Wait()

				if atomic.AddInt32(&inRound, 1) > parties {
					atomic.StoreInt32(&bad, 1)
				}
				time.Sleep(time.Microsecond)
				atomic.AddInt32(&inRound, -1)
			}
		}()
	}

	select {
	case <-finished(wg.Wait):
	case <-time.After(30 * time.Second):
		t.Fatal("тест завис")
	}

	if bad != 0 {
		t.Fatal("участники следующего раунда обогнали предыдущий")
	}
}
