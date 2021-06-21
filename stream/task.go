package stream

import (
	"encoding/json"
	"product_code/check_stream/public"
	"strconv"
	"time"

	log4plus "common/log4go"
)

const (
	TaskStatusInit    = 0
	TaskStatusRunning = 1
	TaskStatusEnd     = -1
)

const TaskFile = "task.json"
const TaskPrefix = "task"

type Task struct {
	Id            int64     `json:"Id"`
	Key           string    `json:"Key"`
	StreamUrl     string    `json:"StreamUrl"`
	Delay         int       `json:"Delay"`
	CheckSequence int       `json:"CheckSequence"`
	Antileech     string    `json:"Antileech"`
	OnlyM3u8      int       `json:"OnlyM3u8"`
	Begin         time.Time `json:"Begin"`
	End           time.Time `json:"End"`

	strm   *CheckStreamEx
	status int
}

func (mgr *StreamManager) AddTask(task *Task) (err error) {
	log4plus.Debug("AddTask: %#v", task)

	task.Id = mgr.IdGenerate()
	task.Key = TaskPrefix + strconv.FormatInt(task.Id, 10)

	mgr.taskLock.Lock()
	defer mgr.taskLock.Unlock()
	mgr.tasks[task.Key] = task
	mgr.saveTask()
	return nil
}

func (mgr *StreamManager) saveTask() (err error) {
	taskBytes, err := json.Marshal(mgr.tasks)
	if err != nil {
		return err
	}
	err = public.SaveFile(string(taskBytes), public.GetCurrentDirectory()+"/"+TaskFile)
	if err != nil {
		return err
	}
	return nil
}

func (mgr *StreamManager) SaveTask() (err error) {
	mgr.taskLock.Lock()
	defer mgr.taskLock.Unlock()

	return mgr.saveTask()
}

func (mgr *StreamManager) CheckTask() {
	mgr.taskLock.Lock()
	defer mgr.taskLock.Unlock()

	tNow := time.Now()

	for key, task := range mgr.tasks {
		if TaskStatusInit == task.status && task.Begin.Before(tNow) && task.End.After(tNow) {
			log4plus.Debug("Task begin: [%s]%s", task.Key, task.StreamUrl)
			task.status = TaskStatusRunning
			task.strm = NewCheckStream(task.Key, task.StreamUrl).
				SetDelayMs(int64(task.Delay) * 1000).
				SetCheckSequence(task.CheckSequence).
				SetAntileech(task.Antileech).
				SetOnlyM3u8(task.OnlyM3u8)
			mgr.StreamAdd(task.strm)
			task.strm.Start()
		} else if task.End.Before(tNow) {
			log4plus.Debug("Task end: %s", task.Key)
			task.status = TaskStatusEnd
			if task.strm != nil {
				task.strm.Stop()
			}
			delete(mgr.tasks, key)
			mgr.saveTask()
		}
	}
}
