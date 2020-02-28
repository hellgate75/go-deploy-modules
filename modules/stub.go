package modules

import (
	"fmt"
	cpmod "github.com/hellgate75/go-deploy-modules/modules/copy"
	semod "github.com/hellgate75/go-deploy-modules/modules/service"
	shmod "github.com/hellgate75/go-deploy-modules/modules/shell"
	"github.com/hellgate75/go-deploy/modules/meta"
)

func init() {

}

var modules map[string]meta.ProxyStub = make(map[string]meta.ProxyStub)

func RegisterModule(name string, stub meta.ProxyStub) {
	fmt.Println("New Module Registration: ", name)
	modules[name] = stub
}

func GetModulesMap() map[string]meta.ProxyStub {
	var modules map[string]meta.ProxyStub = make(map[string]meta.ProxyStub)
	modules["copy"] = cpmod.GetStub()
	modules["service"] = semod.GetStub()
	modules["shell"] = shmod.GetStub()
	return modules
}

func main() {

}
