package stream

import (
	"common/hls"
	"net/url"
	"path"
	"product_code/check_stream/alarm"
	"product_code/check_stream/config"
	"strconv"
	"strings"
	"sync"
	"time"

	log4plus "common/log4go"
)

const (
	StreamStatusInitialized = iota
	StreamStatusRunning
	StreamStatusNormal
	StreamStatusDisconnected
	StreamStatusStoped
)

type CheckStreamEx struct {
	Key           string
	StreamUrlRaw  string
	StreamUrl     string
	UrlInfo       *url.URL
	Timeout       int
	CheckSequence int
	OnlyM3u8      int
	DelayMs       int64
	Antileech     string
	Name          string

	M3u8          *hls.M3u8
	HlsStream     *hls.Stream
	CacheSizeMs   int64 //当前缓存时间(ms)
	LastSequence  int64
	Status        int
	M3u8FailCount int
	LastM3u8Time  time.Time

	Lock    sync.Mutex //临界区
	TsArray []*hls.Ts  //ts的数组
	TsMap   map[string]*hls.Ts
}

type CheckStreamStatus struct {
	Key       string `json:"key"`
	StreamUrl string `json:"url"`
	Status    int    `json:"status"`
}

func NewCheckStream(key, streamUrl string) *CheckStreamEx {
	strm := &CheckStreamEx{
		Key:           key,
		StreamUrlRaw:  streamUrl,
		StreamUrl:     streamUrl,
		Timeout:       5,
		CheckSequence: 0,
		OnlyM3u8:      -1,
		DelayMs:       30 * 1000,
		LastM3u8Time:  time.Now(),
		TsMap:         make(map[string]*hls.Ts),
		Status:        StreamStatusInitialized,
	}
	strm.UrlInfo, _ = url.Parse(streamUrl)
	return strm
}

func (strm *CheckStreamEx) SetDelayMs(delayMs int64) *CheckStreamEx {
	strm.DelayMs = delayMs
	return strm
}

func (strm *CheckStreamEx) SetCheckSequence(checkSequence int) *CheckStreamEx {
	strm.CheckSequence = checkSequence
	return strm
}

func (strm *CheckStreamEx) SetAntileech(antileech string) *CheckStreamEx {
	strm.Antileech = antileech
	return strm
}

func (strm *CheckStreamEx) SetOnlyM3u8(onlyM3u8 int) *CheckStreamEx {
	strm.OnlyM3u8 = onlyM3u8
	return strm
}

func (strm *CheckStreamEx) SetName(name string) *CheckStreamEx {
	strm.Name = name
	return strm
}

func (strm *CheckStreamEx) IsHls() bool {
	return strings.HasSuffix(strm.UrlInfo.Path, "m3u8")
}

func (strm *CheckStreamEx) OnStreamStatistics(hlsStream *hls.Stream, statistics *hls.StreamStatistics) {
	//	log4plus.Error("[%s]OnStreamStatistics m3u8Url=%s m3u8downloadcount=%d", hlsStream.Key, hlsStream.M3u8Url, hlsStream.M3u8DownloadCount)
}

func (strm *CheckStreamEx) OnStreamError(hlsStream *hls.Stream, err hls.Error) {
	log4plus.Error("[%s]OnStreamError m3u8Url=%s err=%s", hlsStream.Key, hlsStream.M3u8Url, err.Err)
	switch err.Code {
	case hls.ErrorCodeM3u8DownloadFail:
		strm.M3u8FailCount++
		if strm.M3u8FailCount >= config.GetInstance().DownloadRetryCount {
			alarm.NotifyWechat(hlsStream.Key, strm.Name, err.Code, "m3u8 download fail url=%s err=%s", hlsStream.M3u8Url, err.Err)
			strm.M3u8FailCount = 0
		}
	case hls.ErrorCodeM3u8FormatError:
		alarm.NotifyWechat(hlsStream.Key, strm.Name, err.Code, "m3u8 format error url=%s", hlsStream.M3u8Url)
	case hls.ErrorCodeTsDownloadFail:
		ts := err.Data.(*hls.Ts)
		alarm.NotifyWechat(hlsStream.Key, strm.Name, err.Code, "ts download fail url=%s err=%s", ts.TsUrl, err.Err)
	}
}

func (strm *CheckStreamEx) OnStreamM3u8New(hlsStream *hls.Stream, m3u8 *hls.M3u8) {
	log4plus.Debug("[%s]OnStreamM3u8New m3u8Url=%s m3u8String=%s", hlsStream.Key, hlsStream.M3u8Url, m3u8.M3u8String)

	if m3u8.IsSecond() {
		strm.M3u8FailCount = 0
		strm.LastM3u8Time = time.Now()
		if 0 == strm.LastSequence {
			strm.LastSequence = m3u8.Sequence - 1
		}

		//check sequence
		if (strm.CheckSequence == 1) && (strm.LastSequence+1 != m3u8.Sequence) {
			alarm.NotifyWechat(hlsStream.Key, strm.Name, hls.ErrorCodeM3u8SequenceNotContinuous, "m3u8 sequence not continuous url=%s last=%d current=%d",
				m3u8.M3u8Url,
				strm.LastSequence,
				m3u8.Sequence)
			log4plus.Warn("[%s]m3u8 sequence not continuous m3u8Name=%s last=%d current=%d",
				hlsStream.Key,
				hlsStream.M3u8Name,
				strm.LastSequence,
				m3u8.Sequence)
		}
		strm.LastSequence = m3u8.Sequence

		//check ts repeat
		tsMap := make(map[string]bool)
		for _, ts := range m3u8.TsEntries {
			if tsMap[ts.Name] == true {
				alarm.NotifyWechat(hlsStream.Key, strm.Name, hls.ErrorCodeM3u8TsRepeat, "m3u8 ts repeat url=%s ts=%s",
					m3u8.M3u8Url,
					ts.Name)
				log4plus.Warn("[%s]m3u8 ts repeat url=%s ts=%s",
					hlsStream.Key,
					m3u8.M3u8Url,
					ts.Name)
			}
			tsMap[ts.Name] = true
		}
	}
}

func (strm *CheckStreamEx) OnStreamM3u8TsDownloaded(hlsStream *hls.Stream, m3u8 *hls.M3u8) {
	//	log4plus.Debug("[%s]OnStreamM3u8TsDownloaded m3u8Name=%s", hlsStream.Key, hlsStream.M3u8Name)
}

func (strm *CheckStreamEx) OnStreamTsNew(hlsStream *hls.Stream, ts *hls.Ts) {
	log4plus.Debug("[%s]OnStreamTsNew m3u8Name=%s tsUrl=%s tsDuration=%d", strm.Key, hlsStream.M3u8Name, ts.TsUrl, ts.Duration)
	//过滤异常的ts
	duration := int(ts.Duration / 1000)
	if duration > config.GetInstance().StreamDurationMax || duration < config.GetInstance().StreamDurationMin {
		alarm.NotifyWechat(hlsStream.Key, strm.Name, hls.ErrorCodeTsDurationAbnormal, "ts duration abnormal duration=%d url=%s", duration, ts.TsUrl)
		log4plus.Warn("[%s]ts duration abnormal duration=%d url=%s", hlsStream.Key, duration, ts.TsUrl)
	} else if ts.Size <= 0 {
		alarm.NotifyWechat(hlsStream.Key, strm.Name, hls.ErrorCodeTsSizeZero, "ts size zero url=%s", ts.TsUrl)
		log4plus.Warn("[%s]ts size zero url=%s", hlsStream.Key, ts.TsUrl)
	} else {
		strm.TsAdd(ts)
	}
}

func (strm *CheckStreamEx) checkStatus() {
	if time.Now().Sub(strm.LastM3u8Time) > time.Duration(strm.DelayMs)*time.Millisecond && strm.Status != StreamStatusDisconnected {
		strm.Status = StreamStatusDisconnected
		strm.LastSequence = 0
		alarm.NotifyWechat(strm.Key, strm.Name, hls.ErrorCodeStreamDisconnected, "stream disconnect url=%s", strm.HlsStream.M3u8Url)
		log4plus.Warn("[%s] CheckStreamContext.play disconnected m3u8=%s", strm.Key, strm.HlsStream.M3u8Name)
	}
}

func (strm *CheckStreamEx) Start() {
	go strm.start()
}

func (strm *CheckStreamEx) Stop() {
	strm.Status = StreamStatusStoped
}

func (strm *CheckStreamEx) start() {
	log4plus.Debug("stream started key=%s url=%s", strm.Key, strm.StreamUrlRaw)
	if strm.IsHls() {
		strm.HlsStream = hls.NewStream(strm.Key, strm.StreamUrl, "").
			SetTimeout(strm.Timeout).
			SetRetryCount(config.GetInstance().DownloadRetryCount).
			SetRetryWait(config.GetInstance().DownloadRetryWait).
			SetAntileechRemote(strm.Antileech).
			SetCallback(hls.StreamCallback{
				strm.OnStreamError,
				strm.OnStreamM3u8New,
				strm.OnStreamM3u8TsDownloaded,
				strm.OnStreamTsNew,
				strm.OnStreamStatistics})
		if strm.OnlyM3u8 > 0 {
			strm.HlsStream.SetOnlyM3u8(true)
		}

		var err error
		for {
			if strm.Status == StreamStatusStoped {
				if strm.HlsStream != nil {
					strm.HlsStream.StopAndWait()
				}
				break
			}
			time.Sleep(time.Second)

			if strm.M3u8 != nil && strm.M3u8.IsSecond() {
				strm.checkStatus()

				continue
			}

			strm.M3u8, err = strm.HlsStream.DownloadM3u8()
			if err != nil {
				strm.OnStreamError(strm.HlsStream, hls.Error{hls.ErrorCodeM3u8DownloadFail, nil, err.Error()})
				log4plus.Error("[%s]CheckStream.Run M3u8Download err=%s ulr=%s timeout=%d",
					strm.Key, err.Error(), strm.HlsStream.M3u8Url, strm.HlsStream.Timeout)
				continue
			}

			//判断是1级m3u8还是2级m3u8
			if strm.M3u8.IsTop() {
				if strm.Status == StreamStatusInitialized {
					strm.Status = StreamStatusRunning

					remotePath := strm.HlsStream.M3u8UrlInfo.Scheme + "://" + strm.HlsStream.M3u8UrlInfo.Host + path.Dir(strm.HlsStream.M3u8UrlInfo.Path)
					bandMax, bandMin := strm.M3u8.M3u8Entries[0], strm.M3u8.M3u8Entries[0]
					for _, entry := range strm.M3u8.M3u8Entries {
						if entry.Bandwidth > bandMax.Bandwidth {
							bandMax = entry
						}
						if entry.Bandwidth < bandMin.Bandwidth {
							bandMin = entry
						}
					}

					var m3u8Entry *hls.M3u8Entry = nil
					if "max" == config.GetInstance().StreamBandwidth {
						m3u8Entry = bandMax
					} else if "min" == config.GetInstance().StreamBandwidth {
						m3u8Entry = bandMin
					} else {
						band, _ := strconv.ParseInt(config.GetInstance().StreamBandwidth, 10, 64)
						for _, entry := range strm.M3u8.M3u8Entries {
							if band == entry.Bandwidth {
								m3u8Entry = entry
							}
						}
					}
					if nil == m3u8Entry {
						continue
					}

					hlsStream := strm.HlsStream.FindStream(m3u8Entry.Name)
					if hlsStream == nil {
						hlsStream = strm.HlsStream.AddStream(remotePath + "/" + m3u8Entry.RelativeToParent + "/" + m3u8Entry.Name + "?" + m3u8Entry.UrlInfo.RawQuery)
						hlsStream.Pull()
					}
					strm.Play()
				}
			} else if strm.M3u8.IsSecond() {
				strm.HlsStream.Pull()
				strm.Play()
			} else {
				log4plus.Error("[%s]CheckStream.Run unknown vile format ulr=%s timeout=%d",
					strm.Key, strm.HlsStream.M3u8Url, strm.HlsStream.Timeout)
			}
		}

		//TODO finalize
	} else {
		log4plus.Error("[%s] CheckStream.Run err=invalid file type!", strm.Key)
		return
	}
}

func (strm *CheckStreamEx) TsAdd(ts *hls.Ts) {
	strm.Lock.Lock()
	defer strm.Lock.Unlock()

	if _, exists := strm.TsMap[ts.Name]; exists {
		return
	}

	strm.TsArray = append(strm.TsArray, ts)
	strm.TsMap[ts.Name] = ts
	strm.CacheSizeMs += ts.Duration
}

func (strm *CheckStreamEx) popFront() *hls.Ts {
	if len(strm.TsArray) <= 0 {
		return nil
	}
	ts := strm.TsArray[0]
	strm.TsArray = strm.TsArray[1:]
	delete(strm.TsMap, ts.Name)
	strm.CacheSizeMs -= ts.Duration
	return ts
}

func (strm *CheckStreamEx) PopFront() *hls.Ts {
	strm.Lock.Lock()
	defer strm.Lock.Unlock()

	return strm.popFront()
}

func (strm *CheckStreamEx) TsCount() int {
	strm.Lock.Lock()
	defer strm.Lock.Unlock()

	return len(strm.TsArray)
}

func (strm *CheckStreamEx) IsFull() bool {
	strm.Lock.Lock()
	defer strm.Lock.Unlock()

	return strm.CacheSizeMs >= strm.DelayMs
}

func (strm *CheckStreamEx) IsEmpty() bool {
	strm.Lock.Lock()
	defer strm.Lock.Unlock()

	return strm.CacheSizeMs <= 0
}

func (strm *CheckStreamEx) Play() {
	go strm.play()
}

func (strm *CheckStreamEx) play() {
	var lastLogicEnd = time.Now().UnixNano() / 1e6
	var lastSleep = int64(0)
	for {
		var tWait int64 = 1000
		// log4plus.Debug("[%s] CheckStreamContext.play TsArrayCount=%d  m3u8=%s", strm.Key, strm.TsCount(), strm.HlsStream.M3u8Name)

		if strm.Status == StreamStatusStoped {
			break
		} else if strm.Status != StreamStatusNormal {
			if strm.IsFull() {
				lastStatus := strm.Status
				strm.Status = StreamStatusNormal
				if lastStatus == StreamStatusDisconnected {
					alarm.NotifyWechat(strm.Key, strm.Name, hls.ErrorCodeStreamRecover, "stream recover url=%s", strm.HlsStream.M3u8Url)
					log4plus.Warn("[%s] CheckStreamContext.play recover m3u8=%s", strm.Key, strm.HlsStream.M3u8Name)
				}
			}
		} else {
			ts := strm.PopFront()
			if ts != nil {
				tWait = ts.Duration
				log4plus.Debug("[%s] CheckStreamContext.play play m3u8=%s ts=%s", strm.Key, strm.HlsStream.M3u8Name, ts.Name)
			} else {
				strm.Status = StreamStatusDisconnected
				strm.LastSequence = 0
				alarm.NotifyWechat(strm.Key, strm.Name, hls.ErrorCodeStreamDisconnected, "stream disconnect url=%s", strm.HlsStream.M3u8Url)
				log4plus.Warn("[%s] CheckStreamContext.play disconnected m3u8=%s", strm.Key, strm.HlsStream.M3u8Name)
			}
		}

		tLogicEnd := time.Now().UnixNano() / 1e6

		//计算等待时间
		tElapse := tLogicEnd - lastLogicEnd - lastSleep
		tWaitModify := tWait - tElapse

		lastLogicEnd = tLogicEnd
		lastSleep = tWaitModify

		//开始等待
		time.Sleep(time.Duration(tWaitModify) * time.Millisecond)
	}
}

func (mgr *StreamManager) StreamStatus(streamUrl string) (statuses []*CheckStreamStatus) {
	{
		mgr.streamLock.Lock()
		for _, strm := range mgr.streamMap {
			if streamUrl == "" || streamUrl == strm.StreamUrl {
				statuses = append(statuses, &CheckStreamStatus{
					Key:       strm.Key,
					StreamUrl: strm.StreamUrl,
					Status:    strm.Status,
				})
			}
		}
		mgr.streamLock.Unlock()
	}

	{
		mgr.taskLock.Lock()
		for _, task := range mgr.tasks {
			if task.strm != nil {
				continue
			}

			if streamUrl == "" || streamUrl == task.StreamUrl {
				statuses = append(statuses, &CheckStreamStatus{
					Key:       task.Key,
					StreamUrl: task.StreamUrl,
					Status:    StreamStatusInitialized,
				})
			}
		}
		mgr.taskLock.Unlock()
	}

	return statuses
}
