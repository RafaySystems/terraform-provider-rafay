package rafay

import (
	"fmt"
	"testing"

	configv2 "github.com/RafaySystems/rafay-common/proto/types/config"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/mock"
)

type clusterOverrideDepsMock struct {
	mock.Mock
}

func (m *clusterOverrideDepsMock) DeleteClusterOverride(name, projectID, overrideType string) error {
	args := m.Called(name, projectID, overrideType)
	return args.Error(0)
}

func (m *clusterOverrideDepsMock) GetProjectIdByName(name string) (string, error) {
	args := m.Called(name)
	return args.String(0), args.Error(1)
}

func (m *clusterOverrideDepsMock) UpdateClusterOverride(name, projectID string, spec models.ClusterOverrideSpec, status models.ClusterOverrideStatus, createIfNotPresent bool, labels map[string]string, annotations map[string]string) error {
	args := m.Called(name, projectID, spec, status, createIfNotPresent, labels, annotations)
	return args.Error(0)
}

func (m *clusterOverrideDepsMock) GetClusterOverride(name, projectID, overrideType string) (*models.ClusterOverride, error) {
	args := m.Called(name, projectID, overrideType)
	if co, ok := args.Get(0).(*models.ClusterOverride); ok {
		return co, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *clusterOverrideDepsMock) GetProjectNameById(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func setupClusterOverrideMocks(m *clusterOverrideDepsMock) func() {
	origDelete := deleteClusterOverrideFunc
	origGetProjectID := getProjectIdByNameFunc
	origUpdate := updateClusterOverrideFunc
	origGetOverride := getClusterOverrideFunc
	origGetProjectName := getProjectNameByIdFunc
	deleteClusterOverrideFunc = m.DeleteClusterOverride
	getProjectIdByNameFunc = m.GetProjectIdByName
	updateClusterOverrideFunc = m.UpdateClusterOverride
	getClusterOverrideFunc = m.GetClusterOverride
	getProjectNameByIdFunc = m.GetProjectNameById
	return func() {
		deleteClusterOverrideFunc = origDelete
		getProjectIdByNameFunc = origGetProjectID
		updateClusterOverrideFunc = origUpdate
		getClusterOverrideFunc = origGetOverride
		getProjectNameByIdFunc = origGetProjectName
	}
}

func testProviderFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"rafay": func() (*schema.Provider, error) {
			return New("v1")(), nil
		},
	}
}

const clusterOverrideInlineValuesConfig = `
resource "rafay_cluster_override" "tfdemo1" {
  metadata {
    name    = "tfdemocluster-override1"
    project = "terraform"
    labels = {
      "rafay.dev/overrideScope" = "clusterLabels"
      "rafay.dev/overrideType"  = "valuesFile"
    }
  }
  spec {
    cluster_selector = "rafay.dev/clusterName in (cluster-1)"
    cluster_placement {
      placement_type = "ClusterSpecific"
      cluster_labels {
        key   = "rafay.dev/clusterName"
        value = "cluster-1"
      }
    }
    resource_selector = "rafay.dev/name=override-addon"
    type              = "ClusterOverrideTypeAddon"
    override_values   = <<-EOS
    replicaCount: 1
    image:
      repository: nginx
      pullPolicy: Always
      tag: "1.19.8"
    service:
      type: ClusterIP
      port: 8080
    EOS
  }
}
`

const clusterOverrideGitRepoConfig = `
resource "rafay_cluster_override" "gitrepo" {
  metadata {
    name    = "tfdemocluster-override2"
    project = "terraform"
    labels = {
      "rafay.dev/overrideScope" = "clusterLabels"
      "rafay.dev/overrideType"  = "valuesFile"
    }
  }
  spec {
    cluster_selector = "key in (value)"
    cluster_placement {
      placement_type = "ClusterLabels"
      cluster_labels {
        key   = "key"
        value = "value"
      }
    }
    resource_selector = "rafay.dev/name=aws-lb-controller"
    type              = "ClusterOverrideTypeAddon"
    value_repo_ref    = "git-repo-name"
    values_repo_artifact_meta {
      git_options {
        revision = "main"
        repo_artifact_files {
          name          = "overrides.yaml"
          relative_path = "yaml/overrides.yaml"
          file_type     = "FileTypeNotSet"
        }
      }
    }
  }
}
`

func TestResourceClusterOverrideCreateAndDelete(t *testing.T) {
	mockDeps := new(clusterOverrideDepsMock)
	cleanup := setupClusterOverrideMocks(mockDeps)
	defer cleanup()

	const (
		resourceName = "rafay_cluster_override.tfdemo1"
		overrideName = "tfdemocluster-override1"
		projectName  = "terraform"
		projectID    = "proj-terraform-id"
	)

	mockDeps.On("GetProjectIdByName", projectName).Return(projectID, nil).Times(3)
	mockDeps.On("GetClusterOverride", overrideName, projectID, "ClusterOverrideTypeAddon").Return(nil, fmt.Errorf("not found")).Once()
	mockDeps.On("UpdateClusterOverride", overrideName, projectID, mock.Anything, mock.Anything, true, mock.Anything, mock.Anything).Return(nil).Once()
	mockDeps.On("DeleteClusterOverride", overrideName, projectID, "ClusterOverrideTypeAddon").Return(nil).Once()

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testProviderFactories(),
		CheckDestroy: func(state *terraform.State) error {
			if !mockDeps.AssertCalled(t, "DeleteClusterOverride", overrideName, projectID, "ClusterOverrideTypeAddon") {
				return fmt.Errorf("delete not invoked")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: clusterOverrideInlineValuesConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "metadata.0.name", overrideName),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.project", projectName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.type", "ClusterOverrideTypeAddon"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.cluster_selector", "rafay.dev/clusterName in (cluster-1)"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.cluster_placement.0.cluster_labels.0.key", "rafay.dev/clusterName"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.cluster_placement.0.cluster_labels.0.value", "cluster-1"),
				),
			},
		},
	})

	mockDeps.AssertExpectations(t)
}

func TestResourceClusterOverrideImportGitRepo(t *testing.T) {
	mockDeps := new(clusterOverrideDepsMock)
	cleanup := setupClusterOverrideMocks(mockDeps)
	defer cleanup()

	const (
		resourceName = "rafay_cluster_override.gitrepo"
		overrideName = "tfdemocluster-override2"
		projectName  = "terraform"
		projectID    = "proj-terraform-id"
	)

	mockDeps.On("GetProjectIdByName", projectName).Return(projectID, nil).Times(4)
	mockDeps.On("GetClusterOverride", overrideName, projectID, "ClusterOverrideTypeAddon").Return(nil, fmt.Errorf("not found")).Once()
	mockDeps.On("GetClusterOverride", overrideName, projectID, "ClusterOverrideTypeAddon").Return(sampleGitRepoOverride(projectID), nil).Once()
	mockDeps.On("UpdateClusterOverride", overrideName, projectID, mock.Anything, mock.Anything, true, mock.Anything, mock.Anything).Return(nil).Once()
	mockDeps.On("DeleteClusterOverride", overrideName, projectID, "ClusterOverrideTypeAddon").Return(nil).Once()
	mockDeps.On("GetProjectNameById", projectID).Return(projectName, nil).Once()

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testProviderFactories(),
		CheckDestroy: func(state *terraform.State) error {
			if !mockDeps.AssertCalled(t, "DeleteClusterOverride", overrideName, projectID, "ClusterOverrideTypeAddon") {
				return fmt.Errorf("delete not invoked")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config:        clusterOverrideGitRepoConfig,
				ImportState:   true,
				ResourceName:  resourceName,
				ImportStateId: fmt.Sprintf("%s/%s", overrideName, projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "metadata.0.name", overrideName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.value_repo_ref", "git-repo-name"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.values_repo_artifact_meta.0.git_options.0.revision", "main"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.values_repo_artifact_meta.0.git_options.0.repo_artifact_files.0.relative_path", "yaml/overrides.yaml"),
				),
			},
		},
	})

	mockDeps.AssertExpectations(t)
}

func sampleGitRepoOverride(projectID string) *models.ClusterOverride {
	return &models.ClusterOverride{
		RafayMeta: models.RafayMeta{
			Name: "tfdemocluster-override2",
			Labels: map[string]string{
				"rafay.dev/overrideScope": "clusterLabels",
				"rafay.dev/overrideType":  "valuesFile",
			},
		},
		ClusterOverrideSpec: models.ClusterOverrideSpec{
			ClusterSelector: "key in (value)",
			ClusterPlacement: models.PlacementSpec{
				PlacementType: models.PlacementType("ClusterLabels"),
				ClusterLabels: []*models.PlacementLabel{
					{Key: "key", Value: "value"},
				},
			},
			ResourceSelector: "rafay.dev/name=aws-lb-controller",
			Type:             "ClusterOverrideTypeAddon",
			RepositoryRef:    "git-repo-name",
			RepoArtifactMeta: models.RepoArtifactMeta{
				Git: &models.GitOptions{
					Revision: "main",
					RepoArtifactFiles: []models.RepoFile{
						{Name: "overrides.yaml", RelPath: "yaml/overrides.yaml", FileType: "FileTypeNotSet"},
					},
				},
			},
			ShareMode: configv2.CUSTOM.String(),
		},
		ClusterOverrideStatus: models.ClusterOverrideStatus{
			Projects: []models.ProjectOverrides{{ProjectID: projectID}},
		},
	}
}
