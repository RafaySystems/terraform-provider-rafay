package rafay

import (
	"errors"
	"fmt"
	"strings"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
)

// takes input given in the format of the terraform schema and populate the backend structure for that resource.
// convert from tf schema --> V3 schema in rafay-common proto

// GkeV3ConfigObject
func expandToGkeV3ConfigObject(p []interface{}) (*infrapb.ClusterSpec_Gke, error) {
	obj := &infrapb.ClusterSpec_Gke{
		Gke: &infrapb.GkeV3ConfigObject{}}

	if len(p) == 0 || p[0] == nil {
		return obj, errors.New("got nil or empty object for gke config")
	}
	in := p[0].(map[string]interface{})

	/*
		gcp project
		location
		controlplaneversion
		network
		security
		Feature
		nodepools
		prebootstrapcommands
	*/

	if v, ok := in["gcp_project"].(string); ok && len(v) > 0 {
		obj.Gke.GcpProject = v
	} else if !ok {
		return nil, errors.New("missing gcp project name")
	}

	if v, ok := in["control_plane_version"].(string); ok && len(v) > 0 {
		obj.Gke.ControlPlaneVersion = v
	} else if !ok {
		return nil, errors.New("missing controlplane version in input")
	}

	// location
	if v, ok := in["location"].([]interface{}); ok && len(v) > 0 {
		var err error
		obj.Gke.Location, err = expandToGkeV3Location(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand gke location from schema " + err.Error())
		}
	} else if !ok {
		// TODO: throw error that location is missing??
	}

	// network

	// security

	// feature

	// nodepools

	// prebootstrapCommands

	if v, ok := in["pre_bootstrap_commands"].([]interface{}); ok && len(v) > 0 {
		obj.Gke.PreBootstrapCommands = toArrayString(v)
	}

	return obj, nil

}

func expandToGkeV3Location(p []interface{}) (*infrapb.GkeLocation, error) {
	obj := &infrapb.GkeLocation{}

	if len(p) == 0 || p[0] == nil {
		return obj, errors.New("got nil for gke location")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	// GkeDefaultNodeLocation

	// zonal/regional

	if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
		var err error
		if strings.EqualFold(obj.Type, GKE_ZONAL_CLUSTER_TYPE) {
			obj.Config, err = expandToGkeV3ZonalCluster(v)
			return nil, fmt.Errorf("failed to expand gke zonal cluster " + err.Error())
		} else if strings.EqualFold(obj.Type, GKE_REGIONAL_CLUSTER_TYPE) {
			obj.Config, err = expandToGkeV3RegionalCluster(v)
			return nil, fmt.Errorf("failed to expand gke regional cluster " + err.Error())
		}
	}

	return obj, nil
}

func expandToGkeV3ZonalCluster(p []interface{}) (*infrapb.GkeLocation_Zonal, error) {
	obj := &infrapb.GkeLocation_Zonal{
		Zonal: &infrapb.GkeZonalCluster{},
	}

	if len(p) == 0 || p[0] == nil {
		return obj, errors.New("got nil for gke zonal cluster")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["zone"].(string); ok && len(v) > 0 {
		obj.Zonal.Zone = v
	}

	return obj, nil
}

func expandToGkeV3RegionalCluster(p []interface{}) (*infrapb.GkeLocation_Regional, error) {
	obj := &infrapb.GkeLocation_Regional{
		Regional: &infrapb.GkeRegionalCluster{},
	}

	if len(p) == 0 || p[0] == nil {
		return obj, errors.New("got nil for gke regional cluster")
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["region"].(string); ok && len(v) > 0 {
		obj.Regional.Region = v
	}

	if v, ok := in["zone"].(string); ok && len(v) > 0 {
		obj.Regional.Zone = v
	}

	return obj, nil
}
