package bigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fym201/bigo/utl"
)

const (
	LogLevelInfo = iota + 1
	LogLevelDebug
	LogLevelError
	LogLevelNone
)

type SubConfig map[string]interface{}

type StaticOpt struct {
	Path        string `json:"PATH"`         //静态目录的路径
	Prefix      string `json:"PREFIX"`       //此目录在url中的前缀，如果不设置则为/
	SkipLogging bool   `json:"SKIP_LOGGING"` //默认在DEV,TEST模式下为false,在PROD模式下为true. 是否跳过log输出
	IndexFile   string `json:"INDEX_FILE"`   //如果指定，则以此文件为索引文件
}

//如果以【app -c configPath】的命令行形式指定了文件，那么直接加载这个文件，否则
//查找当前工作目录下的config.json
//查找app所在目录下的config.json
type Config struct {
	AppName  string `json:"APP_NAME"` //应用名称
	WorkDir  string `json:"WORKDIR"`  //工作目录,默认为运行目录
	LogDir   string `json:"LOGDIR"`   //日志目录,默认为$WORKDIR/log
	LogLevel int    `json:"LOGLEVEL"` //日志等级,0.info 1.debug 3.error 4.none,默认在DEV,TEST下为0,PROD下为3

	HttpAddr string `json:"HTTP_ADDR"` //http监听地址,默认在0.0.0.0
	HttpPort int    `json:"HTTP_PORT"` //http监听端口,默认为3000

	HttpsAddr     string `json:"HTTPS_ADDR"`      //https监听地址,默认在0.0.0.0
	HttpsPort     int    `json:"HTTPS_PORT"`      //https监听地址,默认为443
	HttpsCertFile string `json:"HTTPS_CERT_FILE"` //https证书文件路径
	HttpsKeyFile  string `json:"HTTPS_KEY_FILE"`  //https证书key文件路径
	EnableHttps   bool   `json:"ENABLE_HTTPS"`    //是否开启https服务，默认false
	ForceHttps    bool   `json:"FORCE_HTTPS"`     //是否强制将http请求转为https,默认为false

	EnableGzip bool `json:"ENABLE_GZIP"` //是否开启gzip,默认true,如果开启,并且客户端接受gzip的话责对会话进行gzip压缩
	ForceGzip  bool `json:"FORCE_GZIP"`  //是否强制开启gzip,默认false,如果开启,不管客户端接不接受,都会以gzip进行传输

	EnableMinify           bool        `json:"ENABLE_MINIFY"`             //默认为true,是否对.html .js .css 进行最小化处理
	StaticExtensionsToGzip []string    `json:"STATIC_EXTENSIONS_TO_GZIP"` //使用gzip进行压缩传输的静态文件后缀,受客户端协议或FORCE_GZIP影响
	Statics                []StaticOpt `json:"STATICS"`                   //静态目录,数组
	RunMode                string      `json:"RUN_MODE"`                  //运行模式，DEV为开发模式，PROD为发布模式,TEST为测试模式
	DevOpt                 *SubConfig  `json:"DEV"`                       //对于DEV模式下的配置，RUN_MODE为DEV时会覆盖顶级配置
	TestOpt                *SubConfig  `json:"TEST"`                      //对于TEST模式下的配置，RUN_MODE为TEST时会覆盖顶级配置
	ProdOpt                *SubConfig  `json:"PROD"`                      //对于PROD模式下的配置，RUN_MODE为PROD时会覆盖顶级配置

	CutomOpt *map[string]interface{} `json:"CUSTOM"` //自定义选项
}

var (
	_config  *Config
	workPath string //工作目录
	confPath string //配置文件目录
	appPath  string //二进制程序目录
)

//加载系统配置
func LoadConfig() (c *Config, err error) {
	defer func() {
		if err != nil {
			fmt.Println("\nCant not load config with error:[", err.Error(), "] \n......now use default config\n")
			_config = new(Config)
			checkConfig(_config)
			c = _config
		}
	}()

	appPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	workPath, _ = os.Getwd()
	workPath, _ = filepath.Abs(workPath)

	confPath = utl.GetCmdArg("c")
	if confPath == "" {
		confPath = filepath.Join(workPath, "config.json")
		if !utl.IsExist(confPath) {
			confPath = filepath.Join(appPath, "config.json")
		}
	}

	if !utl.IsExist(confPath) {
		err = errors.New("Can not load config at path:" + confPath)
		return
	}

	var fd *os.File
	fd, err = os.Open(confPath)
	if err != nil {
		return
	}
	defer fd.Close()
	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		return
	}

	newBuf := make([]byte, len(buf))
	isBegin := false //是否遇到了 //注释
	for i, ch := range buf {
		if isBegin {
			if ch == '\n' {
				isBegin = false
				newBuf[i] = '\n'
				continue
			}
			newBuf[i] = ' '
		} else {
			if ch == '/' && i+1 < len(buf) && buf[i+1] == '/' {
				isBegin = true
				newBuf[i] = ' '
				continue
			}
			newBuf[i] = ch
		}
	}

	var conf = Config{EnableGzip: true, EnableMinify: true, RunMode: "DEV"}
	err = json.Unmarshal(newBuf, &conf)
	if err != nil {
		return
	}
	_config = &conf
	c = _config

	loadSubConfig()

	checkConfig(c)

	b, _ := json.Marshal(c)
	fmt.Println(string(b))
	return
}

func GetConfig() *Config {
	if _config == nil {
		LoadConfig()
	}
	return _config
}

//将子配置应用到主配置上
func loadSubConfig() {
	var subConfig *SubConfig
	switch _config.RunMode {
	case "DEV":
		subConfig = _config.DevOpt
		_config.DevOpt = nil
	case "PROD":
		subConfig = _config.ProdOpt
		_config.ProdOpt = nil
	case "TEST":
		subConfig = _config.TestOpt
		_config.TestOpt = nil
	}

	if subConfig != nil {
		b, _ := json.Marshal(subConfig)
		if err := json.Unmarshal(b, _config); err != nil {
			fmt.Println(err)
		}
	}
}

//检测配置合法性，并设置默认配置
func checkConfig(conf *Config) {
	if conf.AppName == "" {
		conf.AppName = "Bigo"
	}

	if conf.WorkDir == "" {
		conf.WorkDir = workPath
	} else {
		if err := os.Chdir(conf.WorkDir); err != nil {
			panic(err)
		}
		workPath = conf.WorkDir
	}

	if conf.RunMode == "" {
		conf.RunMode = "DEV"
	}
	lms := map[string]string{"DEV": Dev, "TEST": Test, "PROD": Prod}
	SetEnv(lms[conf.RunMode])

	if conf.LogDir == "" && conf.RunMode == "PROD" {
		conf.LogDir = conf.WorkDir + "/log"
	}

	if conf.LogLevel == 0 {
		lm := map[string]int{"DEV": 1, "TEST": 1, "PROD": 3}

		conf.LogLevel = lm[conf.RunMode]
	}

	if conf.HttpPort == 0 {
		conf.HttpPort = 3000
	}

	if conf.EnableHttps {
		if conf.HttpsPort == 0 {
			conf.HttpsPort = 443
		}
		if conf.HttpsCertFile == "" || !utl.IsExist(conf.HttpsCertFile) {
			panic(fmt.Sprint("HttpsCertFile[%s]:file not find", conf.HttpsCertFile))
		}

		if conf.HttpsKeyFile == "" || !utl.IsExist(conf.HttpsKeyFile) {
			panic(fmt.Sprint("HttpsKeyFile[%s]:file not find", conf.HttpsKeyFile))
		}
	}

}
