package quay_io_ci_images_distributor

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	controllerruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	imagev1 "github.com/openshift/api/image/v1"

	"github.com/openshift/ci-tools/pkg/api"
	"github.com/openshift/ci-tools/pkg/testhelper"
)

func init() {
	if err := imagev1.AddToScheme(scheme.Scheme); err != nil {
		panic(fmt.Sprintf("failed to register imagev1 scheme: %v", err))
	}
}

func TestARTImages(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	testCases := []struct {
		name           string
		client         controllerruntimeclient.Client
		artImages      []ArtImage
		ignoredSources []IgnoredSource
		expected       map[string]Source
		expectedError  error
	}{
		{
			name: "basic case",
			artImages: []ArtImage{
				{Namespace: "openshift", Name: regexp.MustCompile("^release$")},
				{Namespace: "ocp", Name: regexp.MustCompile("^builder$")},
				{Namespace: "origin", Name: regexp.MustCompile("^scos.*")},
				{Namespace: "ocp", Name: regexp.MustCompile(`^4.\d+$`), Tag: regexp.MustCompile("^rhel-coreos.*")},
			},
			ignoredSources: []IgnoredSource{{Source: Source{ImageStreamTagReference: api.ImageStreamTagReference{Namespace: "openshift", Name: "release", Tag: "ignore"}}, Reason: "unit-test"}},
			client: fakeclient.NewClientBuilder().WithRuntimeObjects(
				&imagev1.ImageStream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "openshift",
						Name:      "release",
					},
					Status: imagev1.ImageStreamStatus{
						Tags: []imagev1.NamedTagEventList{
							{
								Tag: "a",
							},
							{
								Tag: "ignore",
							},
						},
					},
				},
				&imagev1.ImageStream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "some",
						Name:      "release",
					},
					Status: imagev1.ImageStreamStatus{
						Tags: []imagev1.NamedTagEventList{
							{
								Tag: "a",
							},
						},
					},
				},
				&imagev1.ImageStream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "openshift",
						Name:      "some",
					},
					Status: imagev1.ImageStreamStatus{
						Tags: []imagev1.NamedTagEventList{
							{
								Tag: "a",
							},
						},
					},
				},
				&imagev1.ImageStream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ocp",
						Name:      "builder",
					},
					Status: imagev1.ImageStreamStatus{
						Tags: []imagev1.NamedTagEventList{
							{
								Tag: "a",
							},
							{
								Tag: "b",
							},
						},
					},
				},
				&imagev1.ImageStream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ocp",
						Name:      "4.18",
					},
					Status: imagev1.ImageStreamStatus{
						Tags: []imagev1.NamedTagEventList{
							{
								Tag: "a",
							},
							{
								Tag: "rhel-coreos",
							},
							{
								Tag: "rhel-coreos-extensions",
							},
						},
					},
				},
				&imagev1.ImageStream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ocp",
						Name:      "some-4.18",
					},
					Status: imagev1.ImageStreamStatus{
						Tags: []imagev1.NamedTagEventList{
							{
								Tag: "a",
							},
							{
								Tag: "rhel-coreos",
							},
							{
								Tag: "rhel-coreos-extensions",
							},
						},
					},
				},
				&imagev1.ImageStream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ocp",
						Name:      "3.18",
					},
					Status: imagev1.ImageStreamStatus{
						Tags: []imagev1.NamedTagEventList{
							{
								Tag: "a",
							},
							{
								Tag: "rhel-coreos",
							},
							{
								Tag: "rhel-coreos-extensions",
							},
						},
					},
				},
				&imagev1.ImageStream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "origin",
						Name:      "4.18",
					},
					Status: imagev1.ImageStreamStatus{
						Tags: []imagev1.NamedTagEventList{
							{
								Tag: "a",
							},
							{
								Tag: "rhel-coreos",
							},
							{
								Tag: "rhel-coreos-extensions",
							},
						},
					},
				},
				&imagev1.ImageStream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "origin",
						Name:      "scos-4.16",
					},
					Status: imagev1.ImageStreamStatus{
						Tags: []imagev1.NamedTagEventList{
							{
								Tag: "a",
							},
							{
								Tag: "b",
							},
						},
					},
				},
				&imagev1.ImageStream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "origin",
						Name:      "4.16-okd-scos-2024-05-06-185114",
					},
					Status: imagev1.ImageStreamStatus{
						Tags: []imagev1.NamedTagEventList{
							{
								Tag: "a",
							},
						},
					},
				},
			).Build(),
			expected: map[string]Source{
				"openshift/release:a": {ImageStreamTagReference: api.ImageStreamTagReference{Namespace: "openshift", Name: "release", Tag: "a"}},
				"ocp/builder:a":       {ImageStreamTagReference: api.ImageStreamTagReference{Namespace: "ocp", Name: "builder", Tag: "a"}},
				"ocp/builder:b":       {ImageStreamTagReference: api.ImageStreamTagReference{Namespace: "ocp", Name: "builder", Tag: "b"}},
				"origin/scos-4.16:a":  {ImageStreamTagReference: api.ImageStreamTagReference{Namespace: "origin", Name: "scos-4.16", Tag: "a"}},
				"origin/scos-4.16:b":  {ImageStreamTagReference: api.ImageStreamTagReference{Namespace: "origin", Name: "scos-4.16", Tag: "b"}},
				"ocp/4.18:rhel-coreos": {
					ImageStreamTagReference: api.ImageStreamTagReference{Namespace: "ocp", Name: "4.18", Tag: "rhel-coreos"},
				},
				"ocp/4.18:rhel-coreos-extensions": {
					ImageStreamTagReference: api.ImageStreamTagReference{Namespace: "ocp", Name: "4.18", Tag: "rhel-coreos-extensions"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, actualError := ARTImages(context.TODO(), tc.client, tc.artImages, tc.ignoredSources)
			if diff := cmp.Diff(tc.expectedError, actualError, testhelper.EquateErrorMessage); diff != "" {
				t.Errorf("%s: actual does not match expected, diff: %s", tc.name, diff)
			}
			if diff := cmp.Diff(tc.expected, actual); diff != "" {
				t.Errorf("%s: actual does not match expected, diff: %s", tc.name, diff)
			}
		})
	}
}

func TestIgnored(t *testing.T) {
	testCases := []struct {
		name           string
		ignoredSources []IgnoredSource
		s              Source
		expected       bool
	}{
		{
			name: "source from tag and ignored source from image",
			ignoredSources: []IgnoredSource{
				{
					Source: Source{
						Image: "registry.ci.openshift.org/ocp/builder:rhel-7-golang-1.13",
					},
					Reason: "broken image",
				},
			},
			s: Source{
				ImageStreamTagReference: api.ImageStreamTagReference{
					Namespace: "ocp",
					Name:      "builder",
					Tag:       "rhel-7-golang-1.13",
				},
			},
			expected: true,
		},
		{
			name: "source from image and ignored source from tag",
			ignoredSources: []IgnoredSource{
				{
					Source: Source{
						ImageStreamTagReference: api.ImageStreamTagReference{
							Namespace: "ocp",
							Name:      "builder",
							Tag:       "rhel-7-golang-1.13",
						},
					},

					Reason: "broken image",
				},
			},
			s: Source{
				Image: "registry.ci.openshift.org/ocp/builder:rhel-7-golang-1.13",
			},
			expected: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ignored(tc.ignoredSources, tc.s, "unit-test")
			if diff := cmp.Diff(tc.expected, actual); diff != "" {
				t.Errorf("%s: actual does not match expected, diff: %s", tc.name, diff)
			}
		})
	}
}
