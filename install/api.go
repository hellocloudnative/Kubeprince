package install
import "github.com/wonderivan/logger"

func setKubeadmAPI(version string) {
	major, _ := GetMajorMinorInt(version)
	switch {
	case major < 120:
		KubeadmAPI = KubeadmV1beta1
	case major < 123 && major >= 120:
		KubeadmAPI = KubeadmV1beta2
	case major >= 123:
		KubeadmAPI = KubeadmV1beta3
	default:
		KubeadmAPI = KubeadmV1beta3
	}
	logger.Debug("KubeadmApi: %s", KubeadmAPI)
}
