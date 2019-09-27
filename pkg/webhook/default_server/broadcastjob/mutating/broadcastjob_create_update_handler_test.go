package mutating

import (
	"context"
	"reflect"
	"testing"

	"github.com/appscode/jsonpatch"
	"github.com/openkruise/kruise/pkg/apis"
	appsv1alpha1 "github.com/openkruise/kruise/pkg/apis/apps/v1alpha1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

func TestHandle(t *testing.T) {
	_ = apis.AddToScheme(scheme.Scheme)

	oldBroadcastJobStr := `{
    "metadata": {
        "creationTimestamp": null
    },
    "spec": {
        "parallelism": 100,
        "template": {
            "metadata": {
                "creationTimestamp": null
            },
            "spec": {
                "containers": null,
                "restartPolicy": "Always",
                "terminationGracePeriodSeconds": 30,
                "dnsPolicy": "ClusterFirst",
                "securityContext": {},
                "schedulerName": "fake-scheduler"
            }
        },
        "paused": false
    },
    "status": {
        "active": 0,
        "succeeded": 0,
        "failed": 0,
        "desired": 0,
        "phase": ""
    }
}`

	decoder, _ := admission.NewDecoder(scheme.Scheme)
	handler := BroadcastJobCreateUpdateHandler{
		Decoder: decoder,
	}

	req := types.Request{
		AdmissionRequest: &admissionv1beta1.AdmissionRequest{
			Object: runtime.RawExtension{
				Raw: []byte(oldBroadcastJobStr),
			},
		},
	}
	resp := handler.Handle(context.TODO(), req)

	expectedPatches := []jsonpatch.JsonPatchOperation{
		{
			Operation: "add",
			Path:      "/spec/completionPolicy",
			Value:     map[string]interface{}{"type": string(appsv1alpha1.Always)},
		},
		{
			Operation: "add",
			Path:      "/spec/failurePolicy",
			Value:     map[string]interface{}{"type": string(appsv1alpha1.FailurePolicyTypeFailFast)},
		},
		{
			Operation: "remove",
			Path:      "/spec/paused",
		},
	}

	if !reflect.DeepEqual(expectedPatches, resp.Patches) {
		t.Fatalf("expected patches %+v, got patches %+v", expectedPatches, resp.Patches)
	}
}
