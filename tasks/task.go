package tasks

import (
	"fmt"
	"time"
)

func NewRetryPolicy(tries int64, backoff int64, delay int64) *RetryPolicy {
	return &RetryPolicy{
		tries:   tries,
		backoff: backoff,
		delay:   delay,
	}
}

type RetryPolicy struct {
	tries   int64
	backoff int64
	delay   int64
}

type Task struct {
	Policy *RetryPolicy
	F      func() error
}

func NewTask(f func() error, policy *RetryPolicy) *Task {
	if f == nil {
		panic("f cannot be nil, please provide task function `f` such a f() -> error")
	}
	return &Task{
		Policy: policy,
		F:      f,
	}
}

func (t *Task) Run() error {
	if t.Policy != nil {
		tries := t.Policy.tries
		delay := t.Policy.delay
		backoff := t.Policy.backoff
		var err error
		for i := tries; i > 0; i-- {
			err = t.F()
			if err != nil {
				if i > 1 {
					fmt.Printf("An error occured: %v. Retrying in %d seconds...\n", err, delay)
					time.Sleep(time.Second * time.Duration(delay))
					delay *= backoff
				}
			} else {
				return nil
			}
		}
		return fmt.Errorf("task failed after %d attempts: %w", tries, err)
	}

	return t.F()
}
