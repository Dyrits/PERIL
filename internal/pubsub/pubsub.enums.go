package pubsub

type QueueType int

const (
	QueueTypeDurable QueueType = iota
	QueueTypeTransient
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)
