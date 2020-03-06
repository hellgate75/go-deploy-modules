package service

import (
	"errors"
	"fmt"
	//	internal "github.com/hellgate75/go-deploy-modules/modules"
	"github.com/hellgate75/go-deploy/log"
	"github.com/hellgate75/go-deploy/modules/meta"
	"github.com/hellgate75/go-deploy/net/generic"
	"github.com/hellgate75/go-deploy/types/defaults"
	"github.com/hellgate75/go-deploy/types/module"
	"github.com/hellgate75/go-deploy/types/threads"
	"reflect"
	"strings"
	"time"
)

var Logger log.Logger = log.NewLogger(log.VerbosityLevelFromString(meta.GetVerbosity()))

var ERROR_TYPE reflect.Type = reflect.TypeOf(errors.New(""))

/*
* Service command structure
 */
type serviceCommand struct {
	Name         string
	State        string
	WithVars     []string
	WithList     []string
	host         defaults.HostValue
	session      module.Session
	config       defaults.ConfigPattern
	client       generic.NetworkClient
	start        time.Time
	lastDuration time.Duration
	uuid         string
	started      bool
	finished     bool
	paused       bool
	_running     bool
}

func (service *serviceCommand) SetClient(client generic.NetworkClient) {
	service.client = client
}

func (service *serviceCommand) Run() error {
	service.started = true
	service.start = time.Now()
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", err))
		}
		service._running = false
		service.finished = true
		service.paused = false
		service.started = false
	}()
	Logger.Warnf("Service command not implemented, service data: %s", service.String())
	service.started = false
	service.finished = true
	return err
}
func (service *serviceCommand) Stop() error {
	service._running = false
	return nil
}
func (service *serviceCommand) Kill() error {
	return nil
}
func (service *serviceCommand) Pause() error {
	if !service.paused && service.started {
		service.paused = true
		service.started = false
		service.lastDuration += time.Now().Sub(service.start)
		return nil
	}
	return errors.New("Process not running or already paused")
}
func (service *serviceCommand) Resume() error {
	if service.paused && !service.started {
		service.paused = false
		service.started = true
		service.start = time.Now()
	}
	return errors.New("Process running or not paused")
}
func (service *serviceCommand) IsRunning() bool {
	return service.started
}
func (service *serviceCommand) IsPaused() bool {
	return service.paused
}
func (service *serviceCommand) IsComplete() bool {
	return !service.started && !service.paused && service.finished
}
func (service *serviceCommand) UUID() string {
	return service.uuid
}
func (service *serviceCommand) Equals(r threads.StepRunnable) bool {
	if r != nil {
		return service.uuid == r.UUID()
	}
	return false
}
func (service *serviceCommand) UpTime() time.Duration {
	return time.Now().Sub(service.start) + service.lastDuration
}
func (service *serviceCommand) Clone() threads.StepRunnable {
	return &serviceCommand{
		Name:         service.Name,
		State:        service.State,
		WithVars:     service.WithVars,
		WithList:     service.WithList,
		host:         service.host,
		session:      service.session,
		config:       service.config,
		start:        time.Now(),
		lastDuration: 0 * time.Second,
		uuid:         module.NewSessionId(),
	}
}
func (service *serviceCommand) SetHost(host defaults.HostValue) {
	service.host = host
}
func (service *serviceCommand) SetSession(session module.Session) {
	service.session = session
}
func (service *serviceCommand) SetConfig(config defaults.ConfigPattern) {
	service.config = config
}

func (service serviceCommand) String() string {
	return fmt.Sprintf("serviceCommand {Name: %v, State: %v, WithVars: [%v], WithList: [%v]}", service.Name, service.State, service.WithVars, service.WithList)
}

func (service *serviceCommand) Convert(cmdValues interface{}) (threads.StepRunnable, error) {
	var superError error = nil
	defer func() {
		if r := recover(); r != nil {
			if ERROR_TYPE.AssignableTo(reflect.TypeOf(r)) {
				superError = r.(error)
			} else {
				superError = errors.New(fmt.Sprintf("%v", r))
			}
		}

	}()
	var name, state string
	var withVars []string = make([]string, 0)
	var withList []string = make([]string, 0)
	var valType string = fmt.Sprintf("%T", cmdValues)
	if len(valType) > 3 && "map" == valType[0:3] {
		for key, value := range cmdValues.(map[string]interface{}) {
			var elemValType string = fmt.Sprintf("%T", value)
			Logger.Debug(fmt.Sprintf("service.%s -> type: %s", strings.ToLower(key), elemValType))
			if strings.ToLower(key) == "name" {
				if elemValType == "string" {
					name = fmt.Sprintf("%v", value)
				} else {
					return nil, errors.New("Unable to parse command: service.name, with aguments of type " + elemValType + ", expected type string")
				}
			} else if strings.ToLower(key) == "state" {
				if elemValType == "string" {
					state = fmt.Sprintf("%v", value)
				} else {
					return nil, errors.New("Unable to parse command: service.state, with aguments of type " + elemValType + ", expected type string")
				}
			} else if strings.ToLower(key) == "withvars" {
				if elemValType == "[]string" {
					for _, val := range value.([]string) {
						withVars = append(withVars, val)
					}
				} else if elemValType == "[]interface {}" {
					for _, val := range value.([]interface{}) {
						withVars = append(withVars, fmt.Sprintf("%v", val))
					}
				} else {
					return nil, errors.New("Unable to parse command: service.withVars, with aguments of type " + elemValType + ", expected type []string")
				}
			} else if strings.ToLower(key) == "withlist" {
				if elemValType == "[]string" {
					for _, val := range value.([]string) {
						withList = append(withList, val)
					}
				} else if elemValType == "[]interface {}" {
					for _, val := range value.([]interface{}) {
						withList = append(withList, fmt.Sprintf("%v", val))
					}
				} else {
					return nil, errors.New("Unable to parse command: service.withList, with aguments of type " + elemValType + ", expected type []string")
				}
			} else {
				return nil, errors.New("Unknown command: service." + key)
			}
		}
	} else {
		return nil, errors.New("Unable to parse command: service, with aguments of type " + valType + ", expected type map[string]interfce{}")
	}
	if superError != nil {
		return nil, superError
	}
	return &serviceCommand{
		Name:         name,
		State:        state,
		WithVars:     withVars,
		WithList:     withList,
		start:        time.Now(),
		lastDuration: 0 * time.Second,
		uuid:         module.NewSessionId(),
	}, nil
}

var Converter meta.Converter = &serviceCommand{}

type stub struct{}

func (stub *stub) Discover(module string) (meta.Converter, error) {
	if module == "service" {
		return &serviceCommand{}, nil
	}
	return nil, errors.New("Wrong module")
}

func GetStub() meta.ProxyStub {
	return &stub{}
}

//func init() {
//	internal.RegisterModule("service", SeekModuleComponent)
//}

func main() {

}
