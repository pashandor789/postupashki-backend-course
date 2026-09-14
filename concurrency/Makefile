PKGS := 01_spinlock 02_mutex 03_semaphore 04_waitgroup 05_once 06_rwmutex 07_barrier

.PHONY: all test race stress bench vet check clean one $(PKGS) \
        spinlock mutex semaphore waitgroup once rwmutex barrier

all: check

# Всё для одного примитива: форматирование, vet, тесты, гонки, стресс.
#   make mutex        или   make 02_mutex
one:
	@gofmt -l $(DIR)
	go vet ./$(DIR)/
	go test ./$(DIR)/
	go test -race ./$(DIR)/
	go test -race -count=20 -timeout=10m ./$(DIR)/
	@echo "$(DIR): всё зелёное"

spinlock:  ; @$(MAKE) --no-print-directory one DIR=01_spinlock
mutex:     ; @$(MAKE) --no-print-directory one DIR=02_mutex
semaphore: ; @$(MAKE) --no-print-directory one DIR=03_semaphore
waitgroup: ; @$(MAKE) --no-print-directory one DIR=04_waitgroup
once:      ; @$(MAKE) --no-print-directory one DIR=05_once
rwmutex:   ; @$(MAKE) --no-print-directory one DIR=06_rwmutex
barrier:   ; @$(MAKE) --no-print-directory one DIR=07_barrier

$(PKGS):
	@$(MAKE) --no-print-directory one DIR=$@

test:
	go test ./...

race:
	go test -race ./...

stress:
	go test -race -count=100 -timeout=30m ./...

bench:
	go test -bench=. -benchmem -run=^$$ ./...

vet:
	@gofmt -l .
	go vet ./...

check: vet race stress
	@echo "всё зелёное"

clean:
	go clean -testcache
