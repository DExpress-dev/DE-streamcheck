// stream_check project public.go
package public

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"product_code/check_stream/config"
	"strings"
	"time"
)

func UrlDir(url string) string {
	index := strings.LastIndex(url, "/")
	if index == -1 {
		return ""
	}
	return string(url[0:index])
}

//700/20180102/20180102T182312.ts -> localpath/700/20180102/20180102T182312.ts
//20180102T182312.ts -> localpath/20180102/20180102T182312.ts)
func GetTsLocalName(tsName string) string {
	var preFolder string
	if !strings.Contains(tsName, "/") {
		preFolder = time.Now().Format("20060102")
	}
	return preFolder + "/" + tsName
}

func GetTsLocalFile(localPath, tsName string) string {
	return localPath + "/" + GetTsLocalName(tsName)
}

//判断文件是否存在
func FileExist(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			return false
		}
	}
	return true
}

//获取当前路径
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}
	return strings.Replace(dir, "\\", "/", -1)
}

//分行
func SplitLine(str string) []string {
	return strings.FieldsFunc(str, func(s rune) bool {
		if s == '\n' || s == '\r' {
			return true
		}
		return false
	})
}

func SaveFile(content, localFile string) error {
	localPath := path.Dir(localFile)
	err := os.MkdirAll(localPath, os.ModePerm)
	if err != nil {
		return err
	}

	localFileTmp := localFile + ".tmp"
	file, err := os.Create(localFileTmp)
	if err != nil {
		return err
	}
	io.WriteString(file, content)
	file.Close()
	os.Rename(localFileTmp, localFile)
	return nil
}

// execute cmd line
func ShellExecute(s string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", s)

	var cout bytes.Buffer
	cmd.Stdout = &cout

	var cerr bytes.Buffer
	cmd.Stderr = &cerr

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return cout.String(), nil
}

func NotifyWechat(msg string) error {
	for _, user := range config.GetInstance().WechatUsers {
		strCmd := fmt.Sprintf(`./wechat --corpid=%s --corpsecret=%s --agentid=%s --msg="%s" --user=%s`, config.GetInstance().WechatCorpId, config.GetInstance().WechatCorpSecret, config.GetInstance().WechatAgentId, msg, user)
		_, err := ShellExecute(strCmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func NowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func HttpFormGetString(r *http.Request, param string) (value string, err error) {
	if len(r.Form[param]) <= 0 {
		return "", fmt.Errorf("param %s not found", param)
	}
	return strings.TrimSpace(r.Form[param][0]), nil
}

func HttpError(w http.ResponseWriter, result int, msg string) {
	w.Write([]byte(fmt.Sprintf("{\"result\":%d,\"msg\":\"%s\"}", result, msg)))
}

func TimeFromString(str string) (time.Time, error) {
	return time.ParseInLocation("2006-1-2 15:04:05", str, time.Local)
}
