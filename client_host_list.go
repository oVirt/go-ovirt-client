package ovirtclient

import (
	"fmt"
	ovirtsdk "github.com/ovirt/go-ovirt"
)

func (o *oVirtClient) ListHosts() ([]Host, error) {
	response, err := o.conn.SystemService().HostsService().List().Send()
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts (%w)", err)
	}
	sdkHosts, ok := response.Hosts()
	if !ok {
		return nil, fmt.Errorf("host list response didn't contain hosts")
	}
	result := make([]Host, len(sdkHosts.Slice()))
	for i, sdkHost := range sdkHosts.Slice() {
		result[i], err = convertSDKHost(sdkHost)
		if err != nil {
			return nil, fmt.Errorf("failed to convert host %d in listing (%w)", i, err)
		}
	}
	return result, nil
}

// TODO: I think we should remove the wrapping code of converting from the SDK since it makes using functions of the client
// in other parts problematic, a lot of the time we just need the SDK version... this is an example
func (o *oVirtClient) ListHostsInCluster(clusterID string) ([]Host, error) {
	sdkHosts, err := o.listHostsInCluster(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts in cluster %s (%w)", clusterID, err)
	}
	result := make([]Host, len(sdkHosts.Slice()))
	for i, sdkHost := range sdkHosts.Slice() {
		result[i], err = convertSDKHost(sdkHost)
		if err != nil {
			return nil, fmt.Errorf("failed to convert host %d in listing (%w)", i, err)
		}
	}
	return result, nil
}

func (o *oVirtClient) listHostsInCluster(clusterID string) (*ovirtsdk.HostSlice, error) {
	clusterResp, err := o.conn.SystemService().ClustersService().ClusterService(clusterID).Get().Send()
	if err != nil {
		return nil, fmt.Errorf("failed find cluster with id %s (%w)", clusterID, err)
	}
	clusterName := clusterResp.MustCluster().MustName()
	hostsInClusterResp, err := o.conn.SystemService().HostsService().List().Search(
		fmt.Sprintf("cluster=%s", clusterName)).Send()
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts in cluster %s (%w)", clusterID, err)
	}
	hosts, ok := hostsInClusterResp.Hosts()
	if !ok {
		return nil, fmt.Errorf("host list response didn't contain hosts")
	}
	return hosts, nil
}
