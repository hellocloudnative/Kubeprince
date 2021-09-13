package install

import (
	"github.com/wonderivan/logger"
	"fmt"
	"os"
	"strings"
	"sync"
)
type PrinceUpdate  struct {
	PrinceInstaller
	updateAll bool
}

// BuildUpdate update the build resources.
func BuildUpdate(updateNodes, updateMasters []string) {
	i := &PrinceUpdate{updateAll: false}
	masters := Masters
	nodes := Nodes
	//1. 升级masters
	if len(updateMasters) != 0 {
		if !UpdateForce { // false
			prompt := fmt.Sprintf("update command will update masters [%s], continue update (y/n)?", strings.Join(updateMasters, ","))
			result := Confirm(prompt)
			if !result {
				logger.Debug("update masters command is skip")
				goto node
			}
		}
		//只升级masters
		i.Masters = updateMasters
	}

	//2. 升级nodes
node:
	if len(updateNodes) != 0 {
		if !UpdateForce{ // flase
			prompt := fmt.Sprintf("update command will update nodes [%s], continue update (y/n)?", strings.Join(updateNodes, ","))
			result := Confirm(prompt)
			if !result {
				logger.Debug("update nodes command is skip")
				goto all
			}
		}
		//只升级nodes
		i.Nodes = updateNodes
	}
	//3. 升级所有节点
all:
	if len(updateNodes) == 0 && len(updateMasters) == 0 && UpdateAll {
		if !UpdateForce { // flase
			result := Confirm(`update command will update all masters and nodes, continue update (y/n)?`)
			if !result {
				logger.Debug("update all node command is skip")
				goto end
			}
		}
		// 所有master节点
		i.Masters = masters
		// 所有node节点
		i.Nodes = nodes
		i.updateAll = true
	}
end:
	if len(i.Masters) == 0 && len(i.Nodes) == 0 {
		logger.Warn("update nodes and masters is empty,please check.")
		os.Exit(-1)
	}
	i.CheckValid()
	i.Update()
	if i.updateAll {
		logger.Info("if update all  and  and check kubeprince  files ")
	}

}


//Update update cluster.
func (i *PrinceUpdate) Update() {
	var wg sync.WaitGroup
	if len(i.Nodes) > 0 {
		//1. 再升级nodes
		for _, node := range i.Nodes {
			wg.Add(1)
			go func(node string) {
				defer wg.Done()
				i.updateNode(node)
			}(node)
		}
		wg.Wait()
	}
	if len(i.Masters) > 0 {
		//2. 先升级master
		for _, master := range i.Masters {
			wg.Add(1)
			go func(master string) {
				defer wg.Done()
				i.updateMaster(master)
			}(master)
		}
		wg.Wait()
	}

}

func (i *PrinceUpdate) updateNode(node string) {
	update(node)
}

func (s *PrinceUpdate) updateMaster(master string) {
	 update(master)
}

func update(host string) {
	home, _ := os.UserHomeDir()
	cfgPath := home + defaultKubePath
	cmd := fmt.Sprintf("cd %s/shell && bash update.sh",cfgPath)
	_ = SSHConfig.CmdAsync(host, cmd)
	cmd = "echo  update  test "
	_ = SSHConfig.CmdAsync(host, cmd)


}

