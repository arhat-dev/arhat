module arhat.dev/arhat

go 1.15

replace (
	cloud.google.com/go => cloud.google.com/go v0.63.0
	github.com/Microsoft/go-winio => github.com/Microsoft/go-winio v0.4.14
	github.com/Microsoft/hcsshim => github.com/Microsoft/hcsshim v0.8.9
	github.com/OpenPeeDeeP/depguard => github.com/OpenPeeDeeP/depguard v1.0.1
	github.com/PuerkitoBio/purell => github.com/PuerkitoBio/purell v1.1.1
	github.com/StackExchange/wmi => github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d
	github.com/alecthomas/participle => github.com/alecthomas/participle v0.5.0
	github.com/alecthomas/units => github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d
	github.com/asaskevich/govalidator => github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.31.7
	github.com/aws/aws-sdk-go-v2 => github.com/aws/aws-sdk-go-v2 v0.23.0
	github.com/containerd/typeurl => github.com/containerd/typeurl v1.0.1
	github.com/creack/pty => github.com/creack/pty v1.1.11
	github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/docker/spdystream => github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c
	github.com/dsnet/golib => github.com/dsnet/golib v0.0.0-20190531212259-571cdbcff553
	github.com/fsnotify/fsnotify => github.com/fsnotify/fsnotify v1.4.9
	github.com/golang/protobuf => github.com/golang/protobuf v1.4.2
	github.com/google/gofuzz => github.com/google/gofuzz v1.0.0
	github.com/gorilla/mux => github.com/gorilla/mux v1.8.0
	github.com/jmespath/go-jmespath => github.com/jmespath/go-jmespath v0.3.0
	github.com/pion/dtls/v2 => github.com/pion/dtls/v2 v2.0.1
	github.com/spf13/cobra => github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify => github.com/stretchr/testify v1.3.0
	github.com/vishvananda/netlink => github.com/vishvananda/netlink v1.0.1-0.20190725224917-b4e9f47a11c0
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200329194405-dd816f0735f8
	go.uber.org/atomic => go.uber.org/atomic v1.6.0
	go.uber.org/zap => go.uber.org/zap v1.15.0
	google.golang.org/api => google.golang.org/api v0.21.0
	google.golang.org/appengine => google.golang.org/appengine v1.6.6
	google.golang.org/grpc => github.com/grpc/grpc-go v1.29.1
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
)

// prometheus related
replace (
	github.com/prometheus-community/windows_exporter => github.com/prometheus-community/windows_exporter v0.14.0
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model => github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common => github.com/prometheus/common v0.13.0
	github.com/prometheus/node_exporter => github.com/prometheus/node_exporter v1.0.1
	github.com/prometheus/procfs => github.com/prometheus/procfs v0.1.3
	honnef.co/go/tools => honnef.co/go/tools v0.0.1-2020.1.5
)

// libpod v2.0.6
replace (
	github.com/containernetworking/cni => github.com/containernetworking/cni v0.8.0
	github.com/containernetworking/plugins => github.com/containernetworking/plugins v0.8.7
	github.com/containers/buildah => github.com/containers/buildah v1.15.1
	github.com/containers/common => github.com/containers/common v0.21.0
	github.com/containers/image/v5 => github.com/containers/image/v5 v5.5.2
	github.com/containers/libpod/v2 => github.com/containers/libpod/v2 v2.0.6
	github.com/containers/psgo => github.com/containers/psgo v1.5.1
	github.com/containers/storage => github.com/containers/storage v1.23.4
	github.com/coreos/go-systemd/v22 => github.com/coreos/go-systemd/v22 v22.1.0
	github.com/docker/distribution => github.com/docker/distribution v2.7.1+incompatible
	github.com/godbus/dbus/v5 => github.com/godbus/dbus/v5 v5.0.3
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc90
	github.com/opencontainers/runtime-spec => github.com/opencontainers/runtime-spec v1.0.3-0.20200520003142-237cc4f519e2
	github.com/opencontainers/runtime-tools => github.com/opencontainers/runtime-tools v0.9.1-0.20200714183735-07406c5828aa
	github.com/opencontainers/selinux => github.com/opencontainers/selinux v1.6.0
	github.com/openshift/imagebuilder => github.com/openshift/imagebuilder v1.1.6
)

// go experimental
replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/exp => github.com/golang/exp v0.0.0-20200513190911-00229845015e
	golang.org/x/lint => github.com/golang/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/net => github.com/golang/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/oauth2 => github.com/golang/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync => github.com/golang/sync v0.0.0-20200317015054-43a5402ce75a
	golang.org/x/sys => github.com/golang/sys v0.0.0-20200821140526-fda516888d29
	golang.org/x/text => github.com/golang/text v0.3.2
	golang.org/x/tools => github.com/golang/tools v0.0.0-20200811032001-fd80f4dbb3ea
	golang.org/x/xerrors => github.com/golang/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

// Kubernetes v1.18.8
replace (
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v37.1.0+incompatible
	github.com/container-storage-interface/spec => github.com/container-storage-interface/spec v1.3.0
	github.com/containerd/containerd => github.com/containerd/containerd v1.3.4
	github.com/evanphx/json-patch => github.com/evanphx/json-patch/v5 v5.0.0
	github.com/heketi/heketi => github.com/heketi/heketi v9.0.1-0.20190917153846-c2e2a4ab7ab9+incompatible
	github.com/mindprince/gonvml => github.com/mindprince/gonvml v0.0.0-20190828220739-9ebdce4bb989
	k8s.io/api => github.com/kubernetes/api v0.18.8
	k8s.io/apiextensions-apiserver => github.com/kubernetes/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => github.com/kubernetes/apimachinery v0.18.8
	k8s.io/apiserver => github.com/kubernetes/apiserver v0.18.8
	k8s.io/cli-runtime => github.com/kubernetes/cli-runtime v0.18.8
	k8s.io/client-go => github.com/kubernetes/client-go v0.18.8
	k8s.io/cloud-provider => github.com/kubernetes/cloud-provider v0.18.8
	k8s.io/cluster-bootstrap => github.com/kubernetes/cluster-bootstrap v0.18.8
	k8s.io/code-generator => github.com/kubernetes/code-generator v0.18.8
	k8s.io/component-base => github.com/kubernetes/component-base v0.18.8
	k8s.io/cri-api => github.com/kubernetes/cri-api v0.18.8
	k8s.io/csi-translation-lib => github.com/kubernetes/csi-translation-lib v0.18.8
	k8s.io/klog => github.com/kubernetes/klog v1.0.0
	k8s.io/klog/v2 => github.com/kubernetes/klog/v2 v2.3.0
	k8s.io/kube-aggregator => github.com/kubernetes/kube-aggregator v0.18.8
	k8s.io/kube-controller-manager => github.com/kubernetes/kube-controller-manager v0.18.8
	k8s.io/kube-proxy => github.com/kubernetes/kube-proxy v0.18.8
	k8s.io/kube-scheduler => github.com/kubernetes/kube-scheduler v0.18.8
	k8s.io/kubectl => github.com/kubernetes/kubectl v0.18.8
	k8s.io/kubelet => github.com/kubernetes/kubelet v0.18.8
	k8s.io/kubernetes => github.com/kubernetes/kubernetes v1.18.8
	k8s.io/legacy-cloud-providers => github.com/kubernetes/legacy-cloud-providers v0.18.8
	k8s.io/metrics => github.com/kubernetes/metrics v0.18.8
	k8s.io/sample-apiserver => github.com/kubernetes/sample-apiserver v0.18.8
	k8s.io/utils => github.com/kubernetes/utils v0.0.0-20200821003339-5e75c0163111
)

// azure autorest
replace (
	github.com/Azure/go-amqp => github.com/Azure/go-amqp v0.12.6
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.11.3
	github.com/Azure/go-autorest/autorest/adal => github.com/Azure/go-autorest/autorest/adal v0.9.1
	github.com/Azure/go-autorest/autorest/azure/auth => github.com/Azure/go-autorest/autorest/azure/auth v0.4.0
	github.com/Azure/go-autorest/autorest/date => github.com/Azure/go-autorest/autorest/date v0.2.0
	github.com/Azure/go-autorest/autorest/mocks => github.com/Azure/go-autorest/autorest/mocks v0.3.0
	github.com/Azure/go-autorest/autorest/to => github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/Azure/go-autorest/autorest/validation => github.com/Azure/go-autorest/autorest/validation v0.2.0
)

require (
	arhat.dev/abbot-proto v0.0.0-20200908152444-df59421a64d2
	arhat.dev/aranya-proto v0.0.0-20200923000120-7097a702cf8b
	arhat.dev/pkg v0.0.0-20200922111851-9707225a9ad0
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/containerd/containerd v1.3.3 // indirect
	github.com/containers/common v0.21.0
	github.com/containers/image/v5 v5.5.2
	github.com/containers/libpod/v2 v2.0.6
	github.com/containers/storage v1.23.4
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v1.4.2-0.20191219165747-a9416c67da9f
	github.com/docker/go-connections v0.4.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gogo/protobuf v1.3.1
	github.com/goiiot/libmqtt v0.9.5
	github.com/goiiot/libserial v0.0.0-20200819082916-cd2b4aa0cf09
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mholt/archiver/v3 v3.3.0
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mssola/user_agent v0.5.2
	github.com/opencontainers/image-spec v1.0.2-0.20190823105129-775207bd45b6
	github.com/opencontainers/runtime-spec v1.0.3-0.20200728170252-4d89ac9fbff6
	github.com/pbnjay/memory v0.0.0-20190104145345-974d429e7ae4
	github.com/pion/dtls/v2 v2.0.1-0.20200503085337-8e86b3a7d585
	github.com/pion/logging v0.2.2
	github.com/plgd-dev/go-coap/v2 v2.0.4
	github.com/prometheus-community/windows_exporter v0.14.0
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.13.0
	github.com/prometheus/node_exporter v1.0.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/vishvananda/netns v0.0.0-20200520041808-52d707b772fe // indirect
	go.uber.org/multierr v1.5.0
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de // indirect
	golang.org/x/sys v0.0.0-20200824131525-c12d262b63d8
	golang.org/x/tools v0.0.0-20200811032001-fd80f4dbb3ea // indirect
	google.golang.org/grpc v1.31.1
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/client-go v0.18.8
	k8s.io/utils v0.0.0-20200821003339-5e75c0163111 // indirect
)
