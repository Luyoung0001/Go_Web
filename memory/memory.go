package memory

import (
	"Go_Web/session"
	"container/list"
	"sync"
	"time"
)

// Provider 结构体

type Provider struct {
	lock     sync.Mutex               // 用来锁
	sessions map[string]*list.Element // 用来存储在内存
	list     *list.List               // 用来做gc
}

// Provider 实现接口 Provider

func (pder *Provider) SessionInit(sid string) (session.Session, error) {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	v := make(map[interface{}]interface{}, 0)
	newsess := &SessionStore{sid: sid, timeAccessed: time.Now(), value: v}
	element := pder.list.PushBack(newsess)
	pder.sessions[sid] = element
	return newsess, nil
}

func (pder *Provider) SessionRead(sid string) (session.Session, error) {
	if element, ok := pder.sessions[sid]; ok {
		return element.Value.(*SessionStore), nil
	} else {
		sess, err := pder.SessionInit(sid)
		return sess, err
	}
}

func (pder *Provider) SessionDestroy(sid string) error {
	if element, ok := pder.sessions[sid]; ok {
		delete(pder.sessions, sid)
		pder.list.Remove(element)
		return nil
	}
	return nil
}

func (pder *Provider) SessionGC(maxlifetime int64) {
	pder.lock.Lock()
	defer pder.lock.Unlock()

	for {
		element := pder.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*SessionStore).timeAccessed.Unix() + maxlifetime) < time.Now().Unix() {
			pder.list.Remove(element)
			delete(pder.sessions, element.Value.(*SessionStore).sid)
		} else {
			break
		}
	}
}

func (pder *Provider) SessionUpdate(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if element, ok := pder.sessions[sid]; ok {
		element.Value.(*SessionStore).timeAccessed = time.Now()
		pder.list.MoveToFront(element)
		return nil
	}
	return nil
}

var pder = &Provider{list: list.New()}

// SessionStore 结构体

type SessionStore struct {
	sid          string                      // session id唯一标识
	timeAccessed time.Time                   // 最后访问时间
	value        map[interface{}]interface{} // 值
}

// SessionStore 实现 Session 接口

func (st *SessionStore) Set(key, value interface{}) error {
	st.value[key] = value
	err := pder.SessionUpdate(st.sid)
	if err != nil {
		return err
	}
	return nil
}

func (st *SessionStore) Get(key interface{}) interface{} {
	err := pder.SessionUpdate(st.sid)
	if err != nil {
		return nil
	}
	if v, ok := st.value[key]; ok {
		return v
	} else {
		return nil
	}
}

func (st *SessionStore) Delete(key interface{}) error {
	delete(st.value, key)
	err := pder.SessionUpdate(st.sid)
	if err != nil {
		return err
	}
	return nil
}

func (st *SessionStore) SessionID() string {
	return st.sid
}

func init() {
	pder.sessions = make(map[string]*list.Element, 0)
	session.Register("memory", pder)
}
