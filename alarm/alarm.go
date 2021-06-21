package alarm

import (
	"fmt"
	"product_code/check_stream/config"
	"product_code/check_stream/public"
	"sync"
	"time"

	log4plus "common/log4go"
)

type NotifyContext struct {
	SendedMessages map[int]*NotifyMessage
	Lock           sync.Mutex
}

type NotifyMessage struct {
	Key     string
	Name    string
	ErrCode int
	Msg     string

	LastSend int64
}

type NotifyManager struct {
	WechatMessages []*NotifyMessage
	NotifyContexts map[string]*NotifyContext

	Lock sync.Mutex
}

var notifyManager *NotifyManager

func init() {
	notifyManager = &NotifyManager{
		NotifyContexts: make(map[string]*NotifyContext),
	}
	notifyManager.Run()
}

func (manager *NotifyManager) Run() {
	go manager.run()
}

func MessageLen() int {
	notifyManager.Lock.Lock()
	defer notifyManager.Lock.Unlock()
	return len(notifyManager.WechatMessages)
}

func (manager *NotifyManager) run() {
	for {
		time.Sleep(time.Second)

		manager.Lock.Lock()
		msgs := manager.WechatMessages
		manager.WechatMessages = make([]*NotifyMessage, 0)
		manager.Lock.Unlock()

		manager.notify(msgs)
		// log4plus.Debug("NotifyManager message notify len=%d", len(msgs))
	}
}

func CheckNotify(msg *NotifyMessage) bool {
	return notifyManager.CheckNotify(msg)
}

func (manager *NotifyManager) CheckNotify(msg *NotifyMessage) bool {
	var ctx *NotifyContext = nil
	var find bool = false
	{
		manager.Lock.Lock()
		ctx, find = manager.NotifyContexts[msg.Key]
		if !find {
			ctx = &NotifyContext{SendedMessages: make(map[int]*NotifyMessage)}
			ctx.SendedMessages[msg.ErrCode] = msg
			msg.LastSend = time.Now().Unix()
			manager.NotifyContexts[msg.Key] = ctx
		}
		manager.Lock.Unlock()
	}
	if !find {
		return true
	}

	var msgSended *NotifyMessage = nil
	{
		ctx.Lock.Lock()
		msgSended, find = ctx.SendedMessages[msg.ErrCode]
		if !find {
			ctx.SendedMessages[msg.ErrCode] = msg
			msg.LastSend = time.Now().Unix()
		}
		ctx.Lock.Unlock()
	}
	if !find {
		return true
	}
	if time.Now().Unix()-msgSended.LastSend < config.GetInstance().AlarmInterval {
		return false
	}
	msgSended.LastSend = time.Now().Unix()
	return true
}

func (manager *NotifyManager) notify(msgs []*NotifyMessage) {
	for _, msg := range msgs {
		for _, user := range config.GetInstance().WechatUsers {
			strCmd := fmt.Sprintf(`./wechat --corpid=%s --corpsecret=%s --agentid=%s --msg="%s" --user=%s`, config.GetInstance().WechatCorpId, config.GetInstance().WechatCorpSecret, config.GetInstance().WechatAgentId, msg.Msg, user)
			log4plus.Debug("NotifyManager notify strCmd=%s", strCmd)
			public.ShellExecute(strCmd)
		}
	}
}

func (manager *NotifyManager) AddWechatMessage(msg *NotifyMessage) {
	manager.Lock.Lock()
	defer manager.Lock.Unlock()

	manager.WechatMessages = append(manager.WechatMessages, msg)
}

func NotifyWechat(key string, name string, errCode int, errfmt string, args ...interface{}) {
	timeString := fmt.Sprintf("[%s][%s]", public.NowString(), name)
	msg := fmt.Sprintf(timeString+errfmt, args...)
	notifyMessage := &NotifyMessage{
		Key:     key,
		Name:    name,
		ErrCode: errCode,
		Msg:     msg,
	}
	if !notifyManager.CheckNotify(notifyMessage) {
		return
	}
	log4plus.Debug("NotifyManager key=%s errCode=%d", key, errCode)
	notifyManager.AddWechatMessage(notifyMessage)
}
