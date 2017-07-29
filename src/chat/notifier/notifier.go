package notifier

import (
	"chat/models"
	proto "proto/chat"
	"time"
	"utils/log"
	"utils/nats"
)

const (
	UnansweredTopic  = "chat.unanswered"
	SyncEventTopic   = "chat.sync_event"
	JoinTopic        = "chat.member.join"
	NewMessagesTopic = "chat.message.new"
	ReadedTopic      = "chat.message.readed"
	UpdatedTopic     = "chat.message.appended"
)

type Config struct {
	ForUser bool
	Delay   string
}

type notifier struct {
	syncChan       chan *proto.Chat
	unansweredChan chan proto.UnansweredNotify
	newChan        chan proto.NewMessageRequest
	joinChan       chan proto.NewChatMemberRequest
	readedChan     chan proto.MessageReadedRequest
	updateChan     chan proto.MessageAppendedRequest
	// unanswered messages workers
	forUser   []*worker
	forSeller []*worker
}

func New(config map[string]Config) models.Notifier {
	ret := notifier{
		syncChan:       make(chan *proto.Chat, 100),
		unansweredChan: make(chan proto.UnansweredNotify, 100),
		newChan:        make(chan proto.NewMessageRequest, 100),
		joinChan:       make(chan proto.NewChatMemberRequest, 100),
		readedChan:     make(chan proto.MessageReadedRequest, 100),
		updateChan:     make(chan proto.MessageAppendedRequest, 100),
	}

	for name, conf := range config {
		delay, err := time.ParseDuration(conf.Delay)
		if err != nil || delay < time.Second {
			log.Errorf("invalid delay '%v' in unanswered '%v'", conf.Delay, name)
			continue
		}

		n := &worker{
			name:       name,
			delay:      delay,
			forUser:    conf.ForUser,
			queue:      &Queue{},
			ansChan:    make(chan uint64, 100),
			addChan:    make(chan msgInfo, 100),
			chatMap:    make(map[uint64]chatInfo),
			notifyChan: ret.unansweredChan,
		}
		if conf.ForUser {
			ret.forUser = append(ret.forUser, n)
		} else {
			ret.forSeller = append(ret.forSeller, n)
		}
		go n.loop()
	}
	go ret.notifyLoop()
	return &ret
}

func (n *notifier) SyncEvent(chat *models.Conversation) {
	n.syncChan <- chat.Encode()
}

func (n *notifier) NewEvent(chat *models.Conversation, msgs ...*models.Message) {
	n.newChan <- proto.NewMessageRequest{
		Chat:     chat.Encode(),
		Messages: models.EncodeMessages(msgs),
	}

	// @TODO it's possible to optimize this in case when multiple messages are added at once
	// it's only meaningful for large threads from direct though
	for _, msg := range msgs {
		if msg.Member.Role == "SYSTEM" {
			continue
		}
		fromUser := msg.Member.Role == "CUSTOMER" || msg.Member.Role == "UNKNOWN"
		node := msgInfo{
			chatID: uint64(msg.ConversationID),
			msgID:  uint64(msg.ID),
		}

		var add, ans []*worker
		if fromUser {
			add = n.forSeller
			ans = n.forUser
		} else {
			add = n.forUser
			ans = n.forSeller
		}

		for _, w := range add {
			w.addChan <- node
		}
		for _, w := range ans {
			w.ansChan <- node.chatID
		}
	}
}

func (n *notifier) UpdateEvent(chat *models.Conversation, msg *models.Message) {
	n.updateChan <- proto.MessageAppendedRequest{
		Chat:    chat.Encode(),
		Message: msg.Encode(),
	}
}

func (n *notifier) JoinEvent(chat *models.Conversation, members ...*models.Member) {
	c := chat.Encode()
	for _, mbr := range members {
		n.joinChan <- proto.NewChatMemberRequest{
			Chat: c,
			User: mbr.Encode(),
		}
	}
}

func (n *notifier) ReadedEvent(chat *models.Conversation, messageID uint64, userID uint64) {
	n.readedChan <- proto.MessageReadedRequest{
		Chat:      chat.Encode(),
		MessageId: messageID,
		UserId:    userID,
	}
}

func (n *notifier) notifyLoop() {
	send := func(topic string, item interface{}) {
		log.Debug("sending %v notify: %+v", topic, item)
		err := nats.StanPublish(topic, item)
		// notifies have short lifetime, no need to retry probably
		if err != nil {
			log.Errorf("failed to send notify via stan: %v", err)
		}
	}

	for {
		select {
		case join := <-n.joinChan:
			send(JoinTopic, join)
		case readed := <-n.readedChan:
			send(ReadedTopic, readed)
		case sync := <-n.syncChan:
			send(SyncEventTopic, sync)
		case new := <-n.newChan:
			send(NewMessagesTopic, new)
		case updated := <-n.updateChan:
			send(UpdatedTopic, updated)
		case unanswered := <-n.unansweredChan:
			send(UnansweredTopic, &unanswered)
		}
	}
}

type worker struct {
	name    string
	delay   time.Duration
	forUser bool

	queue      *Queue
	ansChan    chan uint64
	addChan    chan msgInfo
	chatMap    map[uint64]chatInfo
	notifyChan chan proto.UnansweredNotify
}

type chatInfo struct {
	firstMsg uint64
	count    uint64
}

type msgInfo struct {
	chatID   uint64
	msgID    uint64
	notifyAT time.Time
}

func (n *worker) loop() {
	empty := true
	//there is no way to create timer without start it...
	timer := time.NewTimer(time.Second)
	if !timer.Stop() {
		<-timer.C
	}
	for {
		select {
		case msg := <-n.addChan:
			info, ok := n.chatMap[msg.chatID]
			if ok {
				info.count++
				n.chatMap[msg.chatID] = info
			} else {
				n.chatMap[msg.chatID] = chatInfo{
					firstMsg: msg.msgID,
					count:    1,
				}
				n.queue.Push(msgInfo{
					chatID:   msg.chatID,
					msgID:    msg.msgID,
					notifyAT: time.Now().Add(n.delay),
				})
				if empty {
					timer.Reset(n.delay)
					empty = false
				}
			}

		case chatID := <-n.ansChan:
			delete(n.chatMap, chatID)

		case <-timer.C:
			now := time.Now()
			var node msgInfo
			for {
				iface := n.queue.Pickup()
				if iface == nil {
					empty = true
					break
				}
				node = iface.(msgInfo)
				if node.notifyAT.After(now) {
					break
				}
				n.queue.Pop()
				info, ok := n.chatMap[node.chatID]
				if !ok || info.firstMsg != node.msgID {
					continue
				}
				delete(n.chatMap, node.chatID)
				select {
				case n.notifyChan <- proto.UnansweredNotify{
					ChatId:  node.chatID,
					Count:   info.count,
					Group:   n.name,
					ForUser: n.forUser,
				}:
				default:
					log.Errorf("notify channel capacity exceed, notify dropped")
				}
			}
			if !empty {
				timer.Reset(node.notifyAT.Sub(now))
			}
		}
	}
}
