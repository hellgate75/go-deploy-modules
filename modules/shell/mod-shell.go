package shell

import (
	"errors"
	"fmt"
	//	internal "github.com/hellgate75/go-deploy-modules/modules"
	"bytes"
	"github.com/hellgate75/go-deploy/log"
	"github.com/hellgate75/go-deploy/modules/meta"
	"github.com/hellgate75/go-deploy/net/generic"
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
* Shell command structure
 */
type shellCommand struct {
	Exec         string
	RunAs        string
	AsRoot       bool
	WithVars     []string
	WithList     []string
	SaveState    string
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

func (shell *shellCommand) SetClient(client generic.NetworkClient) {
	shell.client = client
}

func (shell *shellCommand) Run() error {
	shell.started = true
	shell.start = time.Now()
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", err))
		}
		shell._running = false
		shell.finished = true
		shell.paused = false
		shell.started = false
	}()
	//	Logger.Warnf("Shell command not implemented, shell data: %s", shell.String())
	Logger.Warnf("Executing command: %s", shell.Exec)
	Logger.Debugf("Host labelled:  %s", shell.host.Name)
	buffer := bytes.NewBuffer([]byte{})
	var command string = shell.Exec
	if shell.WithVars != nil && len(shell.WithVars) > 0 {
		for _, varKey := range shell.WithVars {
			varValue, varValueErr := shell.session.GetVar(varKey)
			if varValueErr == nil {
				command = strings.ReplaceAll(command, "{{ "+varKey+" }}", varValue)
			}
		}
	}
	if shell.WithList != nil && len(shell.WithList) > 0 && strings.Index(command, "{{ item }}") >= 0 {
		var commandCopy string = command
		for _, listItem := range shell.WithList {
			strings.ReplaceAll(commandCopy, "{{ item }}", listItem)

			script := shell.client.Script(commandCopy)
			//	script.SetStdio(buffer, buffer)
			bytesArr, errCmd := script.ExecuteWithFullOutput()
			if errCmd != nil {
				return errors.New("Error Details: " + errCmd.Error() + ", StdErr: " + string(bytesArr) + ", Output: " + buffer.String())
			}
			buffer.Write(bytesArr)
		}
	} else {
		script := shell.client.Script(command)
		//	script.SetStdio(buffer, buffer)
		bytesArr, errCmd := script.ExecuteWithFullOutput()
		if errCmd != nil {
			return errCmd
		}
		buffer.Write(bytesArr)
	}

	if shell.SaveState != "" {
		shell.session.SetVar(shell.SaveState, buffer.String())
	}
	shell.started = false
	shell.finished = true
	return err
}
func (shell *shellCommand) Stop() error {
	shell._running = false
	return nil
}
func (shell *shellCommand) Kill() error {
	return nil
}
func (shell *shellCommand) Pause() error {
	if !shell.paused && shell.started {
		shell.paused = true
		shell.started = false
		shell.lastDuration += time.Now().Sub(shell.start)
		return nil
	}
	return errors.New("Process not running or already paused")
}
func (shell *shellCommand) Resume() error {
	if shell.paused && !shell.started {
		shell.paused = false
		shell.started = true
		shell.start = time.Now()
	}
	return errors.New("Process running or not paused")
}
func (shell *shellCommand) IsRunning() bool {
	return shell.started
}
func (shell *shellCommand) IsPaused() bool {
	return shell.paused
}
func (shell *shellCommand) IsComplete() bool {
	return !shell.started && !shell.paused && shell.finished
}
func (shell *shellCommand) UUID() string {
	return shell.uuid
}
func (shell *shellCommand) Equals(r threads.StepRunnable) bool {
	if r != nil {
		return shell.uuid == r.UUID()
	}
	return false
}

func (shell *shellCommand) UpTime() time.Duration {
	return time.Now().Sub(shell.start) + shell.lastDuration
}
func (shell *shellCommand) Clone() threads.StepRunnable {
	return &shellCommand{
		Exec:         shell.Exec,
		RunAs:        shell.RunAs,
		AsRoot:       shell.AsRoot,
		SaveState:    shell.SaveState,
		WithVars:     shell.WithVars,
		WithList:     shell.WithList,
		host:         shell.host,
		session:      shell.session,
		config:       shell.config,
		start:        time.Now(),
		lastDuration: 0 * time.Second,
		uuid:         module.NewSessionId(),
	}
}
func (shell *shellCommand) SetHost(host defaults.HostValue) {
	shell.host = host
}
func (shell *shellCommand) SetSession(session module.Session) {
	shell.session = session
}
func (shell *shellCommand) SetConfig(config defaults.ConfigPattern) {
	shell.config = config
}

func (shell shellCommand) String() string {
	return fmt.Sprintf("ShellCommand {Exec: %v, RunAs: %v, AsRoot: %v, WithVars: [%v], WithList: [%v]}", shell.Exec, shell.RunAs, strconv.FormatBool(shell.AsRoot), shell.WithVars, shell.WithList)
}

func (shell *shellCommand) Convert(cmdValues interface{}) (threads.StepRunnable, error) {
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
	var valType string = fmt.Sprintf("%T", cmdValues)
	var exec string = ""
	var runAs string = ""
	var asRoot bool = false
	var withVars []string = make([]string, 0)
	var withList []string = make([]string, 0)
	var asVar string = ""
	if len(valType) > 3 && "map" == valType[0:3] {
		for key, value := range cmdValues.(map[string]interface{}) {
			var elemValType string = fmt.Sprintf("%T", value)
			Logger.Debug(fmt.Sprintf("shell.%s -> type: %s", strings.ToLower(key), elemValType))
			if strings.ToLower(key) == "exec" {
				if elemValType == "string" {
					exec = fmt.Sprintf("%v", value)
				} else if elemValType == "[]string" {
					strings.Join(value.([]string), " ")
				} else {
					return nil, errors.New("Unable to parse command: shell.exec, with aguments of type " + elemValType + ", expected type string or []string")
				}
			} else if strings.ToLower(key) == "savestate" {
				if elemValType == "string" {
					asVar = fmt.Sprintf("%v", value)
				} else {
					return nil, errors.New("Unable to parse command: shell.saveState, with aguments of type " + elemValType + ", expected type string")
				}
			} else if strings.ToLower(key) == "runas" {
				if elemValType == "string" {
					runAs = fmt.Sprintf("%v", value)
				} else {
					return nil, errors.New("Unable to parse command: shell.runAs, with aguments of type " + elemValType + ", expected type string")
				}
			} else if strings.ToLower(key) == "asroot" {
				if elemValType == "string" {
					bl, err := strconv.ParseBool(fmt.Sprintf("%v", value))
					if err != nil {
						return nil, errors.New("Error parsing command: shell.asRoot, cause: " + err.Error())

					} else {
						asRoot = bl
					}

				} else if elemValType == "bool" {
					asRoot = value.(bool)
				} else {
					return nil, errors.New("Unable to parse command: shell.asRoot, with aguments of type " + elemValType + ", expected type bool or string")
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
					return nil, errors.New("Unable to parse command: shell.withVars, with aguments of type " + elemValType + ", expected type []string")
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
					return nil, errors.New("Unable to parse command: shell.withList, with aguments of type " + elemValType + ", expected type []string")
				}
			} else {
				return nil, errors.New("Unknown command: shell." + key)

			}
		}
	} else {
		return nil, errors.New("Unable to parse command: shell, with aguments of type " + valType + ", expected type map[string]interfce{}")
	}
	if exec == "" {
		return nil, errors.New("Missing command: shell.exec -> mandatory field")

	}
	if superError != nil {
		return nil, superError
	}
	return &shellCommand{
		Exec:         exec,
		RunAs:        runAs,
		AsRoot:       asRoot,
		WithVars:     withVars,
		SaveState:    asVar,
		start:        time.Now(),
		lastDuration: 0 * time.Second,
		uuid:         module.NewSessionId(),
	}, nil
}

var Converter meta.Converter = &shellCommand{}

type stub struct{}

func (stub *stub) Discover(module string) (meta.Converter, error) {
	if module == "shell" {
		return &shellCommand{}, nil
	}
	return nil, errors.New("Wrong module")
}

func GetStub() meta.ProxyStub {
	return &stub{}
}

func init() {
	//	internal.RegisterModule("shell", SeekModuleComponent)
}

func main() {

}
