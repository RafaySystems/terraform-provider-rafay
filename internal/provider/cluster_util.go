package provider

import (
	"log"

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
