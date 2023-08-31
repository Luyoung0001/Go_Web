package memory

import (
	"Go_Web/session"
	"container/list"
	"sync"
	"time"
)

// 创建全局 pder
var pder = &Provider{list: list.New()}

// Provider 结构体
// sessions 管理器

type Provider struct {
	lock     sync.Mutex               // 用来锁
	sessions map[string]*list.Element // 存放 sessionStores
	list     *list.List               // 用来做gc
}

// Provider 实现接口 Provider

func (pder *Provider) SessionInit(sid string) (session.Session, error) {
	// 根据 sid 创建一个 SessionStore
	pder.lock.Lock()
	defer pder.lock.Unlock()
	v := make(map[interface{}]interface{})
	// 同时更新两个字段
	newsess := &SessionStore{sid: sid, timeAccessed: time.Now(), value: v}
	// list 用于GC
	element := pder.list.PushBack(newsess)
	// 存放 kv
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

// 服务端 session 销毁

func (pder *Provider) SessionDestroy(sid string) error {
	if element, ok := pder.sessions[sid]; ok {
		delete(pder.sessions, sid)
		pder.list.Remove(element)
		return nil
	}
	return nil
}

// 回收过期的 cookie

func (pder *Provider) SessionGC(maxlifetime int64) {
	pder.lock.Lock()
	defer pder.lock.Unlock()

	for {
		element := pder.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*SessionStore).timeAccessed.Unix() + maxlifetime) < time.Now().Unix() {
			// 更新两者的值

			// 垃圾回收
			pder.list.Remove(element)
			// 删除 map 中的kv
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
		// 这里更新也就更新了个时间,这意味着 session 的生命得到了延长
		element.Value.(*SessionStore).timeAccessed = time.Now()
		pder.list.MoveToFront(element)
		return nil
	}
	return nil
}

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
	pder.sessions = make(map[string]*list.Element)
	// 注册一个名字为"memory"的管理器,这也是首先干的事
	session.Register("memory", pder)
}
