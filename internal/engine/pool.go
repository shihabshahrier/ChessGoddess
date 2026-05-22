package engine

import "context"

// Pool is a fixed set of engine processes for concurrent evaluation.
// Evaluate blocks until an idle engine is available or ctx is cancelled.
type Pool struct {
	idle chan *Engine
	all  []*Engine
}

// NewPool starts size engine processes and completes their UCI handshakes.
func NewPool(stockfishPath string, size int) (*Pool, error) {
	if size < 1 {
		size = 1
	}
	p := &Pool{idle: make(chan *Engine, size)}
	for i := 0; i < size; i++ {
		e, err := New(stockfishPath)
		if err != nil {
			_ = p.Close()
			return nil, err
		}
		if err := e.Initialize(); err != nil {
			_ = e.Close()
			_ = p.Close()
			return nil, err
		}
		p.all = append(p.all, e)
		p.idle <- e
	}
	return p, nil
}

// Evaluate borrows an idle engine, runs the search, and returns it to the pool.
func (p *Pool) Evaluate(ctx context.Context, fen string, depth, multipv int) (*Evaluation, error) {
	select {
	case e := <-p.idle:
		defer func() { p.idle <- e }()
		return e.Evaluate(fen, depth, multipv)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Size reports the number of engine processes in the pool.
func (p *Pool) Size() int { return len(p.all) }

// Close terminates every engine process.
func (p *Pool) Close() error {
	for _, e := range p.all {
		_ = e.Close()
	}
	return nil
}
