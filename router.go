package hfw

//手动匹配路由
import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/hsyan2008/go-logger/logger"
)

type Router struct {
	C ControllerInterface
}

func (this *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	logger.Debug(r.Method, r.URL.Path, "start")
	defer logger.Debug(r.Method, r.URL.Path, "end")

	//把url补全为2段
	trimUrl := strings.Trim(strings.ToLower(r.URL.Path), "/")
	//如果url为/，切分后为1个空元素
	if trimUrl == "" {
		trimUrl = Config.Route.Controller
	}
	urls := strings.SplitN(trimUrl, "/", 3)

	if len(urls) == 0 {
		urls = append(urls, Config.Route.Controller)
	}
	if len(urls) == 1 {
		urls = append(urls, Config.Route.Action)
	}

	ctx := new(HttpContext)
	ctx.ResponseWriter = w
	ctx.Request = r
	ctx.Controll = urls[0]
	ctx.Action = urls[1]
	ctx.Path = fmt.Sprintf("%s/%s", urls[0], urls[1])
	ctx.TemplateFile = fmt.Sprintf("%s.html", ctx.Path)

	reflectVal := reflect.ValueOf(this.C)
	rt := reflectVal.Type()
	ct := reflect.Indirect(reflectVal).Type()

	//初始化Controller
	newInstance := reflect.New(ct) //因为并发，所以重新创建
	noneValue := []reflect.Value{}
	initValue := []reflect.Value{
		reflect.ValueOf(ctx),
	}

	//注意方法必须是大写开头，否则无法调用
	newInstance.MethodByName("Init").Call(initValue)
	defer newInstance.MethodByName("Finish").Call(noneValue)

	//第一个方法不生效，2、3数据无法传递
	// defer newInstance.MethodByName("ServerError").Call(noneValue)
	// defer this.C.ServerError() //C里没有初始化ResponseWriter，报nil指针
	// defer ctx.ServerError()
	defer func() {
		//注意recover只能执行一次
		if err := recover(); err != nil {
			//用户触发的
			if err == StopRunErr {
				return
			}
			buf := make([]byte, 1<<20)
			num := runtime.Stack(buf, false)
			logger.Warn(num, string(buf))

			errMsg := fmt.Sprint(err)
			logger.Warn(errMsg)
			newInstance.MethodByName("ServerError").Call(noneValue)
		}
	}()

	var action string
	//非法的Controller
	if strings.ToLower(ct.Name()) != urls[0] {
		action = "NotFound"
	} else {
		for i := 0; i < rt.NumMethod(); i++ {
			if strings.ToLower(rt.Method(i).Name) == urls[1] {
				action = rt.Method(i).Name
			}
		}
		//非法的Action
		if action == "" {
			action = "NotFound"
		}
	}

	newInstance.MethodByName("Before").Call(noneValue)
	newInstance.MethodByName(action).Call(noneValue)
	newInstance.MethodByName("After").Call(noneValue)
}

func RegisterRoute(pattern string, handler ControllerInterface) {
	s := "/" + strings.Trim(pattern, "/")
	if s == "/" {
		http.Handle(s, &Router{C: handler})
	} else {
		//如果没有这个，会重定向
		http.Handle(s, &Router{C: handler})
		//如果没有这个，会匹配到/
		http.Handle(s+"/", &Router{C: handler})
	}
}

func RegisterStatic(pattern string, dir string) {
	s := "/" + strings.Trim(pattern, "/")
	if s == "/" {
		http.Handle(s, http.FileServer(http.Dir(dir)))
	} else {
		//最后一定要加上/
		http.Handle(s+"/", http.FileServer(http.Dir(dir)))
	}
}

//favicon.ico
func RegisterFile(pattern string, dir string) {
	s := "/" + strings.Trim(pattern, "/")
	http.Handle(s, http.FileServer(http.Dir(dir)))
}
