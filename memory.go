package broadcast

// Memory manages broadcasting in memory.
type Memory struct {
	router      *Router
	messages    chan *Message
	join, leave chan subscriber
	stop        chan interface{}
}

// memoryBroker creates new memory based message broker.
func memoryBroker() *Memory {
	return &Memory{
		router:   NewRouter(),
		messages: make(chan *Message),
		join:     make(chan subscriber),
		leave:    make(chan subscriber),
		stop:     make(chan interface{}),
	}
}

// Serve serves broker.
func (m *Memory) Serve() error {
	for {
		select {
		case ctx := <-m.join:
			ctx.done <- m.handleJoin(ctx)
		case ctx := <-m.leave:
			ctx.done <- m.handleLeave(ctx)
		case msg := <-m.messages:
			m.router.Dispatch(msg)
		case <-m.stop:
			return nil
		}
	}
}

func (m *Memory) handleJoin(sub subscriber) (err error) {
	if sub.pattern != "" {
		_, err = m.router.SubscribePattern(sub.upstream, sub.pattern)
		return nil
	}

	m.router.Subscribe(sub.upstream, sub.topics...)
	return nil
}

func (m *Memory) handleLeave(sub subscriber) error {
	if sub.pattern != "" {
		m.router.UnsubscribePattern(sub.upstream, sub.pattern)
		return nil
	}

	m.router.Unsubscribe(sub.upstream, sub.topics...)
	return nil
}

// Stop closes the consumption and disconnects broker.
func (m *Memory) Stop() {
	close(m.stop)
}

// Subscribe broker to one or multiple channels.
func (m *Memory) Subscribe(upstream chan *Message, topics ...string) error {
	ctx := subscriber{upstream: upstream, topics: topics, done: make(chan error)}

	m.join <- ctx
	return <-ctx.done
}

func (m *Memory) SubscribePattern(upstream chan *Message, pattern string) error {
	ctx := subscriber{upstream: upstream, pattern: pattern, done: make(chan error)}

	m.join <- ctx
	return <-ctx.done
}

// Unsubscribe broker from one or multiple channels.
func (m *Memory) Unsubscribe(upstream chan *Message, topics ...string) error {
	ctx := subscriber{upstream: upstream, topics: topics, done: make(chan error)}

	m.leave <- ctx
	return <-ctx.done
}

func (m *Memory) UnsubscribePattern(upstream chan *Message, pattern string) error {
	ctx := subscriber{upstream: upstream, pattern: pattern, done: make(chan error)}

	m.leave <- ctx
	return <-ctx.done
}

// Publish one or multiple Channel.
func (m *Memory) Publish(messages ...*Message) error {
	for _, msg := range messages {
		m.messages <- msg
	}

	return nil
}
