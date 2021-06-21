package module

import (
	"common/hls"
)

type Module interface {
	OnStreamError(stream *hls.Stream, err hls.Error)
	OnStreamM3u8New(stream *hls.Stream, m3u8 *hls.M3u8)
	OnStreamM3u8TsDownloaded(stream *hls.Stream, m3u8 *hls.M3u8)
	OnStreamTsNew(stream *hls.Stream, ts *hls.Ts)
}
