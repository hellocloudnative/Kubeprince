package install

import (
	"k8s.io/client-go/kubernetes"
)
type KubeprinceUpgrade struct {
	PrinceConfig
	NewVersion   string
	NewPkgUrl    string
	IPtoHostName map[string]string
	Client       *kubernetes.Clientset
}
