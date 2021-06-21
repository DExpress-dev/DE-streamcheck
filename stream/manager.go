package stream

import (
	"common/config/goini"
	"encoding/json"
	"io/ioutil"
	"os"
	"product_code/check_stream/config"
	"product_code/check_stream/module"
	"product_code/check_stream/public"
	"strings"
	"sync"
	"sync/atomic"

	log4plus "common/log4go"
)

type StreamManager struct {
	streamLock sync.Mutex
	streamMap  map[string]*CheckStreamEx //下载的对象

	taskLock sync.Mutex
	tasks    map[string]*Task

	modules []module.Module

	IdSeed int64
}

var manager *StreamManager

func init() {
	manager = NewStreamManager()
}

func GetManager() *StreamManager {
	return manager
}

func NewStreamManager() *StreamManager {
	log4plus.Info("NewStreamManager")
	mgr := &StreamManager{
		streamMap: make(map[string]*CheckStreamEx),
		tasks:     make(map[string]*Task),
	}
	return mgr
}

func (mgr *StreamManager) StreamAdd(strm *CheckStreamEx) {
	timeout := int(strm.DelayMs)/1000/(config.GetInstance().DownloadRetryCount+1) - config.GetInstance().DownloadRetryWait
	if timeout < config.GetInstance().DownloadTimeoutMin {
		timeout = config.GetInstance().DownloadTimeoutMin
	}
	strm.Timeout = timeout

	mgr.streamLock.Lock()
	defer mgr.streamLock.Unlock()
	mgr.streamMap[strm.Key] = strm
}

func (mgr *StreamManager) LoadConfig() {
	log4plus.Info("StreamManager.LoadConfig")

	urlPath := public.GetCurrentDirectory() + "/" + config.GetInstance().StreamUrlFile
	if !public.FileExist(urlPath) {
		return
	}

	urlIni := goini.Init(urlPath)
	sessions := urlIni.ReadSessions()
	for _, key := range sessions {
		streamUrl := strings.TrimSpace(urlIni.Read_string(key, "url", ""))
		if key == "" || streamUrl == "" {
			continue
		}

		delay := urlIni.ReadInt(key, "delay", 30)
		checkSequence := urlIni.ReadInt(key, "check_sequence", -1)
		if checkSequence < 0 {
			checkSequence = config.GetInstance().StreamCheckSequence
		}
		antileech := urlIni.Read_string(key, "antileech", "")
		onlyM3u8 := urlIni.ReadInt(key, "only_m3u8", -1)
		name := urlIni.Read_string(key, "name", "")

		mgr.StreamAdd(NewCheckStream(key, streamUrl).
			SetDelayMs(int64(delay) * 1000).
			SetCheckSequence(checkSequence).
			SetAntileech(antileech).
			SetName(name).
			SetOnlyM3u8(onlyM3u8))
	}
	return
}

func (mgr *StreamManager) LoadModule() {
	//	mgr.modules = append(mgr.modules, &module.ModuleM3u8TsDurationMismatch{})
}

func (mgr *StreamManager) LoadTask() (err error) {
	fTask, err := os.Open(public.GetCurrentDirectory() + "/" + TaskFile)
	if err != nil {
		return err
	}
	defer fTask.Close()

	taskBytes, err := ioutil.ReadAll(fTask)
	if err != nil {
		return err
	}
	json.Unmarshal(taskBytes, &mgr.tasks)
	for _, task := range mgr.tasks {
		if task.Id > mgr.IdSeed {
			mgr.IdSeed = task.Id
		}
		log4plus.Info("task loaded %#v", task)
	}
	return nil
}

func (mgr *StreamManager) Initailzie() {
	mgr.LoadConfig()
	mgr.LoadModule()
	mgr.LoadTask()
}

func (mgr *StreamManager) Start() {
	for _, strm := range mgr.streamMap {
		strm.Start()
	}
}

func (mgr *StreamManager) IdGenerate() int64 {
	return atomic.AddInt64(&mgr.IdSeed, 1)
}

func (mgr *StreamManager) DeleteStream(key string) {
	mgr.streamLock.Lock()
	defer mgr.streamLock.Unlock()

	delete(mgr.streamMap, key)
}
