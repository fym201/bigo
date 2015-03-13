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

type LogLevel int

const (
	LogLevelNone LogLevel = iota
	LogLevelInfo
	LogLevelDebug
	LogLevelError
)

type SubConfig map[string]interface{}

//静态文件配置
type StaticOpt struct {
	Path        string `json:"Path"`        //静态目录的路径
	Prefix      string `json:"Prefix"`      //此目录在url中的前缀，如果不设置则为/
	SkipLogging bool   `json:"SkipLogging"` //默认在DEV,TEST模式下为false,在PROD模式下为true. 是否跳过log输出
	IndexFile   string `json:"IndexFile"`   //如果指定，则以此文件为索引文件
}

//本地化配置
type I18nOpt struct {
	Enable          bool     `json:"Enable"`          //是否开启
	SubUrl          string   `json:"SubUrl"`          //子目录，默认为空
	Directory       string   `json:"Directory"`       //存放本地化文件的目录，默认为 "conf/locale"
	CustomDirectory string   `json:"CustomDirectory"` //用于重载的本地化文件目录,默认为 "custom/conf/locale"
	Langs           []string `json:"Langs"`           //支持的语言，顺序是有意义的
	Names           []string `json:"Names"`           //语言的本地化名称,和上面一一对应
	Format          string   `json:"Format"`          //本地化文件命名风格，默认为 "locale_%s.ini"
	Parameter       string   `json:"Parameter"`       //指示当前语言的 URL 参数名，默认为 "lang"
	Redirect        bool     `json:"Redirect"`        //当通过 URL 参数指定语言时是否重定向，默认为 false
	TmplName        string   `json:"TmplName"`        //存放在模板中的本地化对象变量名称，默认为 "i18n"
}

//模板引擎配置
type TmplOpt struct {
	Enable          bool     `json:"Enable"`          //是否启动模板引擎，默认为false
	Directory       string   `json:"Directory"`       //模板文件目录，默认为 "views"
	Extensions      []string `json:"Extensions"`      //模板文件后缀，默认为 [".tmpl", ".html"]
	Delims          []string `json:"Delims"`          //模板语法分隔符，默认为 ["{{", "}}"]
	Charset         string   `json:"Charset"`         //追加的 Content-Type 头信息，默认为 "UTF-8"
	IndentJSON      bool     `json:"IndentJSON"`      //渲染具有缩进格式的 JSON，默认为不缩进
	IndentXML       bool     `json:"IndentXML"`       //渲染具有缩进格式的 XML，默认为不缩进
	HTMLContentType string   `json:"HTMLContentType"` //默认为 "text/html"
}

//如果以【app -c configPath】的命令行形式指定了文件，那么直接加载这个文件，否则
//查找当前工作目录下的config.json
//查找app所在目录下的config.json
type Config struct {
	AppName  string   `json:"AppName"`  //应用名称
	WorkDir  string   `json:"WorkDir"`  //工作目录,默认为运行目录
	LogDir   string   `json:"LogDir"`   //日志目录,默认为$WORKDIR/log
	LogLevel LogLevel `json:"LogLevel"` //日志等级,0.none 1.info 2.debug 3.error ,默认在DEV,TEST下为1,PROD下为3

	HttpAddr string `json:"HttpAddr"` //http监听地址,默认在0.0.0.0
	HttpPort int    `json:"HttpPort"` //http监听端口,默认为3000

	HttpsAddr     string `json:"HttpsAddr"`     //https监听地址,默认在0.0.0.0
	HttpsPort     int    `json:"HttpsPort"`     //https监听地址,默认为443
	HttpsCertFile string `json:"HttpsCertFile"` //https证书文件路径
	HttpsKeyFile  string `json:"HttpsKeyFile"`  //https证书key文件路径
	EnableHttps   bool   `json:"EnableHttps"`   //是否开启https服务，默认false
	ForceHttps    bool   `json:"ForceHttps"`    //是否强制将http请求转为https,默认为false

	EnableGzip bool `json:"EnableGzip"` //是否开启gzip,默认true,如果开启,并且客户端接受gzip的话责对会话进行gzip压缩
	ForceGzip  bool `json:"ForceGzip"`  //是否强制开启gzip,默认false,如果开启,不管客户端接不接受,都会以gzip进行传输

	EnableMinify           bool        `json:"EnableMinify"`           //默认为true,是否对.html .js .css 进行最小化处理
	StaticExtensionsToGzip []string    `json:"StaticExtensionsToGzip"` //使用gzip进行压缩传输的静态文件后缀,受客户端协议或FORCE_GZIP影响
	Statics                []StaticOpt `json:"Statics"`                //静态目录,数组
	I18n                   *I18nOpt    `json:"i18n"`                   //本地化配置
	Tmpl                   *TmplOpt    `json:"Tmpl"`                   //模板引擎配置
	RunMode                string      `json:"RunMode"`                //运行模式，DEV为开发模式，PROD为发布模式,TEST为测试模式
	DevOpt                 *SubConfig  `json:"DEV"`                    //对于DEV模式下的配置，RUN_MODE为DEV时会覆盖顶级配置
	TestOpt                *SubConfig  `json:"TEST"`                   //对于TEST模式下的配置，RUN_MODE为TEST时会覆盖顶级配置
	ProdOpt                *SubConfig  `json:"PROD"`                   //对于PROD模式下的配置，RUN_MODE为PROD时会覆盖顶级配置

	CutomOpt map[string]interface{} `json:"Cutom"` //自定义选项
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
	isString := false
	for i, ch := range buf {

		if isBegin {
			if ch == '\n' {
				isBegin = false
				newBuf[i] = '\n'
				continue
			}
			newBuf[i] = ' '
		} else {
			if ch == '"' {
				isString = !isString
			}

			if !isString && ch == '/' && i+1 < len(buf) && buf[i+1] == '/' {
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

	//b, _ := json.Marshal(c)
	//fmt.Println(string(b))
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
		lm := map[string]LogLevel{"DEV": LogLevelInfo, "TEST": LogLevelInfo, "PROD": LogLevelError}

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

	if conf.I18n != nil && conf.I18n.Enable {
		if conf.I18n.Directory == "" {
			conf.I18n.Directory = "conf/locale"
		}

		if conf.I18n.CustomDirectory == "" {
			conf.I18n.CustomDirectory = "custom/conf/locale"
		}

		if conf.I18n.Format == "" {
			conf.I18n.Format = "locale_%s.ini"
		}

		if conf.I18n.TmplName == "" {
			conf.I18n.TmplName = "i18n"
		}

	}

	if conf.Tmpl != nil {
		if conf.Tmpl.Directory == "" {
			conf.Tmpl.Directory = "views"
		}

		if len(conf.Tmpl.Extensions) == 0 {
			conf.Tmpl.Extensions = []string{".tmpl", ".html"}
		}

		if len(conf.Tmpl.Delims) == 0 {
			conf.Tmpl.Delims = []string{"{{", "}}"}
		}

		if conf.Tmpl.Charset == "" {
			conf.Tmpl.Charset = "UTF-8"
		}

		if conf.Tmpl.HTMLContentType == "" {
			conf.Tmpl.HTMLContentType = "text/html"
		}
	}

}

func (c *Config) Custom(key string) interface{} {
	return c.CutomOpt[key]
}

func (c *Config) CustomString(key string) string {
	return utl.ToStr(c.CutomOpt[key])
}

func (c *Config) CustomInt(key string) int {
	return utl.Str(utl.ToStr(c.CutomOpt[key])).MustInt()
}

func (c *Config) CustomInt64(key string) int64 {
	return utl.Str(utl.ToStr(c.CutomOpt[key])).MustInt64()
}

func (c *Config) CustomFloat32(key string) float32 {
	return utl.Str(utl.ToStr(c.CutomOpt[key])).MustFloat32()
}

func (c *Config) CustomFloat64(key string) float64 {
	return utl.Str(utl.ToStr(c.CutomOpt[key])).MustFloat64()
}
