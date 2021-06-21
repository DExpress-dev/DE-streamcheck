package config

import (
	"strconv"
	"strings"

	"github.com/widuu/goini"
)

type configInfo struct {
	Listen        string
	AlarmInterval int64

	StreamUrlFile       string
	StreamDurationMax   int
	StreamDurationMin   int
	StreamBandwidth     string
	StreamCheckSequence int

	DownloadTimeoutMin int
	DownloadRetryCount int
	DownloadRetryWait  int

	WechatCorpId     string
	WechatCorpSecret string
	WechatAgentId    string
	WechatUsers      []string
}

var _cfg *configInfo

func GetInstance() *configInfo {
	return _cfg
}

func init() {
	_cfg = new(configInfo)
	configIni := goini.Init("config.ini")

	_cfg.Listen = configIni.Read_string("common", "listen", "")
	_cfg.AlarmInterval = int64(configIni.ReadInt("common", "alarm_interval", 60))

	_cfg.StreamUrlFile = configIni.Read_string("stream", "url_file", "urls.ini")
	tsDurationRange := configIni.Read_string("stream", "duration_range", "1,30")
	fields := strings.Split(tsDurationRange, ",")
	if len(fields) < 2 {
		_cfg.StreamDurationMax = 30
		_cfg.StreamDurationMin = 1
	} else {
		var err error
		_cfg.StreamDurationMin, err = strconv.Atoi(fields[0])
		if err != nil {
			_cfg.StreamDurationMin = 1
		}
		_cfg.StreamDurationMax, err = strconv.Atoi(fields[1])
		if err != nil {
			_cfg.StreamDurationMax = 30
		}
	}
	if _cfg.StreamDurationMax < _cfg.StreamDurationMin {
		_cfg.StreamDurationMin, _cfg.StreamDurationMax = _cfg.StreamDurationMax, _cfg.StreamDurationMin
	}
	_cfg.StreamBandwidth = configIni.Read_string("stream", "bandwidth", "max")
	_cfg.StreamCheckSequence = configIni.ReadInt("stream", "check_sequence", 0)

	_cfg.DownloadTimeoutMin = configIni.ReadInt("download", "timeout_min", 5)
	_cfg.DownloadRetryCount = configIni.ReadInt("download", "retry_count", 3)
	_cfg.DownloadRetryWait = configIni.ReadInt("download", "retry_wait", 2)

	_cfg.WechatCorpId = configIni.Read_string("wechat", "corp_id", "wwdaae54803f75694b")
	_cfg.WechatCorpSecret = configIni.Read_string("wechat", "corp_secret", "r2YqOIHrP2kxf-Y1uc9hVlI8vhIY2Om6GIPT0_NRHaA")
	_cfg.WechatAgentId = configIni.Read_string("wechat", "agent_id", "1000016")
	wechatUserString := configIni.Read_string("wechat", "user", "ZhouJun")
	_cfg.WechatUsers = strings.Split(wechatUserString, ";")
}
