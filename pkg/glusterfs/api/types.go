//
// Copyright (c) 2016 The heketi Authors
//
// This file is licensed to you under your choice of the GNU Lesser
// General Public License, version 3 or any later version (LGPLv3 or
// later), as published by the Free Software Foundation,
// or under the Apache License, Version 2.0 <LICENSE-APACHE2 or
// http://www.apache.org/licenses/LICENSE-2.0>.
//
// You may not use this file except in compliance with those terms.
//

//
// Please see https://github.com/cloud-tools/heketi/wiki/API
// for documentation
//
package api

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

var (
	// Restricting the deviceName to much smaller subset of Unix Path
	// as unix path takes almost everything except NULL
	deviceNameRe = regexp.MustCompile("^/[a-zA-Z0-9_.:/-]+$")

	// Volume name constraints decided by looking at
	// "cli_validate_volname" function in cli-cmd-parser.c of gluster code
	volumeNameRe = regexp.MustCompile("^[a-zA-Z0-9_-]+$")

	blockVolNameRe = regexp.MustCompile("^[a-zA-Z0-9_-]+$")

	tagNameRe = regexp.MustCompile("^[a-zA-Z0-9_.-]+$")
)

// ValidateUUID is written this way because heketi UUID does not
// conform to neither UUID v4 nor v5.
func ValidateUUID(value interface{}) error {
	s, _ := value.(string)
	err := validation.Validate(s, validation.RuneLength(32, 32), is.Hexadecimal)
	if err != nil {
		return fmt.Errorf("%v is not a valid UUID", s)
	}
	return nil
}

// State
type EntryState string

const (
	EntryStateUnknown EntryState = ""
	EntryStateOnline  EntryState = "online"
	EntryStateOffline EntryState = "offline"
	EntryStateFailed  EntryState = "failed"
)

func ValidateEntryState(value interface{}) error {
	s, _ := value.(EntryState)
	err := validation.Validate(s, validation.Required, validation.In(EntryStateOnline, EntryStateOffline, EntryStateFailed))
	if err != nil {
		return fmt.Errorf("%v is not valid state", s)
	}
	return nil
}

type DurabilityType string

const (
	DurabilityReplicate      DurabilityType = "replicate"
	DurabilityDistributeOnly DurabilityType = "none"
	DurabilityEC             DurabilityType = "disperse"
)

func ValidateDurabilityType(value interface{}) error {
	s, _ := value.(DurabilityType)
	err := validation.Validate(s, validation.Required, validation.In(DurabilityReplicate, DurabilityDistributeOnly, DurabilityEC))
	if err != nil {
		return fmt.Errorf("%v is not a valid durability type", s)
	}
	return nil
}

// Common
type StateRequest struct {
	State EntryState `json:"state"`
}

func (statereq StateRequest) Validate() error {
	return validation.ValidateStruct(&statereq,
		validation.Field(&statereq.State, validation.Required, validation.By(ValidateEntryState)),
	)
}

// Storage values in KB
type StorageSize struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
	Used  uint64 `json:"used"`
}

type HostAddresses struct {
	Manage  sort.StringSlice `json:"manage"`
	Storage sort.StringSlice `json:"storage"`
}

func ValidateManagementHostname(value interface{}) error {
	s, _ := value.(sort.StringSlice)
	for _, fqdn := range s {
		err := validation.Validate(fqdn, validation.Required, is.Host)
		if err != nil {
			return fmt.Errorf("%v is not a valid manage hostname", s)
		}
	}
	return nil
}

func ValidateStorageHostname(value interface{}) error {
	s, _ := value.(sort.StringSlice)
	for _, ip := range s {
		err := validation.Validate(ip, validation.Required, is.Host)
		if err != nil {
			return fmt.Errorf("%v is not a valid storage hostname", s)
		}
	}
	return nil
}

func (hostadd HostAddresses) Validate() error {
	return validation.ValidateStruct(&hostadd,
		validation.Field(&hostadd.Manage, validation.Required, validation.By(ValidateManagementHostname)),
		validation.Field(&hostadd.Storage, validation.Required, validation.By(ValidateStorageHostname)),
	)
}

// Brick
type BrickInfo struct {
	Id       string `json:"id"`
	Path     string `json:"path"`
	DeviceId string `json:"device"`
	NodeId   string `json:"node"`
	VolumeId string `json:"volume"`

	// Size in KB
	Size uint64 `json:"size"`
}

// Device
type Device struct {
	Name string            `json:"name"`
	Tags map[string]string `json:"tags,omitempty"`
}

func (dev Device) Validate() error {
	return validation.ValidateStruct(&dev,
		validation.Field(&dev.Name, validation.Required, validation.Match(deviceNameRe)),
		validation.Field(&dev.Tags, validation.By(ValidateTags)),
	)
}

type DeviceAddRequest struct {
	Device
	NodeId      string `json:"node"`
	DestroyData bool   `json:"destroydata,omitempty"`
}

func (devAddReq DeviceAddRequest) Validate() error {
	return validation.ValidateStruct(&devAddReq,
		validation.Field(&devAddReq.Device, validation.Required),
		validation.Field(&devAddReq.NodeId, validation.Required, validation.By(ValidateUUID)),
		validation.Field(&devAddReq.DestroyData, validation.In(true, false)),
	)
}

type DeviceInfo struct {
	Device
	Storage StorageSize `json:"storage"`
	Id      string      `json:"id"`
}

type DeviceInfoResponse struct {
	DeviceInfo
	State  EntryState  `json:"state"`
	Bricks []BrickInfo `json:"bricks"`
}

// Node
type NodeAddRequest struct {
	Zone      int               `json:"zone"`
	Hostnames HostAddresses     `json:"hostnames"`
	ClusterId string            `json:"cluster"`
	Tags      map[string]string `json:"tags,omitempty"`
}

func (req NodeAddRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Zone, validation.Required, validation.Min(1)),
		validation.Field(&req.Hostnames, validation.Required),
		validation.Field(&req.ClusterId, validation.Required, validation.By(ValidateUUID)),
		validation.Field(&req.Tags, validation.By(ValidateTags)),
	)
}

type NodeInfo struct {
	NodeAddRequest
	Id string `json:"id"`
}

type NodeInfoResponse struct {
	NodeInfo
	State       EntryState           `json:"state"`
	DevicesInfo []DeviceInfoResponse `json:"devices"`
}

// Cluster

type ClusterFlags struct {
	Block bool   `json:"block"`
	File  bool   `json:"file"`
	Side  string `json:"side,omitempty"`
}

type Cluster struct {
	Volumes []VolumeInfoResponse `json:"volumes"`
	Nodes   []NodeInfoResponse   `json:"nodes"`
	Id      string               `json:"id"`
	ClusterFlags
	MasterSlaveCluster
}

type TopologyInfoResponse struct {
	ClusterList []Cluster `json:"clusters"`
}

type ClusterCreateRequest struct {
	ClusterFlags
}

type ClusterSetFlagsRequest struct {
	ClusterFlags
}

type ClusterInfoResponse struct {
	Id        string           `json:"id"`
	Nodes     sort.StringSlice `json:"nodes"`
	Volumes   sort.StringSlice `json:"volumes"`
	Snapshots sort.StringSlice `json:"snapshots"`
	ClusterFlags
	BlockVolumes sort.StringSlice `json:"blockvolumes"`
	MasterSlaveCluster
}

type ClusterListResponse struct {
	Clusters []string `json:"clusters"`
}

// Durabilities
type ReplicaDurability struct {
	Replica int `json:"replica,omitempty"`
}

type DisperseDurability struct {
	Data       int `json:"data,omitempty"`
	Redundancy int `json:"redundancy,omitempty"`
}

// Volume
type VolumeDurabilityInfo struct {
	Type      DurabilityType     `json:"type,omitempty"`
	Replicate ReplicaDurability  `json:"replicate,omitempty"`
	Disperse  DisperseDurability `json:"disperse,omitempty"`
}

type VolumeCreateRequest struct {
	// Size in GiB
	Size                 int                  `json:"size"`
	Clusters             []string             `json:"clusters,omitempty"`
	Name                 string               `json:"name"`
	Durability           VolumeDurabilityInfo `json:"durability,omitempty"`
	Gid                  int64                `json:"gid,omitempty"`
	GlusterVolumeOptions []string             `json:"glustervolumeoptions,omitempty"`
	Block                bool                 `json:"block,omitempty"`
	Snapshot             struct {
		Enable bool    `json:"enable"`
		Factor float32 `json:"factor"`
	} `json:"snapshot"`
}

func (volCreateRequest VolumeCreateRequest) Validate() error {
	return validation.ValidateStruct(&volCreateRequest,
		validation.Field(&volCreateRequest.Size, validation.Required, validation.Min(1)),
		validation.Field(&volCreateRequest.Clusters, validation.By(ValidateUUID)),
		validation.Field(&volCreateRequest.Name, validation.Match(volumeNameRe)),
		validation.Field(&volCreateRequest.Durability, validation.Skip),
		validation.Field(&volCreateRequest.Gid, validation.Skip),
		validation.Field(&volCreateRequest.GlusterVolumeOptions, validation.Skip),
		validation.Field(&volCreateRequest.Block, validation.In(true, false)),
		// This is possibly a bug in validation lib, ignore next two lines for now
		// validation.Field(&volCreateRequest.Snapshot.Enable, validation.In(true, false)),
		// validation.Field(&volCreateRequest.Snapshot.Factor, validation.Min(1.0)),
	)
}

type BlockRestriction string

const (
	Unrestricted   BlockRestriction = ""
	Locked         BlockRestriction = "locked"
	LockedByUpdate BlockRestriction = "locked-by-update"
)

func (br BlockRestriction) String() string {
	switch br {
	case Unrestricted:
		return "(none)"
	case Locked:
		return "locked"
	case LockedByUpdate:
		return "locked-by-update"
	default:
		return "unknown"

	}
}

type VolumeInfo struct {
	VolumeCreateRequest
	Id       string `json:"id"`
	Remvolid string `json:"remvolid"`
	Cluster  string `json:"cluster"`
	Mount    struct {
		GlusterFS struct {
			Hosts      []string          `json:"hosts"`
			MountPoint string            `json:"device"`
			Options    map[string]string `json:"options"`
		} `json:"glusterfs"`
	} `json:"mount"`
	BlockInfo struct {
		FreeSize     int              `json:"freesize,omitempty"`
		ReservedSize int              `json:"reservedsize,omitempty"`
		BlockVolumes sort.StringSlice `json:"blockvolume,omitempty"`
		Restriction  BlockRestriction `json:"restriction,omitempty"`
	} `json:"blockinfo,omitempty"`
}

type VolumeInfoResponse struct {
	VolumeInfo
	Bricks []BrickInfo `json:"bricks"`
}

type VolumeListResponse struct {
	Volumes []string `json:"volumes"`
}

type VolumeExpandRequest struct {
	Size int `json:"expand_size"`
}

func (volExpandReq VolumeExpandRequest) Validate() error {
	return validation.ValidateStruct(&volExpandReq,
		validation.Field(&volExpandReq.Size, validation.Required, validation.Min(1)),
	)
}

type VolumeCloneRequest struct {
	Name string `json:"name,omitempty"`
}

func (vcr VolumeCloneRequest) Validate() error {
	return validation.ValidateStruct(&vcr,
		validation.Field(&vcr.Name, validation.Match(volumeNameRe)),
	)
}

type VolumeSnapshotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func (vscr VolumeSnapshotRequest) Validate() error {
	return nil
}

type SnapshotListResponse struct {
	Snapshots []string `json:"snapshots"`
}

type SnapshotDeleteRequest struct {
}

type SnapshotInfoResponse struct {
	SnapshotInfo
}

type VolumeBlockRestrictionRequest struct {
	Restriction BlockRestriction `json:"restriction"`
}

func (vbrr VolumeBlockRestrictionRequest) Validate() error {
	return validation.ValidateStruct(&vbrr,
		validation.Field(&vbrr.Restriction,
			validation.In(Unrestricted, Locked)))
}

// BlockVolume

type BlockVolumeCreateRequest struct {
	// Size in GiB
	Size     int      `json:"size"`
	Clusters []string `json:"clusters,omitempty"`
	Name     string   `json:"name"`
	Hacount  int      `json:"hacount,omitempty"`
	Auth     bool     `json:"auth,omitempty"`
}

func (blockVolCreateReq BlockVolumeCreateRequest) Validate() error {
	return validation.ValidateStruct(&blockVolCreateReq,
		validation.Field(&blockVolCreateReq.Size, validation.Required, validation.Min(1)),
		validation.Field(&blockVolCreateReq.Clusters, validation.By(ValidateUUID)),
		validation.Field(&blockVolCreateReq.Name, validation.Match(blockVolNameRe)),
		validation.Field(&blockVolCreateReq.Hacount, validation.Min(1)),
		validation.Field(&blockVolCreateReq.Auth, validation.Skip),
	)
}

type BlockVolumeInfo struct {
	BlockVolumeCreateRequest
	Id          string `json:"id"`
	BlockVolume struct {
		Hosts    []string `json:"hosts"`
		Iqn      string   `json:"iqn"`
		Lun      int      `json:"lun"`
		Username string   `json:"username"`
		Password string   `json:"password"`
		/*
			Options   map[string]string `json:"options"`  // needed?...
		*/
	} `json:"blockvolume"`
	Cluster            string `json:"cluster,omitempty"`
	BlockHostingVolume string `json:"blockhostingvolume,omitempty"`
}

type BlockVolumeInfoResponse struct {
	BlockVolumeInfo
}

type BlockVolumeListResponse struct {
	BlockVolumes []string `json:"blockvolumes"`
}

type LogLevelInfo struct {
	// should contain one or more logger to log-level-name mapping
	LogLevel map[string]string `json:"loglevel"`
}

type TagsChangeType string

const (
	UnknownTagsChangeType TagsChangeType = ""
	SetTags               TagsChangeType = "set"
	UpdateTags            TagsChangeType = "update"
	DeleteTags            TagsChangeType = "delete"
)

// Common tag post body
type TagsChangeRequest struct {
	Tags   map[string]string `json:"tags"`
	Change TagsChangeType    `json:"change_type"`
}

func (tcr TagsChangeRequest) Validate() error {
	return validation.ValidateStruct(&tcr,
		validation.Field(&tcr.Tags, validation.By(ValidateTags)),
		validation.Field(&tcr.Change,
			validation.Required,
			validation.In(SetTags, UpdateTags, DeleteTags)))
}

func ValidateTags(v interface{}) error {
	t, ok := v.(map[string]string)
	if !ok {
		return fmt.Errorf("tags must be a map of strings to strings")
	}
	if len(t) > 32 {
		return fmt.Errorf("too many tags specified (%v), up to %v supported",
			len(t), 32)
	}
	for k, v := range t {
		if len(k) == 0 {
			return fmt.Errorf("tag names may not be empty")
		}
		if err := validation.Validate(k, validation.RuneLength(1, 32)); err != nil {
			return fmt.Errorf("tag name %v: %v", k, err)
		}
		if err := validation.Validate(v, validation.RuneLength(0, 64)); err != nil {
			return fmt.Errorf("value of tag %v: %v", k, err)
		}
		if !tagNameRe.MatchString(k) {
			return fmt.Errorf("invalid characters in tag name %+v", k)
		}
	}
	return nil
}

// GeoReplicationActionType defines the different actions relevant to geo-rep sessions, except for delete
type GeoReplicationActionType string

// Supported GeoReplication action types
const (
	GeoReplicationActionCreate GeoReplicationActionType = "create"
	GeoReplicationActionConfig GeoReplicationActionType = "config"
	GeoReplicationActionStart  GeoReplicationActionType = "start"
	GeoReplicationActionStop   GeoReplicationActionType = "stop"
	GeoReplicationActionStatus GeoReplicationActionType = "status"
	GeoReplicationActionPause  GeoReplicationActionType = "pause"
	GeoReplicationActionResume GeoReplicationActionType = "resume"
	GeoReplicationActionDelete GeoReplicationActionType = "delete"
)

type GeoReplicationStatus struct {
	Volumes []GeoReplicationVolume `json:"volume"`
}
type GeoReplicationVolume struct {
	VolumeName string                 `json:"name"`
	Sessions   GeoReplicationSessions `json:"sessions"`
}

type GeoReplicationSessions struct {
	SessionList []GeoReplicationSession `json:"session"`
}

type GeoReplicationSession struct {
	SessionSlave string               `json:"session_slave"`
	Pairs        []GeoReplicationPair `json:"pair"`
}
type GeoReplicationPair struct {
	MasterNode               string `json:"master_node"`
	MasterBrick              string `json:"master_brick"`
	SlaveUser                string `json:"slave_user"`
	Slave                    string `json:"slave"`
	SlaveNode                string `json:"slave_node"`
	Status                   string `json:"status"`
	CrawlStatus              string `json:"crawl_status"`
	Entry                    string `json:"entry"`
	Data                     string `json:"data"`
	Meta                     string `json:"meta"`
	Failures                 string `json:"failures"`
	CheckpointCompleted      string `json:"checkpoint_completed"`
	MasterNodeUUID           string `json:"master_node_uuid"`
	LastSynced               string `json:"last_synced"`
	CheckpointTime           string `json:"checkpoint_time"`
	CheckpointCompletionTime string `json:"checkpoint_completion_time"`
}

type GeoReplicationInfo struct {
	SlaveHost    string `json:"slavehost"`
	SlaveVolume  string `json:"slavevolume"`
	SlaveSSHPort int    `json:"slavesshport"`
}

//VolumeGeoReplicationRequest is the body for a GeoReplication POST request
type GeoReplicationRequest struct {
	Action       GeoReplicationActionType `json:"action"`
	ActionParams map[string]string        `json:"actionparams,omitempty"`
	GeoReplicationInfo
}

// MsaterSlave
type MasterSlaveCluster struct {
	Remoteid string `json:"remoteid"`
	Status   string `json:"status"`
}

type ClusterSetMasterSlaveRequest struct {
	MasterSlaveCluster
}

type MasterSlaveClusterStatus struct {
	MasterSlaveCluster
}

func (v *MasterSlaveClusterStatus) String() string {
	var s string
	s = fmt.Sprintf("Cluster remote id: %v\n"+
		"Cluster status: %v\n",
		v.Remoteid,
		v.Status)
	return s
}

type MasterSlaveVolpair struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Remvolid string `json:"remvolid"`
}

type MasterSlaveStatus struct {
	Id       string               `json:"id"`
	Remoteid string               `json:"remoteid"`
	Status   string               `json:"status"`
	Volumes  []MasterSlaveVolpair `json:"volumes"`
	Side     string               `json:"side,omitempty"`
}

func (v *MasterSlaveStatus) String() string {
	var s string
	var vols string
	s = fmt.Sprintf("Master Cluster Id: %v\n"+
		"Slave Cluster Id: %v\n"+
		"Cluster Status: %v\n"+
		"Master Cluster Side: %v\n",
		v.Id,
		v.Remoteid,
		v.Status,
		v.Side) +
		fmt.Sprintf("Volumes: \n")

	for _, vol := range v.Volumes {
		vols += fmt.Sprintf("\tVolume id: %v\n"+
			"\tVolume Name: %v\n"+
			"\tVolume remote id: %v\n \n",
			vol.Id,
			vol.Name,
			vol.Remvolid)
	}

	return s + vols
}

type SnapshotInfo struct {
	originVolume VolumeInfo
	VolumeSnapshotRequest
	Id string `json:"id"`
}

// Constructors

func NewVolumeInfoResponse() *VolumeInfoResponse {

	info := &VolumeInfoResponse{}
	info.Mount.GlusterFS.Options = make(map[string]string)
	info.Bricks = make([]BrickInfo, 0)

	return info
}

func (v *GeoReplicationStatus) String() string {
	var s string
	for _, vol := range v.Volumes {
		s += fmt.Sprintf("Master Volume Name: %v\n"+
			"Session slave: %v\n",
			vol.VolumeName,
			vol.Sessions.SessionList[0].SessionSlave)

		s += "Pairs:\n"
		for _, p := range vol.Sessions.SessionList[0].Pairs {
			s += fmt.Sprintf("\tMaster Node: %s\n"+
				"\tMaster Brick: %s\n"+
				"\tSlave User: %s\n"+
				"\tSlave: %s\n"+
				"\tSlave Node: %s\n"+
				"\tStatus: %s\n"+
				"\tCrawl Status: %s\n"+
				"\tEntry: %s\n"+
				"\tData: %s\n"+
				"\tMeta: %s\n"+
				"\tFailures: %s\n"+
				"\tCheckpoint Completed: %s\n"+
				"\tMaster Node UUID: %s\n"+
				"\tLast Synced: %s\n"+
				"\tCheckpoint Time: %s\n"+
				"\tCheckpoint Completion Time: %s\n\n",
				p.MasterNode,
				p.MasterBrick,
				p.SlaveUser,
				p.Slave,
				p.SlaveNode,
				p.Status,
				p.CrawlStatus,
				p.Entry,
				p.Data,
				p.Meta,
				p.Failures,
				p.CheckpointCompleted,
				p.MasterNodeUUID,
				p.LastSynced,
				p.CheckpointTime,
				p.CheckpointCompletionTime,
			)
		}
	}

	return s
}

// String functions
func (v *VolumeInfoResponse) String() string {
	s := fmt.Sprintf("Name: %v\n"+
		"Size: %v\n"+
		"Volume Id: %v\n"+
		"Cluster Id: %v\n"+
		"Mount: %v\n"+
		"Mount Options: backup-volfile-servers=%v\n"+
		"Block: %v\n"+
		"Free Size: %v\n"+
		"Reserved Size: %v\n"+
		"Block Hosting Restriction: %v\n"+
		"Block Volumes: %v\n"+
		"Durability Type: %v\n"+
		"Remote Volume Id: %v\n",
		v.Name,
		v.Size,
		v.Id,
		v.Cluster,
		v.Mount.GlusterFS.MountPoint,
		v.Mount.GlusterFS.Options["backup-volfile-servers"],
		v.Block,
		v.BlockInfo.FreeSize,
		v.BlockInfo.ReservedSize,
		v.BlockInfo.Restriction,
		v.BlockInfo.BlockVolumes,
		v.Durability.Type,
		v.Remvolid)

	switch v.Durability.Type {
	case DurabilityEC:
		s += fmt.Sprintf("Disperse Data: %v\n"+
			"Disperse Redundancy: %v\n",
			v.Durability.Disperse.Data,
			v.Durability.Disperse.Redundancy)
	case DurabilityReplicate:
		s += fmt.Sprintf("Distributed+Replica: %v\n",
			v.Durability.Replicate.Replica)
	}

	if v.Snapshot.Enable {
		s += fmt.Sprintf("Snapshot Factor: %.2f\n",
			v.Snapshot.Factor)
	}

	/*
		s += "\nBricks:\n"
		for _, b := range v.Bricks {
			s += fmt.Sprintf("Id: %v\n"+
				"Path: %v\n"+
				"Size (GiB): %v\n"+
				"Node: %v\n"+
				"Device: %v\n\n",
				b.Id,
				b.Path,
				b.Size/(1024*1024),
				b.NodeId,
				b.DeviceId)
		}
	*/

	return s
}

func NewBlockVolumeInfoResponse() *BlockVolumeInfoResponse {

	info := &BlockVolumeInfoResponse{}
	// Nothing to Construct now maybe for future

	return info
}

// String functions
func (v *BlockVolumeInfoResponse) String() string {
	s := fmt.Sprintf("Name: %v\n"+
		"Size: %v\n"+
		"Volume Id: %v\n"+
		"Cluster Id: %v\n"+
		"Hosts: %v\n"+
		"IQN: %v\n"+
		"LUN: %v\n"+
		"Hacount: %v\n"+
		"Username: %v\n"+
		"Password: %v\n"+
		"Block Hosting Volume: %v\n",
		v.Name,
		v.Size,
		v.Id,
		v.Cluster,
		v.BlockVolume.Hosts,
		v.BlockVolume.Iqn,
		v.BlockVolume.Lun,
		v.Hacount,
		v.BlockVolume.Username,
		v.BlockVolume.Password,
		v.BlockHostingVolume)

	/*
		s += "\nBricks:\n"
		for _, b := range v.Bricks {
			s += fmt.Sprintf("Id: %v\n"+
				"Path: %v\n"+
				"Size (GiB): %v\n"+
				"Node: %v\n"+
				"Device: %v\n\n",
				b.Id,
				b.Path,
				b.Size/(1024*1024),
				b.NodeId,
				b.DeviceId)
		}
	*/

	return s
}

type OperationsInfo struct {
	Total    uint64 `json:"total"`
	InFlight uint64 `json:"in_flight"`
	// state based counts:
	Stale uint64 `json:"stale"`
	New   uint64 `json:"new"`
}

type AdminState string

const (
	AdminStateNormal   AdminState = "normal"
	AdminStateReadOnly AdminState = "read-only"
	AdminStateLocal    AdminState = "local-client"
)

type AdminStatus struct {
	State AdminState `json:"state"`
}

func (as AdminStatus) Validate() error {
	return validation.ValidateStruct(&as,
		validation.Field(&as.State,
			validation.Required,
			validation.In(AdminStateNormal, AdminStateReadOnly, AdminStateLocal)))
}
