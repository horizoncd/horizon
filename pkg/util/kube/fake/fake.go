package fake

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/remotecommand"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
	"k8s.io/kubectl/pkg/scheme"
)

func NewFakeClient() *fake.RESTClient {
	codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
	ns := scheme.Codecs.WithoutConversion()
	restClient := &fake.RESTClient{
		GroupVersion:         schema.GroupVersion{Group: "", Version: "v1"},
		NegotiatedSerializer: ns,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch m, url := req.Method, req.URL.String(); {
			case m == http.MethodGet && strings.Contains(url, "foo1"):
				body := cmdtesting.ObjBody(codec, execPod1())
				return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: body}, nil
			case m == http.MethodGet && strings.Contains(url, "foo2"):
				body := cmdtesting.ObjBody(codec, execPod2())
				return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: body}, nil
			case m == http.MethodPost:
				body := cmdtesting.StringBody(req.URL.String())
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: body}, nil
			default:
				return nil, fmt.Errorf("unexpected request")
			}
		}),
	}
	return restClient
}

func NewEmptyClient() *restclient.Config {
	return &restclient.Config{
		APIPath: "/api",
		ContentConfig: restclient.ContentConfig{
			NegotiatedSerializer: scheme.Codecs,
			GroupVersion: &schema.GroupVersion{
				Version: "v1",
			},
		},
	}
}

func execPod1() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "foo1", Namespace: "test", ResourceVersion: "10"},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyAlways,
			DNSPolicy:     corev1.DNSClusterFirst,
			Containers: []corev1.Container{
				{
					Name: "bar",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
}

func execPod2() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "foo2", Namespace: "test", ResourceVersion: "10"},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyAlways,
			DNSPolicy:     corev1.DNSClusterFirst,
			Containers: []corev1.Container{
				{
					Name: "bar",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
}

type RemoteExecutor struct {
	Client *http.Client
}

// Execute is a method of the remote executor. It executes an HTTP request,
// gets the response, and writes it to standard output.
func (f *RemoteExecutor) Execute(
	method string, // Request method
	url *url.URL, // Request URL
	_ *restclient.Config, // REST client configuration
	_ io.Reader, // Standard input
	stdout, // Standard output
	_ io.Writer, // Standard output and error output
	_ bool, // Whether it is a TTY
	_ remotecommand.TerminalSizeQueue, // Terminal size queue
) error {
	// Construct request object
	req, err := http.NewRequest(method, url.String(), nil)
	if err != nil {
		return err
	}

	resp, err := f.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response code %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	stdout.Write(data)
	return nil
}
