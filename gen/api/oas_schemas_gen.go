// Code generated by ogen, DO NOT EDIT.

package api

// Ref: #/components/schemas/Cluster
type Cluster struct {
	ID       string `json:"id"`
	RpcAddr  string `json:"rpcAddr"`
	IsLeader bool   `json:"isLeader"`
}

// GetID returns the value of ID.
func (s *Cluster) GetID() string {
	return s.ID
}

// GetRpcAddr returns the value of RpcAddr.
func (s *Cluster) GetRpcAddr() string {
	return s.RpcAddr
}

// GetIsLeader returns the value of IsLeader.
func (s *Cluster) GetIsLeader() bool {
	return s.IsLeader
}

// SetID sets the value of ID.
func (s *Cluster) SetID(val string) {
	s.ID = val
}

// SetRpcAddr sets the value of RpcAddr.
func (s *Cluster) SetRpcAddr(val string) {
	s.RpcAddr = val
}

// SetIsLeader sets the value of IsLeader.
func (s *Cluster) SetIsLeader(val bool) {
	s.IsLeader = val
}

type Clusters []Cluster

// HealthInternalServerError is response for Health operation.
type HealthInternalServerError struct{}

func (*HealthInternalServerError) healthRes() {}

// HealthOK is response for Health operation.
type HealthOK struct{}

func (*HealthOK) healthRes() {}

// ListClusterBadRequest is response for ListCluster operation.
type ListClusterBadRequest struct{}

func (*ListClusterBadRequest) listClusterRes() {}

// ListClusterInternalServerError is response for ListCluster operation.
type ListClusterInternalServerError struct{}

func (*ListClusterInternalServerError) listClusterRes() {}

// ListClusterNotFound is response for ListCluster operation.
type ListClusterNotFound struct{}

func (*ListClusterNotFound) listClusterRes() {}

// Ref: #/components/schemas/ListClusterResponse
type ListClusterResponse struct {
	Clusters Clusters `json:"clusters"`
}

// GetClusters returns the value of Clusters.
func (s *ListClusterResponse) GetClusters() Clusters {
	return s.Clusters
}

// SetClusters sets the value of Clusters.
func (s *ListClusterResponse) SetClusters(val Clusters) {
	s.Clusters = val
}

func (*ListClusterResponse) listClusterRes() {}