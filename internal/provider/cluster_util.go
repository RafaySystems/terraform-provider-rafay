package provider

import (
	"log"
	"slices"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
)

func getProjectIDFromName(projectName string) (string, error) {
	// derive project id from project name
	resp, err := project.GetProjectByName(projectName)
	if err != nil {
		log.Print("project name missing in the resource")
		return "", err
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		log.Printf("project does not exist")
		return "", err
	}
	return project.ID, nil
}

type clusterCTLResponse struct {
	TaskSetID  string                 `json:"taskset_id,omitempty"`
	Status     string                 `json:"status,omitempty"`
	Operations []*clusterCTLOperation `json:"operations"`
	Error      *errorResponse         `json:"error,omitempty"`
}

type clusterCTLOperation struct {
	Operation    string         `json:"operation,omitempty"`
	ResourceName string         `json:"resource_name,omitempty"`
	Status       string         `json:"status,omitempty"`
	Error        *errorResponse `json:"error,omitempty"`
}

type errorResponse struct {
	Type   string                 `json:"type,omitempty"`
	Status int                    `json:"status,omitempty"`
	Title  string                 `json:"title,omitempty"`
	Detail map[string]interface{} `json:"detail,omitempty"`
}

var BlueprintSyncConditions = []models.ClusterConditionType{
	models.ClusterRegister,
	models.ClusterCheckIn,
	models.ClusterNamespaceSync,
	models.ClusterBlueprintSync,
}

func getClusterConditions(edgeId, projectId string) (bool, bool, error) {
	cluster, err := cluster.GetClusterWithEdgeID(edgeId, projectId, uaDef)
	if err != nil {
		log.Printf("error while getCluster %s", err.Error())
		return false, false, err
	}

	clusterConditions := cluster.Cluster.Conditions
	failureFlag := false
	readyFlag := false
	for _, condition := range clusterConditions {
		if slices.Contains(BlueprintSyncConditions, condition.Type) && condition.Status == models.Failed {
			failureFlag = true
		}
		if condition.Type == models.ClusterReady && condition.Status == models.Success {
			readyFlag = true
		}
	}
	return failureFlag, readyFlag, nil
}
