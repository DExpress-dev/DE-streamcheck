package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"product_code/check_stream/public"
	"product_code/check_stream/stream"

	log4plus "common/log4go"
)

type AddTaskRequest struct {
	StreamUrl     string `json:"url"`
	Delay         int    `json:"delay"`
	CheckSequence int    `json:"check_sequence"`
	Antileech     string `json:"antileech"`
	OnlyM3u8      int    `json:"only_m3u8"`
	Begin         string `json:"begin"`
	End           string `json:"end"`
}

func NewAddTaskRequest() *AddTaskRequest {
	return &AddTaskRequest{
		OnlyM3u8:      1,
		Delay:         30,
		CheckSequence: 0,
	}
}

func HttpAddTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(r.Body)
	req := NewAddTaskRequest()
	err := json.Unmarshal(bodyBytes, req)
	if err != nil {
		log4plus.Error("HttpAddTask: %s", err.Error())
		public.HttpError(w, -1, err.Error())
		return
	}

	log4plus.Debug("HttpAddTask: %#v", req)

	if req.StreamUrl == "" {
		err = fmt.Errorf("url empty")
		log4plus.Error("HttpAddTask: %s", err.Error())
		public.HttpError(w, -1, err.Error())
		return
	}

	if req.Begin == "" {
		err = fmt.Errorf("begin empty")
		log4plus.Error("HttpAddTask: %s", err.Error())
		public.HttpError(w, -1, err.Error())
		return
	}
	beginTime, err := public.TimeFromString(req.Begin)
	if err != nil {
		err = fmt.Errorf("begin invalid")
		log4plus.Error("HttpAddTask: %s", err.Error())
		public.HttpError(w, -1, err.Error())
		return
	}
	if req.End == "" {
		err = fmt.Errorf("end empty")
		log4plus.Error("HttpAddTask: %s", err.Error())
		public.HttpError(w, -1, err.Error())
		return
	}
	endTime, err := public.TimeFromString(req.End)
	if err != nil {
		err = fmt.Errorf("end invalid")
		log4plus.Error("HttpAddTask: %s", err.Error())
		public.HttpError(w, -1, err.Error())
		return
	}

	task := &stream.Task{
		StreamUrl:     req.StreamUrl,
		Delay:         req.Delay,
		CheckSequence: req.CheckSequence,
		Antileech:     req.Antileech,
		OnlyM3u8:      req.OnlyM3u8,
		Begin:         beginTime,
		End:           endTime,
	}
	err = stream.GetManager().AddTask(task)
	if err != nil {
		log4plus.Error("HttpAddTask: %s", err.Error())
		public.HttpError(w, -1, err.Error())
		return
	}

	bytes, _ := json.Marshal(&ResponseCommon{0, ""})
	w.Write(bytes)
}
