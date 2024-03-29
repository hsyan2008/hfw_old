package hfw

//手动匹配路由
import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/hsyan2008/go-logger/logger"
)

// Init和Finish必定会执行，而且不允许被修改
// Before和After之间是业务逻辑，所有Before也是必定会执行
//用户手动StopRun()后，中止业务逻辑，跳过After，继续Finish
type ControllerInterface interface {
	Init(*HttpContext)
	Before()
	After()
	Finish()
	Redirect(string)
	Output()
	NotFound()
	ServerError()
	StopRun()
}

//渲染模板的数据放Data
//Json里的数据放Result
//Layout的功能未实现 TODO
type HttpContext struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Session        *Session
	Layout         string
	Controll       string
	Action         string
	Path           string
	TemplateFile   string
	isJson         bool
	isZip          bool
	Data           map[string]interface{}
	FuncMap        map[string]interface{}
	Result
}

//确认Controller实现了接口 ControllerInterface
var _ ControllerInterface = &Controller{}

var StopRunErr = errors.New("user stop run")

type Controller struct {
	HttpContext
}

func (this *Controller) Init(ctx *HttpContext) {
	// logger.Debug("Controller init")

	this.HttpContext = *ctx
	this.Data = make(map[string]interface{})
	this.FuncMap = make(map[string]interface{})

	this.Session = NewSession(this)

	if strings.Contains(this.Request.URL.RawQuery, "format=json") {
		this.isJson = true
	} else if strings.Contains(this.Request.Header.Get("Accept"), "application/json") {
		this.isJson = true
	}

	if strings.Contains(this.Request.Header.Get("Accept-Encoding"), "gzip") {
		this.isZip = true
	}
}

func (this *Controller) Before() {
	// logger.Debug("Controller Before")
}

func (this *Controller) After() {
	// logger.Debug("Controller After")
}

func (this *Controller) Finish() {
	// logger.Debug("Controller finish")

	// cookie := http.Cookie{Name: Config.Session.SessID, Value: this.Session.newid, Path: "/", HttpOnly: true}
	// http.SetCookie(this.ResponseWriter, &cookie)
	// this.Session.Rename()

	this.Output()
}

func (this *Controller) StopRun() {
	// logger.Debug("StopRun")
	panic(StopRunErr)

	//考虑用runtime.Goexit()，
	//经测试，会执行defer，但连接在这里就中断了，浏览器拿不到结果
}

func (this *Controller) Redirect(url string) {
	http.Redirect(this.ResponseWriter, this.Request, url, http.StatusFound)
	this.StopRun()
}

func (this *Controller) NotFound() {

	logger.Debug(this.Path, "NotFound")
	// this.ResponseWriter.WriteHeader(404) //firefox无法显示页面
	this.TemplateFile = "notfound.html"
	this.Data["title"] = "404错误"
	this.Action = "notfound"
	this.ErrNo = 99404
	this.ErrMsg = "NotFound"
}

//不要手动调用，用于捕获未知错误，手动请用Throw
//该方法不能使用StopRun，也不能panic，因为会被自动调用
func (this *Controller) ServerError() {

	// this.ResponseWriter.WriteHeader(500) //firefox无法显示页面
	this.TemplateFile = "internal.html"
	this.Data["title"] = "500错误"
	this.Action = "servererror"
	this.ErrNo = 99500
	this.ErrMsg = "ServerError"
}

func (this *Controller) Throw(code int64, msg string) {
	this.TemplateFile = "error.html"
	this.Data["title"] = "错误"
	this.Action = "throw"
	this.ErrNo = code
	this.ErrMsg = msg
	this.StopRun()
}

func (this *Controller) CheckErr(err error) {
	if nil != err {
		logger.Error(err)
		this.Throw(99500, "系统错误")
	}
}

func (this *Controller) Output() {
	// logger.Debug("Output")
	if this.ResponseWriter.Header().Get("Location") != "" {
		return
	}
	if this.TemplateFile == "" || this.isJson {
		this.RenderJson()
	} else {
		this.RenderFile()
	}
}

var templates = struct {
	list map[string]*template.Template
	l    *sync.RWMutex
}{
	list: make(map[string]*template.Template),
	l:    &sync.RWMutex{},
}

func (this *Controller) RenderFile() {

	this.ResponseWriter.Header().Set("Content-type", "text/html; charset=utf-8")

	var (
		t   *template.Template
		err error
		ok  bool
	)
	if Config.Template.Cache {
		templates.l.RLock()
		if t, ok = templates.list[this.TemplateFile]; !ok {
			templates.l.RUnlock()
			if len(this.FuncMap) == 0 {
				t = template.Must(template.ParseFiles(Config.Template.Html + this.TemplateFile))
			} else {
				t = template.Must(template.New(filepath.Base(this.TemplateFile)).Funcs(this.FuncMap).ParseFiles(Config.Template.Html + this.TemplateFile))
			}
			t = template.Must(t.ParseGlob(Config.Template.Html + "/widgets/*.html"))

			templates.l.Lock()
			templates.list[this.TemplateFile] = t
			templates.l.Unlock()
		} else {
			templates.l.RUnlock()
		}
	} else {
		if len(this.FuncMap) == 0 {
			t = template.Must(template.ParseFiles(Config.Template.Html + this.TemplateFile))
		} else {
			t = template.Must(template.New(filepath.Base(this.TemplateFile)).Funcs(this.FuncMap).ParseFiles(Config.Template.Html + this.TemplateFile))
		}
		t = template.Must(t.ParseGlob(Config.Template.Html + "/widgets/*.html"))
	}

	if this.isZip {
		this.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		writer := gzip.NewWriter(this.ResponseWriter)
		defer func() {
			_ = writer.Close()
		}()
		err = t.Execute(writer, this)
		if err != nil {
			logger.Warn(err)
		}
		// CheckErr(err)
	} else {
		err = t.Execute(this.ResponseWriter, this)
		if err != nil {
			logger.Warn(err)
		}
		// CheckErr(err)
	}

}

func (this *Controller) RenderJson() {

	this.ResponseWriter.Header().Set("Content-type", "application/json; charset=utf-8")

	b, err := json.Marshal(this.Result)
	CheckErr(err)

	if this.isZip {
		this.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		writer := gzip.NewWriter(this.ResponseWriter)
		defer func() {
			_ = writer.Close()
		}()
		_, err = writer.Write(b)
		CheckErr(err)
	} else {
		_, err = this.ResponseWriter.Write(b)
		CheckErr(err)
	}

}
