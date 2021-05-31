module arhat.dev/arhat

go 1.16

require (
	arhat.dev/aranya-proto v0.3.5
	arhat.dev/arhat-proto v0.4.4
	arhat.dev/libext v0.5.2
	arhat.dev/pkg v0.5.5
	ext.arhat.dev/runtimeutil v0.3.1
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gogo/protobuf v1.3.2
	github.com/goiiot/libmqtt v0.9.6
	github.com/klauspost/compress v1.12.3
	github.com/mholt/archiver/v3 v3.5.0
	github.com/mssola/user_agent v0.5.3
	github.com/nats-io/jwt/v2 v2.0.2
	github.com/nats-io/nats-streaming-server v0.19.0 // indirect
	github.com/nats-io/nats.go v1.11.0
	github.com/nats-io/nkeys v0.3.0
	github.com/nats-io/stan.go v0.9.0
	github.com/pbnjay/memory v0.0.0-20201129165224-b12e5d931931
	github.com/pion/dtls/v2 v2.0.9
	github.com/pion/logging v0.2.2
	github.com/plgd-dev/go-coap/v2 v2.4.0
	github.com/prometheus-community/windows_exporter v0.16.0
	github.com/prometheus/client_golang v1.10.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.25.0
	github.com/prometheus/node_exporter v1.1.2
	github.com/spf13/pflag v1.0.5
	golang.org/x/sys v0.0.0-20210531080801-fdfd190a6549
	google.golang.org/grpc v1.35.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace (
	github.com/creack/pty => github.com/jeffreystoke/pty v1.1.12-0.20210531091229-b834701fbcc6
	github.com/dsnet/golib => github.com/dsnet/golib v0.0.0-20200723050859-c110804dfa93
	github.com/fsnotify/fsnotify => github.com/fsnotify/fsnotify v1.4.9
	github.com/klauspost/compress => github.com/klauspost/compress v1.12.3
	github.com/pion/dtls/v2 => github.com/pion/dtls/v2 v2.0.4
	github.com/spf13/cobra => github.com/spf13/cobra v1.1.1
	go.uber.org/atomic => go.uber.org/atomic v1.7.0
	go.uber.org/zap => go.uber.org/zap v1.16.0
	google.golang.org/grpc => github.com/grpc/grpc-go v1.35.0
	gopkg.in/alecthomas/kingpin.v2 => gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.4.0
)

// prometheus related
replace (
	github.com/Microsoft/go-winio => github.com/Microsoft/go-winio v0.5.0
	github.com/Microsoft/hcsshim => github.com/Microsoft/hcsshim v0.8.14
	github.com/StackExchange/wmi => github.com/jeffreystoke/wmi v1.1.5-0.20201112195122-b993dc474644
	github.com/bi-zone/go-ole => github.com/jeffreystoke/go-ole v1.2.6-0.20201112201217-834244b65d29
	github.com/prometheus-community/windows_exporter => github.com/prometheus-community/windows_exporter v0.16.0
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.10.0
	github.com/prometheus/client_model => github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common => github.com/prometheus/common v0.25.0
	github.com/prometheus/node_exporter => github.com/prometheus/node_exporter v1.1.2
	github.com/prometheus/procfs => github.com/prometheus/procfs v0.3.0
	honnef.co/go/tools => honnef.co/go/tools v0.0.1-2020.1.5
)

// go experimental
replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/exp => github.com/golang/exp v0.0.0-20210526181343-b47a03e3048a
	golang.org/x/lint => github.com/golang/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/net => github.com/golang/net v0.0.0-20210525063256-abc453219eb5
	golang.org/x/sync => github.com/golang/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys => github.com/golang/sys v0.0.0-20210531080801-fdfd190a6549
	golang.org/x/text => github.com/golang/text v0.3.6
	golang.org/x/tools => github.com/golang/tools v0.1.2
	golang.org/x/xerrors => github.com/golang/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

// Kubernetes v1.20.7
replace (
	k8s.io/api => github.com/kubernetes/api v0.20.7
	k8s.io/apiextensions-apiserver => github.com/kubernetes/apiextensions-apiserver v0.20.7
	k8s.io/apimachinery => github.com/kubernetes/apimachinery v0.20.7
	k8s.io/apiserver => github.com/kubernetes/apiserver v0.20.7
	k8s.io/cli-runtime => github.com/kubernetes/cli-runtime v0.20.7
	k8s.io/client-go => github.com/kubernetes/client-go v0.20.7
	k8s.io/cloud-provider => github.com/kubernetes/cloud-provider v0.20.7
	k8s.io/cluster-bootstrap => github.com/kubernetes/cluster-bootstrap v0.20.7
	k8s.io/code-generator => github.com/kubernetes/code-generator v0.20.7
	k8s.io/component-base => github.com/kubernetes/component-base v0.20.7
	k8s.io/component-helpers => github.com/kubernetes/component-helpers v0.20.7
	k8s.io/controller-manager => github.com/kubernetes/controller-manager v0.20.7
	k8s.io/cri-api => github.com/kubernetes/cri-api v0.20.7
	k8s.io/csi-translation-lib => github.com/kubernetes/csi-translation-lib v0.20.7
	k8s.io/klog => github.com/kubernetes/klog v1.0.0
	k8s.io/klog/v2 => github.com/kubernetes/klog/v2 v2.9.0
	k8s.io/kube-aggregator => github.com/kubernetes/kube-aggregator v0.20.7
	k8s.io/kube-controller-manager => github.com/kubernetes/kube-controller-manager v0.20.7
	k8s.io/kube-proxy => github.com/kubernetes/kube-proxy v0.20.7
	k8s.io/kube-scheduler => github.com/kubernetes/kube-scheduler v0.20.7
	k8s.io/kubectl => github.com/kubernetes/kubectl v0.20.7
	k8s.io/kubelet => github.com/kubernetes/kubelet v0.20.7
	k8s.io/kubernetes => github.com/kubernetes/kubernetes v1.20.7
	k8s.io/legacy-cloud-providers => github.com/kubernetes/legacy-cloud-providers v0.20.7
	k8s.io/metrics => github.com/kubernetes/metrics v0.20.7
	k8s.io/mount-utils => github.com/kubernetes/mount-utils v0.20.7
	k8s.io/sample-apiserver => github.com/kubernetes/sample-apiserver v0.20.7
	k8s.io/sample-cli-plugin => github.com/kubernetes/sample-cli-plugin v0.20.7
	k8s.io/sample-controller => github.com/kubernetes/sample-controller v0.20.7
	k8s.io/utils => github.com/kubernetes/utils v0.0.0-20210527160623-6fdb442a123b
)
