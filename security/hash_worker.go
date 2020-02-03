package security

import (
	"context"

	"git.condensat.tech/bank"

	"golang.org/x/crypto/argon2"
)

type HasherWorker struct {
	jobs   chan job
	time   int
	memory int
	thread int
}

func NewHasherWorker(ctx context.Context, time, memory, thread int) bank.Worker {
	return &HasherWorker{
		time:   time,
		memory: memory,
		thread: thread,
	}
}

func (p *HasherWorker) Run(ctx context.Context, numWorkers int) {
	if len(p.jobs) != 0 {
		panic("Worker is already running")
	}
	p.jobs = make(chan job, numWorkers)

	hashWorkerDaemon(ctx, p.jobs, p.time, p.thread, p.memory)
}

type job struct {
	password []byte
	salt     []byte
	hash     chan []byte
}

func hashWorkerDaemon(ctx context.Context, queue chan job, time, thread, memory int) {
	for {
		select {
		case j := <-queue:
			j.hash <- argon2.Key(j.password, j.salt, uint32(time), uint32(memory), uint8(thread), 32)

		case <-ctx.Done():
			return
		}
	}
}

func (p *HasherWorker) doHash(salt, password []byte) []byte {
	jobs := p.jobs

	hash := make(chan []byte)
	go func() {
		jobs <- job{
			salt:     salt,
			password: password,
			hash:     hash,
		}
	}()

	return <-hash
}
