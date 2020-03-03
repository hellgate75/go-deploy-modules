package copy

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
	"strconv"
	"strings"
	"time"
)

var Logger log.Logger = log.NewLogger(log.VerbosityLevelFromString(meta.GetVerbosity()))

var ERROR_TYPE reflect.Type = reflect.TypeOf(errors.New(""))

/*
* Service command structure
 */
type copyCommand struct {
	SourceDir      string
	DestinationDir string
	CreateDest     bool
	WithVars       []string
	WithList       []string
}

func (copyCmd *copyCommand) Run() error {
	return nil
}
func (copyCmd *copyCommand) Stop() error {
	return nil
}
func (copyCmd *copyCommand) Kill() error {
	return nil
}
func (copyCmd *copyCommand) Pause() error {
	return nil
}
func (copyCmd *copyCommand) Resume() error {
	return nil
}
func (copyCmd *copyCommand) IsRunning() bool {
	return false
}
func (copyCmd *copyCommand) IsPaused() bool {
	return false
}
func (copyCmd *copyCommand) IsComplete() bool {
	return false
}
func (copyCmd *copyCommand) UUID() string {
	return ""
}
func (copyCmd *copyCommand) UpTime() time.Duration {
	return time.Now().Sub(time.Now())
}
func (copyCmd *copyCommand) Clone() threads.StepRunnable {
	return nil
}
func (copyCmd *copyCommand) SetHost(host defaults.HostValue) {

}
func (copyCmd *copyCommand) SetSession(session module.Session) {

}
func (copyCmd *copyCommand) SetConfig(config defaults.ConfigPattern) {

}

func (copyCmd copyCommand) String() string {
	return fmt.Sprintf("ServiceCommand {SourceDir: %v, DestDir: %v, CreateDest: %v, WithVars: [%v], WithList: [%v]}", copyCmd.SourceDir, copyCmd.DestinationDir, strconv.FormatBool(copyCmd.CreateDest), copyCmd.WithVars, copyCmd.WithList)
}

func (copyCmd *copyCommand) Convert(cmdValues interface{}) (threads.StepRunnable, error) {
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
	var sourceDir, destDir string
	var withVars []string = make([]string, 0)
	var withList []string = make([]string, 0)
	var createDest bool = false
	var valType string = fmt.Sprintf("%T", cmdValues)
	if len(valType) > 3 && "map" == valType[0:3] {
		for key, value := range cmdValues.(map[string]interface{}) {
			var elemValType string = fmt.Sprintf("%T", value)
			if strings.ToLower(key) == "srcdir" {
				if elemValType == "string" {
					sourceDir = fmt.Sprintf("%v", value)
				} else {
					return nil, errors.New("Unable to parse command: service.srcDir, with aguments of type " + elemValType + ", expected type string")
				}
			} else if strings.ToLower(key) == "destdir" {
				if elemValType == "string" {
					destDir = fmt.Sprintf("%v", value)
				} else {
					return nil, errors.New("Unable to parse command: service.destDir, with aguments of type " + elemValType + ", expected type string")
				}
			} else if strings.ToLower(key) == "createifmissing" {
				if elemValType == "string" {
					bl, err := strconv.ParseBool(fmt.Sprintf("%v", value))
					if err != nil {
						return nil, errors.New("Error parsing command: shell.createIfMissing, cause: " + err.Error())

					} else {
						createDest = bl
					}

				} else if elemValType == "bool" {
					createDest = value.(bool)
				} else {
					return nil, errors.New("Unable to parse command: shell.createIfMissing, with aguments of type " + elemValType + ", expected type bool or string")
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
	return &copyCommand{
		SourceDir:      sourceDir,
		DestinationDir: destDir,
		CreateDest:     createDest,
		WithVars:       withVars,
		WithList:       withList,
	}, nil
}

var Converter meta.Converter = &copyCommand{}

type stub struct{}

func (stub *stub) Discover(module string) (meta.Converter, error) {
	if module == "copy" {
		return &copyCommand{}, nil
	}
	return nil, errors.New("Wrong module")
}

func GetStub() meta.ProxyStub {
	return &stub{}
}

//func init() {
//	internal.RegisterModule("copy", SeekModuleComponent)
//}

func main() {

}
