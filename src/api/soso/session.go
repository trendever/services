package soso

import (
	"github.com/igm/sockjs-go/sockjs"
	"sync"
	"utils/log"
	"utils/nats"
)

var Sessions = NewSessionRepository()

const NatsNewSessionSubject = "api.new_session"

type Session sockjs.Session

type SessionRepository interface {
	//Push adds session to collection
	Push(session Session, uid uint64) int
	//Get retries all active sessions for the user
	Get(uid uint64) []Session
	//Pull removes session object from collection
	Pull(session Session) bool
	//Size returns count of active sessions
	Size(uid uint64) int
}

func NewSessionRepository() SessionRepository {
	return &SessionRepositoryImpl{
		sessions: make(map[string]uint64),
		users:    make(map[uint64][]Session),
	}
}

type SessionRepositoryImpl struct {
	sync.Mutex
	sessions map[string]uint64
	users    map[uint64][]Session
}

func (s *SessionRepositoryImpl) Push(session Session, uid uint64) int {
	s.Lock()
	defer s.Unlock()
	log.Debug("Push session %s for user %v", session.ID(), uid)
	sessions, ok := s.users[uid]
	if !ok {
		sessions = make([]Session, 0)
	}
	if _, ok := s.sessions[session.ID()]; !ok {
		s.users[uid] = append(sessions, session)
		s.sessions[session.ID()] = uid
		nats.StanPublish(NatsNewSessionSubject, uid)
	}
	log.Debug("Session %s for user %v pushed, total %v", session.ID(), uid, len(s.users[uid]))
	return len(s.users[uid])
}

func (s *SessionRepositoryImpl) Get(uid uint64) []Session {
	s.Lock()
	defer s.Unlock()
	sessions, ok := s.users[uid]
	if !ok {
		sessions = make([]Session, 0)
	}
	return sessions
}

func (s *SessionRepositoryImpl) Pull(session Session) bool {
	s.Lock()
	defer s.Unlock()
	uid, ok := s.sessions[session.ID()]
	if !ok {
		return false
	}
	log.Debug("Pull session %s for user %v", session.ID(), uid)
	var found int
	for key, value := range s.users[uid] {
		if value.ID() == session.ID() {
			found = key
		}
	}
	s.users[uid] = append(s.users[uid][:found], s.users[uid][found+1:]...)
	delete(s.sessions, session.ID())
	log.Debug("Session %s for user %v pulled, total %v", session.ID(), uid, len(s.users[uid]))
	return true
}

func (s *SessionRepositoryImpl) Size(uid uint64) int {
	s.Lock()
	defer s.Unlock()
	sessions, ok := s.users[uid]
	if !ok {
		return 0
	}
	return len(sessions)
}
