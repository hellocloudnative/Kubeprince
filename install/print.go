package install

import (
	"encoding/json"
	"github.com/wonderivan/logger"
	"strings"
)

const (success  string = "🎉🎉🎉")
func(p *PrinceInstaller)Print(process ...string){
	if len(process) == 0 {
		configJson,_:= json.Marshal(p)
		logger.Info("[globals]kubeprince config is:",string(configJson))
	}else {
		var sb  strings.Builder
		for _,v :=range process{
			sb.Write([]byte(v))
		}
		logger.Debug(sb.String())
	}
}
func (p *PrinceInstaller) PrintFinish() {
	logger.Info("kubeprince install successful. ",success)
}