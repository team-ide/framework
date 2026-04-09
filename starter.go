package framework

import (
	"context"
	"errors"
	"fmt"
	"github.com/team-ide/framework/util"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"sync"
	"syscall"
	"time"
)

const (

	// buildFlags go build -ldflags '-X system.buildFlags=--isServer' .
	buildFlags = ""
)

var (
	// releaseTime 发布时间 打包时候设置
	releaseTime = ""
	// releaseVersion 发布版本 打包时候设置
	releaseVersion = "0.0.1"
	// gitCommit 打包时候 git 信息
	gitCommit = ""
)

func GetReleaseVersion() string {
	return releaseVersion
}

func GetReleaseTime() string {
	return releaseTime
}

func GetGitCommit() string {
	return gitCommit
}

var (
	pwdDir = ""
)

func GetPwdDir() string {
	return pwdDir
}

func IsVersion() bool {
	for _, v := range os.Args {
		if v == "-version" || v == "-v" {
			return true
		}
	}
	return false
}

func init() {
	var err error

	pwdDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	pwdDir = util.FormatPath(pwdDir)
}

func OutInfo() {
	fmt.Println("Server Release version : " + GetReleaseVersion())
	fmt.Println("Server Release time : " + GetReleaseTime())
	fmt.Println("Server Git commit : " + GetGitCommit())
	fmt.Println("Server Go os : " + runtime.GOOS)
	fmt.Println("Server Go arch : " + runtime.GOARCH)
	fmt.Println("Server Go compiler : " + runtime.Compiler)
	fmt.Println("Server Go version : " + runtime.Version())
}

func NewStarter() (res *Starter) {
	res = &Starter{}
	res.eventListeners = make(map[Event]ListenerList)
	return
}

type Starter struct {
	initConfigFunc    DoFuncList
	initComponentFunc DoFuncList
	initFactoryFunc   DoFuncList
	initTableFunc     DoFuncList
	initDataFunc      DoFuncList

	serverStartFunc DoFuncList

	waitGroupForStop       *sync.WaitGroup
	waitGroupForStopLocker sync.Mutex

	startAt   time.Time
	startLock sync.Mutex

	eventListeners     map[Event]ListenerList
	eventListenersLock sync.Mutex

	//installList InstallList

	shouldWait bool

	isStopped bool
}

func (this_ *Starter) Start() (err error) {
	this_.startLock.Lock()
	defer this_.startLock.Unlock()

	if !this_.startAt.IsZero() {
		Debug("已启动", zap.Any("startAt", this_.startAt))
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New("崩溃异常:" + fmt.Sprint(e))
			Error("崩溃异常", zap.Error(err))
		}
		if err != nil {
			fmt.Println("启动失败:", err)
		} else {
			this_.startAt = time.Now()
		}
	}()

	go func() {
		OnSignal(this_.onSignalStop)
	}()

	err = this_.DoInit()
	if err != nil {
		Error("system do init error", zap.Error(err))
		return
	}

	err = this_.ServerStart()
	if err != nil {
		Error("system run server start error", zap.Error(err))
		return
	}

	this_.CallEvent(EventReady)
	return
}

func (this_ *Starter) DoInit() (err error) {
	Info("system run init start")

	this_.CallEvent(EventSystemInitBefore)

	err = this_.doInit("config", this_.initConfigFunc)
	if err != nil {
		return
	}

	err = this_.doInit("component", this_.initComponentFunc)
	if err != nil {
		return
	}

	err = this_.doInit("factory", this_.initFactoryFunc)
	if err != nil {
		return
	}

	err = this_.doInit("table", this_.initTableFunc)
	if err != nil {
		return
	}

	err = this_.doInit("data", this_.initDataFunc)
	if err != nil {
		return
	}

	this_.CallEvent(EventSystemInitAfter)
	Info("system run init success")

	return
}

//func (this_ *Starter) DoInstall() (err error) {
//	Info("system do install start")
//	this_.CallEvent(EventInstallBefore)
//
//	this_.installList.Sort()
//
//	for _, one := range this_.installFunc {
//		Info("system do install do:" + util.GetStringValue(one))
//		if one.fn == nil {
//			Warn("system do install func is null " + util.GetStringValue(one))
//			continue
//		}
//		err = one.fn()
//		if err != nil {
//			return
//		}
//	}
//	this_.CallEvent(EventInstallAfter)
//	Info("system do install success")
//	return
//}

func (this_ *Starter) ServerStart() (err error) {
	Info("system run server start start")
	this_.CallEvent(EventServerStartBefore)

	this_.serverStartFunc.Sort()

	for _, one := range this_.serverStartFunc {
		Info("system run server start do:" + util.GetStringValue(one))
		if one.fn == nil {
			Warn("system run server start func is null " + util.GetStringValue(one))
			continue
		}
		err = one.fn()
		if err != nil {
			return
		}
	}
	this_.CallEvent(EventServerStartAfter)
	Info("system run server start success")
	return
}

func (this_ *Starter) AddInitConfigFunc(name string, order int, fn func() error) {
	this_.initConfigFunc = append(this_.initConfigFunc, &DoFunc{Name: name, Order: order, fn: fn})
}

func (this_ *Starter) AddInitComponentFunc(name string, order int, fn func() error) {
	this_.initComponentFunc = append(this_.initComponentFunc, &DoFunc{Name: name, Order: order, fn: fn})
}

func (this_ *Starter) AddInitFactoryFunc(name string, order int, fn func() error) {
	this_.initFactoryFunc = append(this_.initFactoryFunc, &DoFunc{Name: name, Order: order, fn: fn})
}

func (this_ *Starter) AddInitTableFunc(name string, order int, fn func() error) {
	this_.initTableFunc = append(this_.initTableFunc, &DoFunc{Name: name, Order: order, fn: fn})
}

func (this_ *Starter) AddInitDataFunc(name string, order int, fn func() error) {
	this_.initDataFunc = append(this_.initDataFunc, &DoFunc{Name: name, Order: order, fn: fn})
}

func (this_ *Starter) AddServerStartFunc(name string, order int, fn func() error) {
	this_.serverStartFunc = append(this_.serverStartFunc, &DoFunc{Name: name, Order: order, fn: fn})
}

func (this_ *Starter) doInit(place string, funcList DoFuncList) (err error) {
	Info("system run " + place + " init start")

	if place == "config" {
		this_.CallEvent(EventConfigInitBefore)
	} else if place == "component" {
		this_.CallEvent(EventComponentInitBefore)
	} else if place == "factory" {
		this_.CallEvent(EventFactoryInitBefore)
	} else if place == "table" {
		this_.CallEvent(EventTableInitBefore)
	} else if place == "data" {
		this_.CallEvent(EventDataInitBefore)
	}

	funcList.Sort()

	for _, one := range funcList {
		Info("system run " + place + " init do:" + util.GetStringValue(one))
		if one.fn == nil {
			Warn("system run " + place + " func is null " + util.GetStringValue(one))
			continue
		}
		err = one.fn()
		if err != nil {
			return
		}
	}

	if place == "config" {
		this_.CallEvent(EventConfigInitAfter)
	} else if place == "component" {
		this_.CallEvent(EventComponentInitAfter)
	} else if place == "factory" {
		this_.CallEvent(EventFactoryInitAfter)
	} else if place == "table" {
		this_.CallEvent(EventTableInitAfter)
	} else if place == "data" {
		this_.CallEvent(EventDataInitAfter)
	}
	Info("system run " + place + " init success")

	return
}

type DoFunc struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
	fn    func() (err error)
}

type DoFuncList []*DoFunc

func (a DoFuncList) Len() int      { return len(a) }
func (a DoFuncList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less 根据 版本号数字 和 顺序号 排序 数值越大 越靠后
func (a DoFuncList) Less(i, j int) bool {
	return a[i].Order < a[j].Order
}
func (a DoFuncList) Sort() {
	sort.Sort(a)
}

type Event string

const (
	EventSystemInitBefore = Event("system-init-before-event")

	EventSystemInitAfter = Event("system-init-after-event")

	EventConfigInitBefore = Event("config-init-before-event")

	EventConfigInitAfter = Event("config-init-after-event")

	EventComponentInitBefore = Event("component-init-before-event")

	EventComponentInitAfter = Event("component-init-after-event")

	EventFactoryInitBefore = Event("factory-init-before-event")

	EventFactoryInitAfter = Event("factory-init-after-event")

	EventTableInitBefore = Event("table-init-before-event")

	EventTableInitAfter = Event("table-init-after-event")

	EventDataInitBefore = Event("data-init-before-event")

	EventDataInitAfter = Event("data-init-after-event")

	//EventInstallBefore = Event("install-before-event")
	//
	//EventInstallAfter = Event("install-after-event")

	EventServerStartBefore = Event("server-start-before-event")

	EventServerStartAfter = Event("server-start-after-event")

	EventReady = Event("ready-event")

	EventStopBefore = Event("stop-before-event")

	EventStop = Event("stop-event")
)

type EventContext interface {
	GetEvent() Event
	GetContext() context.Context
}

type Listener struct {
	event   Event
	order   int
	onEvent func(args ...any)
}

func (this_ *Starter) OnEvent(event Event, onEvent func(args ...any), order int) {
	this_.eventListenersLock.Lock()
	defer this_.eventListenersLock.Unlock()

	listener := &Listener{
		event:   event,
		onEvent: onEvent,
		order:   order,
	}
	if this_.eventListeners == nil {
		this_.eventListeners = make(map[Event]ListenerList)
	}
	this_.eventListeners[event] = append(this_.eventListeners[event], listener)
	this_.eventListeners[event].Sort()
}

func (this_ *Starter) GetListeners(event Event) (res []*Listener) {
	this_.eventListenersLock.Lock()
	defer this_.eventListenersLock.Unlock()

	if this_.eventListeners != nil {
		res = this_.eventListeners[event]
	}
	return
}

func (this_ *Starter) CallEvent(event Event, args ...any) {
	listeners := this_.GetListeners(event)
	if len(listeners) == 0 {
		return
	}
	for _, listener := range listeners {
		doEvent(listener, args...)
	}
	return
}

func doEvent(listener *Listener, args ...any) {
	if listener == nil || listener.onEvent == nil {
		return
	}
	defer func() {
		if e := recover(); e != nil {
			err := errors.New(fmt.Sprint(e))
			Error("listener event [" + string(listener.event) + "] order [" + fmt.Sprintf("%d", listener.order) + "] doEvent panic error:" + err.Error())
		}
	}()

	listener.onEvent(args...)

	return
}

type ListenerList []*Listener

func (p ListenerList) Len() int           { return len(p) }
func (p ListenerList) Less(i, j int) bool { return p[i].order < p[j].order }
func (p ListenerList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ListenerList) Sort()              { sort.Sort(p) }

func (this_ *Starter) onSignalStop() {
	this_.CallEvent(EventStopBefore)
	this_.CallEvent(EventStop)
}

//type Install struct {
//	Version       string       `json:"version"`
//	VersionNumber string       `json:"versionNumber"`
//	Name          string       `json:"name"`
//	Order         int          `json:"order"`
//	Do            func() error `json:"-"`
//	CheckExist    func() bool  `json:"-"`
//}
//
//func (this_ *Starter) AppendInstall(install *Install) {
//	this_.installList = append(this_.installList, install)
//}
//func (this_ *Starter) GetInstallList() InstallList {
//	return this_.installList
//}
//
//func (this_ *Starter) DoInstall1() (err error) {
//	defer func() {
//		if err != nil {
//			return
//		}
//		this_.CallEvent(EventInstallAfter)
//	}()
//	this_.CallEvent(EventInstallBefore)
//
//	list := this_.GetInstallList()
//	if len(list) == 0 {
//		Debug("install list is empty")
//		return
//	}
//
//	reg, err := regexp.Compile(`^v(\d+)\.(\d+)\.(\d+)$`)
//	if err != nil {
//		fmt.Println("regex compile error: ", err)
//		return
//	}
//	for _, one := range list {
//		version := one.Version
//		subMatches := reg.FindStringSubmatch(version)
//		if len(subMatches) > 0 {
//			one.VersionNumber = util.StrPadLeft(subMatches[1], 4, "0")
//			one.VersionNumber += util.StrPadLeft(subMatches[2], 4, "0")
//			one.VersionNumber += util.StrPadLeft(subMatches[3], 4, "0")
//		}
//	}
//	list.Sort()
//
//	for _, one := range list {
//		// 如果 实现了 检查已安装 则根据检测结果判断
//		if one.CheckExist != nil {
//			if one.CheckExist() {
//				Info("install version [" + one.Version + "] name [" + one.Name + "] installed")
//				continue
//			}
//		} else {
//			Warn("install version [" + one.Version + "] name [" + one.Name + "] not check installed")
//		}
//		err = this_.install(one)
//		if err != nil {
//			return err
//		}
//	}
//	return
//}
//
//func (this_ *Starter) install(in *Install) (err error) {
//	Info("install version [" + in.Version + "] name [" + in.Name + "] start")
//	if in.Do == nil {
//		Info("install version [" + in.Version + "] name [" + in.Name + "] do is null")
//		return
//	}
//	err = in.Do()
//	if err != nil {
//		Info("install version [" + in.Version + "] name [" + in.Name + "] error:" + err.Error())
//		return
//	}
//
//	Info("install version [" + in.Version + "] name [" + in.Name + "] success")
//	return
//}
//
//type InstallList []*Install
//
//func (a InstallList) Len() int      { return len(a) }
//func (a InstallList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
//
//// Less 根据 版本号数字 和 顺序号 排序 数值越大 越靠后
//func (a InstallList) Less(i, j int) bool {
//	if a[i].VersionNumber == a[j].VersionNumber {
//		return a[i].Order < a[j].Order
//	}
//	return a[i].VersionNumber < a[j].VersionNumber
//}
//func (a InstallList) Sort() { sort.Sort(a) }

func (this_ *Starter) SetShouldWait(v bool) {
	this_.shouldWait = v
}
func (this_ *Starter) ShouldWait() bool {
	return this_.shouldWait
}

func (this_ *Starter) Wait() {
	this_.waitGroupForStopLocker.Lock()
	if this_.waitGroupForStop == nil {
		this_.waitGroupForStop = &sync.WaitGroup{}
		this_.waitGroupForStop.Add(1)
	}
	this_.waitGroupForStopLocker.Unlock()
	this_.waitGroupForStop.Wait()
}

func (this_ *Starter) CallStop() {
	this_.waitGroupForStopLocker.Lock()
	if this_.waitGroupForStop != nil {
		this_.waitGroupForStop.Done()
		this_.waitGroupForStop = nil
	}
	this_.waitGroupForStopLocker.Unlock()
	os.Exit(0)
}

var (
	onSignalFuncList []func()
	onSignalEnd      func()
	onSignalLock     sync.Mutex
)

func OnSignal(fn func()) {
	onSignalLock.Lock()
	defer onSignalLock.Unlock()

	onSignalFuncList = append(onSignalFuncList, fn)

	if onSignalEnd != nil {
		return
	}
	onSignalEnd = func() {
		defer func() {
			os.Exit(0)
		}()
		for _, one := range onSignalFuncList {
			one()
		}
	}

	c := make(chan os.Signal)
	signal.Notify(c)

	//SIGHUP	1	Term	终端控制进程结束(终端连接断开)
	//SIGINT	2	Term	用户发送INTR字符(Ctrl+C)触发
	//SIGQUIT	3	Core	用户发送QUIT字符(Ctrl+/)触发
	//SIGILL	4	Core	非法指令(程序错误、试图执行数据段、栈溢出等)
	//SIGABRT	6	Core	调用abort函数触发
	//SIGFPE	8	Core	算术运行错误(浮点运算错误、除数为零等)
	//SIGKILL	9	Term	无条件结束程序(不能被捕获、阻塞或忽略)
	//SIGSEGV	11	Core	无效内存引用(试图访问不属于自己的内存空间、对只读内存空间进行写操作)
	//SIGPIPE	13	Term	消息管道损坏(FIFO/Socket通信时，管道未打开而进行写操作)
	//SIGALRM	14	Term	时钟定时信号
	//SIGTERM	15	Term	结束程序(可以被捕获、阻塞或忽略)
	//SIGUSR1	30,10,16	Term	用户保留
	//SIGUSR2	31,12,17	Term	用户保留
	//SIGCHLD	20,17,18	Ign	子进程结束(由父进程接收)
	//SIGCONT	19,18,25	Cont	继续执行已经停止的进程(不能被阻塞)
	//SIGSTOP	17,19,23	Stop	停止进程(不能被捕获、阻塞或忽略)
	//SIGTSTP	18,20,24	Stop	停止进程(可以被捕获、阻塞或忽略)
	//SIGTTIN	21,21,26	Stop	后台程序从终端中读取数据时触发
	//SIGTTOU	22,22,27	Stop	后台程序向终端中写数据时触发

	for s := range c {
		switch s {
		case os.Kill: // kill -9 pid，下面的无效
			Warn("强制退出", zap.Any("signal", s.String()))
			fmt.Println("强制退出", s.String())
			onSignalEnd()
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT: // ctrl + c
			Warn("退出", zap.Any("signal", s.String()))
			fmt.Println("退出", s.String())
			onSignalEnd()
		}
	}
}
