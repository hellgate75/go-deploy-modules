package service

import (
	"errors"
	"fmt"
	//	internal "github.com/hellgate75/go-deploy-modules/modules"
	"github.com/hellgate75/go-deploy/log"
	"github.com/hellgate75/go-deploy/modules/meta"
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
	Name     string
	State    string
	WithVars []string
	WithList []string
}

func (service *serviceCommand) Run() error {
	return nil
}
func (service *serviceCommand) Stop() error {
	return nil
}
func (service *serviceCommand) Kill() error {
	return nil
}
func (service *serviceCommand) Pause() error {
	return nil
}
func (service *serviceCommand) Resume() error {
	return nil
}
func (service *serviceCommand) IsRunning() bool {
	return false
}
func (service *serviceCommand) IsPaused() bool {
	return false
}
func (service *serviceCommand) IsComplete() bool {
	return false
}
func (service *serviceCommand) UUID() string {
	return ""
}
func (service *serviceCommand) UpTime() time.Duration {
	return time.Now().Sub(time.Now())
}
func (service *serviceCommand) Clone() threads.StepRunnable {
	return nil
}
func (service *serviceCommand) SetHost(host defaults.HostValue) {

}
func (service *serviceCommand) SetSession(session module.Session) {

}
func (service *serviceCommand) SetConfig(config defaults.ConfigPattern) {

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
		Name:     name,
		State:    state,
		WithVars: withVars,
		WithList: withList,
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
