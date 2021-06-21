package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"product_code/check_stream/public"
	"product_code/check_stream/stream"

	log4plus "common/log4go"
)

type StreamStatusRequest struct {
	StreamUrl string `json:"url"`
}

type StreamStatusResponse struct {
	ResponseCommon
	Streams []*stream.CheckStreamStatus `json:"streams"`
}

func NewStreamStatusRequest() *StreamStatusRequest {
	return &StreamStatusRequest{}
}

func HttpStreamStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(r.Body)
	req := NewStreamStatusRequest()
	err := json.Unmarshal(bodyBytes, req)
	if err != nil {
		log4plus.Error("HttpStreamStatus: %s", err.Error())
		public.HttpError(w, -1, err.Error())
		return
	}

	log4plus.Debug("HttpStreamStatus: %#v", req)

	resp := &StreamStatusResponse{}
	resp.Streams = stream.GetManager().StreamStatus(req.StreamUrl)

	bytes, _ := json.Marshal(resp)
	w.Write(bytes)
}
