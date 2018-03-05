package common

// channel of channel for balancing tasks between workers
type WorkerPoolType chan WorkerInterface
