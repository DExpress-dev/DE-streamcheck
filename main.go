// stream_check project main.go
package main

/*
const char* build_time(void)
{
    static const char* psz_build_time = "["__DATE__ "  " __TIME__ "]";
    return psz_build_time;

}
*/
import "C"
import (
	"common/utils"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"product_code/check_stream/public"

	log4plus "common/log4go"
)

//版本号
var (
	ver       string = "1.2.11"
	exeName   string = ""
	pidFile   string = ""
	buildTime        = C.GoString(C.build_time())
)

type Flags struct {
	Help    bool
	Version bool
}

func (f *Flags) Init() {
	flag.BoolVar(&f.Help, "h", false, "help")
	flag.BoolVar(&f.Version, "V", false, "show version")
}

func (f *Flags) Check() (needReturn bool) {
	flag.Parse()

	if f.Help {
		flag.Usage()
		needReturn = true
	} else if f.Version {
		verString := "Check Stream Version: " + ver + "\r\n"
		verString += "compile time:" + buildTime + "\r\n"
		fmt.Println(verString)
		needReturn = true
	}

	return needReturn
}

var flags *Flags = &Flags{}

func init() {
	flags.Init()
	exeName = getExeName()
	pidFile = public.GetCurrentDirectory() + "/" + exeName + ".pid"
}

func getExeName() string {
	ret := ""
	ex, err := os.Executable()
	if err == nil {
		ret = filepath.Base(ex)
	}
	return ret
}

func setLog() {
	logJson := "log.json"
	set := false
	if bExist, _ := utils.PathExist(logJson); bExist {
		if err := log4plus.SetupLogWithConf(logJson); err == nil {
			set = true
		}
	}

	if !set {
		fileWriter := log4plus.NewFileWriter()
		fileWriter.SetPathPattern("log/" + exeName + "-%Y%M%D.log")
		log4plus.Register(fileWriter)
		log4plus.SetLevel(log4plus.DEBUG)
	}
}

func writePid() {
	public.SaveFile(fmt.Sprintf("%d", os.Getpid()), pidFile)
}

func main() {
	needReturn := flags.Check()
	if needReturn {
		return
	}

	//set log
	setLog()
	defer log4plus.Close()

	writePid()
	defer os.Remove(pidFile)

	log4plus.Info("main begin version=%s", ver)
	defer log4plus.Close()

	server := NewCheckServer()
	server.Run()
}
