package install

import (
	"github.com/wonderivan/logger"
	"golang.org/x/crypto/ssh"
	"os"
)

func (p  *PrinceInstaller) CheckCalid() {
	hosts := append(Masters, Nodes...)

	var session *ssh.Session
	var errors []error
	for _,h :=range hosts{
		session,err :=Connect(User,Password,PrivateKeyFile,h)
		if err!=nil{
			logger.Error("[%s]  ------------ check [ok]",h)
			logger.Error("[%s] ------------ error[%s]", h, err)
		}else{
			logger.Crit("[%s]  ------------ check [ok]",h)
			logger.Crit("[%s]  ------------ session[%p][ok]", h, session)
		}
	}
	defer func() {
		if session != nil {
			session.Close()
		}
	}()
	if len(errors) > 0 {
		logger.Error("has some linux server is connection ssh is failed")
		os.Exit(1)
	}
}