package queue

import (
	"chat/common"
	"chat/models"
	"proto/chat"
	"sync"
	"time"
	"utils/log"
	"utils/nats"
)

const (
	//EventUnreadNotify is a notification about a not answered message
	EventUnreadNotify = "core.notify.message"
)

type Waiter interface {
	//Push adds message to queue
	Push(*models.Message)
}

type queue struct {
	sync.Mutex
	stack   *common.Queue
	delay   time.Duration
	inbox   chan *models.Message
	chatMap map[uint]time.Time
}

func NewWaiter(delay time.Duration) Waiter {
	q := new(queue)
	q.delay = delay
	q.chatMap = make(map[uint]time.Time)
	q.inbox = make(chan *models.Message, 100)
	q.stack = &common.Queue{}
	q.start()
	return q
}

func (q *queue) Push(item *models.Message) {
	q.inbox <- item
}

func (q *queue) start() {
	go q.inboxLoop()
	go q.queueLoop()
}

func (q *queue) inboxLoop() {
	for msg := range q.inbox {
		if msg.Member == nil {
			log.Errorf("Message without a loaded member!")
			continue
		}
		if msg.Member.Role == chat.MemberRole_CUSTOMER.String() {
			q.answer(msg)
		} else {
			q.add(msg)
		}
	}
}

func (q *queue) add(msg *models.Message) {
	q.Lock()
	defer q.Unlock()
	//add a new message to queue only if a previous message got answered or notification sended
	if _, ok := q.chatMap[msg.ConversationID]; !ok {
		q.chatMap[msg.ConversationID] = msg.CreatedAt
		q.stack.Push(msg)
		log.Debug("Message %v added to queue", msg.ID)
	}

}

func (q *queue) answer(msg *models.Message) {
	q.Lock()
	defer q.Unlock()
	if t, ok := q.chatMap[msg.ConversationID]; ok && msg.CreatedAt.After(t) {
		delete(q.chatMap, msg.ConversationID)
	}

}

func (q *queue) queueLoop() {
	for {
		t, ok := q.nextOutTime()
		now := time.Now()
		if !ok {
			t = now.Add(time.Minute)
		}
		if t.After(now) {
			<-time.After(t.Sub(now))
		} else {
			q.notify()
		}

	}
}

func (q *queue) nextOutTime() (t time.Time, ok bool) {
	q.Lock()
	defer q.Unlock()

	if q.stack.Len() == 0 {
		return
	}

	t = q.stack.Pickup().(*models.Message).CreatedAt.Add(q.delay)
	ok = true
	return
}

func (q *queue) notify() {
	q.Lock()
	defer q.Unlock()

	msg := q.stack.Pop().(*models.Message)

	t, ok := q.chatMap[msg.ConversationID]
	//already answered or notified
	if !ok || !msg.CreatedAt.Equal(t) {
		return
	}

	delete(q.chatMap, msg.ConversationID)
	nats.StanPublish(EventUnreadNotify, msg.Encode())
	log.Debug("Notify about message %v", msg.ID)
}
