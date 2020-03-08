package copy

import (
	"errors"
	"fmt"
	"os"
	"github.com/hellgate75/go-tcp-common/log"
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
* Service command structure
 */
type copyCommand struct {
	SourceDir      string
	DestinationDir string
	FilePerm     	os.FileMode
	CreateDest     bool
	WithVars       []string
	WithList       []string
	host           defaults.HostValue
	session        module.Session
	config         defaults.ConfigPattern
	client         generic.NetworkClient
	start          time.Time
	lastDuration   time.Duration
	uuid           string
	started        bool
	finished       bool
	paused         bool
	_running       bool
}

func (copyCmd *copyCommand) SetClient(client generic.NetworkClient) {
	copyCmd.client = client
}

func (copyCmd *copyCommand) Run() error {
	copyCmd.started = true
	copyCmd.start = time.Now()
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", err))
		}
		copyCmd._running = false
		copyCmd.finished = true
		copyCmd.paused = false
		copyCmd.started = false
	}()
	//Logger.Warnf("Copy Command command not implemented, copy command data: %s", copyCmd.String())
	var sourceDir string = copyCmd.SourceDir
	var destinationDir string = copyCmd.DestinationDir
	var createDestination bool = copyCmd.CreateDest

	transfer := copyCmd.client.FileTranfer()

	if copyCmd.WithList != nil && len(copyCmd.WithList) > 0 {
		for _, listItem := range copyCmd.WithList {
			if strings.Index(sourceDir, "{{ item }}") < 0 {
				if strings.Index(destinationDir, "{{ item }}") < 0 {
					err = errors.New("Neither Source nor Destination folder contain scalable variable '{{ item }}'")
					break
				}
			}
			sourceDirCopy := strings.ReplaceAll(sourceDir, "{{ item }}", listItem)
			destinationDirCopy := strings.ReplaceAll(destinationDir, "{{ item }}", listItem)
			if copyCmd.WithVars != nil && len(copyCmd.WithVars) > 0 {
				for _, varKey := range copyCmd.WithVars {
					varValue, varValueErr := copyCmd.session.GetVar(varKey)
					if varValueErr == nil {
						sourceDirCopy = strings.ReplaceAll(sourceDirCopy, "{{ "+varKey+" }}", varValue)
						destinationDirCopy = strings.ReplaceAll(destinationDirCopy, "{{ "+varKey+" }}", varValue)
					}
					Logger.Debugf("List Item: %s", listItem)
					Logger.Debugf("Source Folder: %s", sourceDirCopy)
					Logger.Debugf("Destination Folder: %s", destinationDirCopy)
					Logger.Debugf("Create Destination Folder: %v", createDestination)
					errX := copySourceToDest(copyCmd, transfer, sourceDirCopy, destinationDirCopy, createDestination)
					if errX != nil {
						err = errX
						break
					}
				}
			}
		}
	} else {
		if copyCmd.WithVars != nil && len(copyCmd.WithVars) > 0 {
			for _, varKey := range copyCmd.WithVars {
				varValue, varValueErr := copyCmd.session.GetVar(varKey)
				if varValueErr == nil {
					sourceDir = strings.ReplaceAll(sourceDir, "{{ "+varKey+" }}", varValue)
					destinationDir = strings.ReplaceAll(destinationDir, "{{ "+varKey+" }}", varValue)
				}
			}
		}
		Logger.Debugf("Source Folder: %s", sourceDir)
		Logger.Debugf("Destination Folder: %s", destinationDir)
		Logger.Debugf("Create Destination Folder: %v", createDestination)
		err = copySourceToDest(copyCmd, transfer, sourceDir, destinationDir, createDestination)

	}
	copyCmd.started = false
	copyCmd.finished = true
	return err
}

func copySourceToDest(copyCmd *copyCommand, transfer generic.FileTransfer, src string, dest string, create bool) error {
	fi, err := os.Stat(src)
	if err != nil {
		return errors.New("Source file/folder doesn't exists...")
	}
	if fi.IsDir() {
		//Folder
		err = transfer.TransferFolderAs(src, dest, copyCmd.FilePerm)
		if err != nil {
			return err
		}
	} else {
		//File
		err = transfer.TransferFileAs(src, dest, copyCmd.FilePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (copyCmd *copyCommand) Stop() error {
	copyCmd._running = false
	return nil
}
func (copyCmd *copyCommand) Kill() error {
	return nil
}
func (copyCmd *copyCommand) Pause() error {
	if !copyCmd.paused && copyCmd.started {
		copyCmd.paused = true
		copyCmd.started = false
		copyCmd.lastDuration += time.Now().Sub(copyCmd.start)
		return nil
	}
	return errors.New("Process not running or already paused")
}
func (copyCmd *copyCommand) Resume() error {
	if copyCmd.paused && !copyCmd.started {
		copyCmd.paused = false
		copyCmd.started = true
		copyCmd.start = time.Now()
	}
	return errors.New("Process running or not paused")
}
func (copyCmd *copyCommand) IsRunning() bool {
	return copyCmd.started
}
func (copyCmd *copyCommand) IsPaused() bool {
	return copyCmd.paused
}
func (copyCmd *copyCommand) IsComplete() bool {
	return !copyCmd.started && !copyCmd.paused && copyCmd.finished
}
func (copyCmd *copyCommand) UUID() string {
	return copyCmd.uuid
}
func (copyCmd *copyCommand) Equals(r threads.StepRunnable) bool {
	if r != nil {
		return copyCmd.uuid == r.UUID()
	}
	return false
}
func (copyCmd *copyCommand) UpTime() time.Duration {
	return time.Now().Sub(copyCmd.start) + copyCmd.lastDuration
}
func (copyCmd *copyCommand) Clone() threads.StepRunnable {
	return &copyCommand{
		SourceDir:      copyCmd.SourceDir,
		DestinationDir: copyCmd.DestinationDir,
		FilePerm:     copyCmd.FilePerm,
		CreateDest:     copyCmd.CreateDest,
		WithVars:       copyCmd.WithVars,
		WithList:       copyCmd.WithList,
		host:           copyCmd.host,
		session:        copyCmd.session,
		config:         copyCmd.config,
		start:          time.Now(),
		lastDuration:   0 * time.Second,
		uuid:           module.NewSessionId(),
	}
}
func (copyCmd *copyCommand) SetHost(host defaults.HostValue) {
	copyCmd.host = host
}
func (copyCmd *copyCommand) SetSession(session module.Session) {
	copyCmd.session = session

}
func (copyCmd *copyCommand) SetConfig(config defaults.ConfigPattern) {
	copyCmd.config = config

}

func (copyCmd copyCommand) String() string {
	return fmt.Sprintf("ServiceCommand {SourceDir: %v, DestDir: %v, CreateDest: %v, FilePerm: %s, WithVars: [%v], WithList: [%v]}", copyCmd.SourceDir, copyCmd.DestinationDir, copyCmd.FilePerm.String(), strconv.FormatBool(copyCmd.CreateDest), copyCmd.WithVars, copyCmd.WithList)
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
	var filePerm os.FileMode = 0664
	var valType string = fmt.Sprintf("%T", cmdValues)
	if len(valType) > 3 && "map" == valType[0:3] {
		for key, value := range cmdValues.(map[string]interface{}) {
			var elemValType string = fmt.Sprintf("%T", value)
			Logger.Debug(fmt.Sprintf("copy.%s -> type: %s", strings.ToLower(key), elemValType))
			if strings.ToLower(key) == "srcdir" {
				if elemValType == "string" {
					sourceDir = fmt.Sprintf("%v", value)
				} else {
					return nil, errors.New("Unable to parse command: copy.srcDir, with aguments of type " + elemValType + ", expected type string")
				}
			} else if strings.ToLower(key) == "destdir" {
				if elemValType == "string" {
					destDir = fmt.Sprintf("%v", value)
				} else {
					return nil, errors.New("Unable to parse command: copy.destDir, with aguments of type " + elemValType + ", expected type string")
				}
			} else if strings.ToLower(key) == "perm" {
				if elemValType == "string" {
					permEx := fmt.Sprintf("%v", value)
					perm, err4b := strconv.Atoi(permEx)
					if err4b == nil {
						filePerm = os.FileMode(perm)
					}
				} else {
					return nil, errors.New("Unable to parse command: copy.perm, with aguments of type " + elemValType + ", expected type string")
				}
			} else if strings.ToLower(key) == "createifmissing" {
				if elemValType == "string" {
					bl, err := strconv.ParseBool(fmt.Sprintf("%v", value))
					if err != nil {
						return nil, errors.New("Error parsing command: copy.createIfMissing, cause: " + err.Error())

					} else {
						createDest = bl
					}

				} else if elemValType == "bool" {
					createDest = value.(bool)
				} else {
					return nil, errors.New("Unable to parse command: copy.createIfMissing, with aguments of type " + elemValType + ", expected type bool or string")
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
					return nil, errors.New("Unable to parse command: copy.withVars, with aguments of type " + elemValType + ", expected type []string")
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
					return nil, errors.New("Unable to parse command: copy.withList, with aguments of type " + elemValType + ", expected type []string")
				}
			} else {
				return nil, errors.New("Unknown command: copy." + key)
			}
		}
	} else {
		return nil, errors.New("Unable to parse command: copy, with aguments of type " + valType + ", expected type map[string]interfce{}")
	}
	if superError != nil {
		return nil, superError
	}
	return &copyCommand{
		SourceDir:      sourceDir,
		DestinationDir: destDir,
		CreateDest:     createDest,
		FilePerm: 		filePerm,
		WithVars:       withVars,
		WithList:       withList,
		start:          time.Now(),
		lastDuration:   0 * time.Second,
		uuid:           module.NewSessionId(),
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
