package parser

import (
	"fmt"
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	apiAttributes "github.com/devfile/api/v2/pkg/attributes"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/pkg/testingutil/filesystem"
	"strings"
	"testing"
)

func TestWriteYamlDevfile(t *testing.T) {

	var (
		schemaVersion = "2.2.0"
		testName      = "TestName"
		uri           = "./relative/path/deploy.yaml"
		uri2          = "./relative/path/deploy2.yaml"
		attributes    = apiAttributes.Attributes{}.PutString(K8sLikeComponentOriginalURIKey, uri)
		attributes2   = apiAttributes.Attributes{}.PutString(K8sLikeComponentOriginalURIKey, uri2)
	)

	t.Run("write yaml devfile", func(t *testing.T) {

		// Use fakeFs
		fs := filesystem.NewFakeFs()

		// DevfileObj
		devfileObj := DevfileObj{
			Ctx: devfileCtx.FakeContext(fs, OutputDevfileYamlPath),
			Data: &v2.DevfileV2{
				Devfile: v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: schemaVersion,
						Metadata: devfilepkg.DevfileMetadata{
							Name: testName,
						},
					},
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Components: []v1.Component{
								{
									Name:       "kubeComp",
									Attributes: attributes,
									ComponentUnion: v1.ComponentUnion{
										Kubernetes: &v1.KubernetesComponent{
											K8sLikeComponent: v1.K8sLikeComponent{
												K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
													Inlined: "placeholder",
												},
											},
										},
									},
								},
								{
									Name:       "openshiftComp",
									Attributes: attributes2,
									ComponentUnion: v1.ComponentUnion{
										Openshift: &v1.OpenshiftComponent{
											K8sLikeComponent: v1.K8sLikeComponent{
												K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
													Inlined: "placeholder",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		devfileObj.Ctx.SetConvertUriToInlined(true)

		// test func()
		err := devfileObj.WriteYamlDevfile()
		if err != nil {
			t.Errorf("TestWriteYamlDevfile() unexpected error: '%v'", err)
		}

		if _, err := fs.Stat(OutputDevfileYamlPath); err != nil {
			t.Errorf("TestWriteYamlDevfile() unexpected error: '%v'", err)
		}

		data, err := fs.ReadFile(OutputDevfileYamlPath)
		if err != nil {
			t.Errorf("TestWriteYamlDevfile() unexpected error: '%v'", err)
		} else {
			content := string(data)
			if strings.Contains(content, "inlined") || strings.Contains(content, K8sLikeComponentOriginalURIKey) {
				t.Errorf("TestWriteYamlDevfile() failed: kubernetes component should not contain inlined or %s", K8sLikeComponentOriginalURIKey)
			}
			if !strings.Contains(content, fmt.Sprintf("uri: %s", uri)) {
				t.Errorf("TestWriteYamlDevfile() failed: kubernetes component does not contain uri")
			}
			if !strings.Contains(content, fmt.Sprintf("uri: %s", uri2)) {
				t.Errorf("TestWriteYamlDevfile() failed: openshift component does not contain uri")
			}
		}
	})
}
