package scheme

import (
	extensionstsuruiov1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	rpaasv1alpha1 "github.com/tsuru/rpaas-operator/api/v1alpha1"
	tsuruv1 "github.com/tsuru/tsuru/provision/kubernetes/pkg/apis/tsuru/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	Scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
	utilruntime.Must(tsuruv1.AddToScheme(Scheme))
	utilruntime.Must(extensionstsuruiov1alpha1.AddToScheme(Scheme))
	utilruntime.Must(rpaasv1alpha1.AddToScheme(Scheme))
	//+kubebuilder:scaffold:scheme
}
