//
// (C) Copyright 2019-2024 Intel Corporation.
// (C) Copyright 2025 Hewlett Packard Enterprise Development LP
//
// SPDX-License-Identifier: BSD-2-Clause-Patent
//

// This file imports all of the DAOS dRPC module/method IDs.

package drpc

import (
	fmt "fmt"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// #cgo CFLAGS: -I${SRCDIR}/../../include
// #include <daos/drpc_modules.h>
import "C"

const moduleMethodOffset = 100

type ModuleID int32

func (id ModuleID) String() string {
	if name, ok := map[ModuleID]string{
		ModuleSecurityAgent: "Agent Security",
		ModuleMgmt:          "Management",
		ModuleSrv:           "Server",
		ModuleSecurity:      "Security",
	}[id]; ok {
		return name
	}

	return fmt.Sprintf("unknown module id: %d", id)
}

func (id ModuleID) GetMethod(methodID int32) (Method, error) {
	if m, ok := map[ModuleID]Method{
		ModuleSecurityAgent: securityAgentMethod(methodID),
		ModuleMgmt:          MgmtMethod(methodID),
		ModuleSrv:           srvMethod(methodID),
		ModuleSecurity:      securityMethod(methodID),
	}[id]; ok {
		if !m.IsValid() {
			return nil, errors.Errorf("invalid method %d for module %s",
				methodID, id)
		}
		return m, nil
	}

	return nil, errors.Errorf("unknown module id %d", id)
}

func (id ModuleID) ID() int32 {
	return int32(id)
}

const (
	// ModuleSecurityAgent is the dRPC module for security tasks in DAOS agent
	ModuleSecurityAgent ModuleID = C.DRPC_MODULE_SEC_AGENT
	// ModuleMgmt is the dRPC module for management service tasks
	ModuleMgmt ModuleID = C.DRPC_MODULE_MGMT
	// ModuleSrv is the dRPC module for tasks relating to server setup
	ModuleSrv ModuleID = C.DRPC_MODULE_SRV
	// ModuleSecurity is the dRPC module for security tasks in DAOS server
	ModuleSecurity ModuleID = C.DRPC_MODULE_SEC
)

type Method interface {
	ID() int32
	Module() ModuleID
	String() string
	IsValid() bool
}

type securityAgentMethod int32

func (m securityAgentMethod) Module() ModuleID {
	return ModuleSecurityAgent
}

func (m securityAgentMethod) ID() int32 {
	return int32(m)
}

func (m securityAgentMethod) String() string {
	if s, ok := map[securityAgentMethod]string{
		MethodRequestCredentials: "request agent credentials",
	}[m]; ok {
		return s
	}

	return fmt.Sprintf("%s:%d", m.Module(), m.ID())
}

// IsValid sanity checks the Method ID is within expected bounds.
func (m securityAgentMethod) IsValid() bool {
	startMethodID := int32(m.Module()) * moduleMethodOffset

	if m.ID() <= startMethodID || m.ID() >= int32(C.NUM_DRPC_SEC_AGENT_METHODS) {
		return false
	}

	return true
}

const (
	// MethodRequestCredentials is a ModuleSecurityAgent method
	MethodRequestCredentials securityAgentMethod = C.DRPC_METHOD_SEC_AGENT_REQUEST_CREDS
)

type MgmtMethod int32

func (m MgmtMethod) Module() ModuleID {
	return ModuleMgmt
}

func (m MgmtMethod) ID() int32 {
	return int32(m)
}

func (m MgmtMethod) String() string {
	if s, ok := map[MgmtMethod]string{
		MethodPrepShutdown:         "PrepShutdown",
		MethodPingRank:             "PingRank",
		MethodSetRank:              "SetRank",
		MethodSetLogMasks:          "SetLogMasks",
		MethodGetAttachInfo:        "GetAttachInfo",
		MethodPoolCreate:           "PoolCreate",
		MethodPoolDestroy:          "PoolDestroy",
		MethodPoolEvict:            "PoolEvict",
		MethodPoolExclude:          "PoolExclude",
		MethodPoolDrain:            "PoolDrain",
		MethodPoolExtend:           "PoolExtend",
		MethodPoolReintegrate:      "PoolReintegrate",
		MethodBioHealth:            "BioHealth",
		MethodSetUp:                "SetUp",
		MethodSmdDevs:              "SmdDevs",
		MethodSmdPools:             "SmdPools",
		MethodPoolGetACL:           "PoolGetACL",
		MethodPoolOverwriteACL:     "PoolOverwriteACL",
		MethodPoolUpdateACL:        "PoolUpdateACL",
		MethodPoolDeleteACL:        "PoolDeleteACL",
		MethodSetFaultyState:       "SetFaultyState",
		MethodReplaceStorage:       "ReplaceStorage",
		MethodListContainers:       "ListContainers",
		MethodPoolQuery:            "PoolQuery",
		MethodPoolQueryTarget:      "PoolQueryTarget",
		MethodPoolSetProp:          "PoolSetProp",
		MethodContSetOwner:         "ContSetOwner",
		MethodGroupUpdate:          "GroupUpdate",
		MethodNotifyPoolConnect:    "NotifyPoolConnect",
		MethodNotifyPoolDisconnect: "NotifyPoolDisconnect",
		MethodNotifyExit:           "NotifyExit",
		MethodPoolGetProp:          "PoolGetProp",
		MethodPoolUpgrade:          "PoolUpgrade",
		MethodLedManage:            "LedManage",
		MethodSetupClientTelemetry: "SetupClientTelemetry",
	}[m]; ok {
		return s
	}

	return fmt.Sprintf("%s:%d", m.Module(), m.ID())
}

// IsValid sanity checks the Method ID is within expected bounds.
func (m MgmtMethod) IsValid() bool {
	startMethodID := int32(m.Module()) * moduleMethodOffset

	if m.ID() <= startMethodID || m.ID() >= int32(C.NUM_DRPC_MGMT_METHODS) {
		return false
	}

	return true
}

const (
	// MethodPrepShutdown is a ModuleMgmt method
	MethodPrepShutdown MgmtMethod = C.DRPC_METHOD_MGMT_PREP_SHUTDOWN
	// MethodPingRank is a ModuleMgmt method
	MethodPingRank MgmtMethod = C.DRPC_METHOD_MGMT_PING_RANK
	// MethodSetRank is a ModuleMgmt method
	MethodSetRank MgmtMethod = C.DRPC_METHOD_MGMT_SET_RANK
	// MethodSetLogMasks is a ModuleMgmt method
	MethodSetLogMasks MgmtMethod = C.DRPC_METHOD_MGMT_SET_LOG_MASKS
	// MethodGetAttachInfo is a ModuleMgmt method
	MethodGetAttachInfo MgmtMethod = C.DRPC_METHOD_MGMT_GET_ATTACH_INFO
	// MethodPoolCreate is a ModuleMgmt method
	MethodPoolCreate MgmtMethod = C.DRPC_METHOD_MGMT_POOL_CREATE
	// MethodPoolDestroy is a ModuleMgmt method
	MethodPoolDestroy MgmtMethod = C.DRPC_METHOD_MGMT_POOL_DESTROY
	// MethodPoolEvict is a ModuleMgmt method to evict pool connections
	MethodPoolEvict MgmtMethod = C.DRPC_METHOD_MGMT_POOL_EVICT
	// MethodPoolExclude is a ModuleMgmt method for excluding pool ranks
	MethodPoolExclude MgmtMethod = C.DRPC_METHOD_MGMT_POOL_EXCLUDE
	// MethodPoolDrain is a ModuleMgmt method for draining pool ranks
	MethodPoolDrain MgmtMethod = C.DRPC_METHOD_MGMT_POOL_DRAIN
	// MethodPoolReintegrate is a ModuleMgmt method for reintegrating pool ranks
	MethodPoolReintegrate MgmtMethod = C.DRPC_METHOD_MGMT_POOL_REINT
	// MethodPoolExtend is a ModuleMgmt method for extending pool
	MethodPoolExtend MgmtMethod = C.DRPC_METHOD_MGMT_POOL_EXTEND
	// MethodBioHealth is a ModuleMgmt method
	MethodBioHealth MgmtMethod = C.DRPC_METHOD_MGMT_BIO_HEALTH_QUERY
	// MethodSetUp is a ModuleMgmt method
	MethodSetUp MgmtMethod = C.DRPC_METHOD_MGMT_SET_UP
	// MethodSmdDevs is a ModuleMgmt method
	MethodSmdDevs MgmtMethod = C.DRPC_METHOD_MGMT_SMD_LIST_DEVS
	// MethodSmdPools is a ModuleMgmt method
	MethodSmdPools MgmtMethod = C.DRPC_METHOD_MGMT_SMD_LIST_POOLS
	// MethodPoolGetACL is a ModuleMgmt method
	MethodPoolGetACL MgmtMethod = C.DRPC_METHOD_MGMT_POOL_GET_ACL
	// MethodPoolOverwriteACL is a ModuleMgmt method
	MethodPoolOverwriteACL MgmtMethod = C.DRPC_METHOD_MGMT_POOL_OVERWRITE_ACL
	// MethodPoolUpdateACL is a ModuleMgmt method
	MethodPoolUpdateACL MgmtMethod = C.DRPC_METHOD_MGMT_POOL_UPDATE_ACL
	// MethodPoolDeleteACL is a ModuleMgmt method
	MethodPoolDeleteACL MgmtMethod = C.DRPC_METHOD_MGMT_POOL_DELETE_ACL
	// MethodSetFaultyState is a ModuleMgmt method
	MethodSetFaultyState MgmtMethod = C.DRPC_METHOD_MGMT_DEV_SET_FAULTY
	// MethodReplaceStorage is a ModuleMgmt method
	MethodReplaceStorage MgmtMethod = C.DRPC_METHOD_MGMT_DEV_REPLACE
	// MethodListContainers is a ModuleMgmt method
	MethodListContainers MgmtMethod = C.DRPC_METHOD_MGMT_LIST_CONTAINERS
	// MethodPoolQuery defines a method for querying a pool
	MethodPoolQuery MgmtMethod = C.DRPC_METHOD_MGMT_POOL_QUERY
	// MethodPoolQueryTarget defines a method for querying a pool engine's targets
	MethodPoolQueryTarget MgmtMethod = C.DRPC_METHOD_MGMT_POOL_QUERY_TARGETS
	// MethodPoolSetProp defines a method for setting a pool property
	MethodPoolSetProp MgmtMethod = C.DRPC_METHOD_MGMT_POOL_SET_PROP
	// MethodContSetOwner defines a method for setting the container's owner
	MethodContSetOwner MgmtMethod = C.DRPC_METHOD_MGMT_CONT_SET_OWNER
	// MethodGroupUpdate defines a method for updating the group map
	MethodGroupUpdate MgmtMethod = C.DRPC_METHOD_MGMT_GROUP_UPDATE
	// MethodNotifyPoolConnect defines a method to indicate a successful pool connect call
	MethodNotifyPoolConnect MgmtMethod = C.DRPC_METHOD_MGMT_NOTIFY_POOL_CONNECT
	// MethodNotifyPoolDisconnect defines a method to indicate a successful pool disconnect call
	MethodNotifyPoolDisconnect MgmtMethod = C.DRPC_METHOD_MGMT_NOTIFY_POOL_DISCONNECT
	// MethodNotifyExit defines a method for signaling a clean client shutdown
	MethodNotifyExit MgmtMethod = C.DRPC_METHOD_MGMT_NOTIFY_EXIT
	// MethodPoolGetProp defines a method for getting pool properties
	MethodPoolGetProp MgmtMethod = C.DRPC_METHOD_MGMT_POOL_GET_PROP
	// MethodCheckerStart defines a method for starting the checker
	MethodCheckerStart MgmtMethod = C.DRPC_METHOD_MGMT_CHK_START
	// MethodCheckerStop defines a method for stopping the checker
	MethodCheckerStop MgmtMethod = C.DRPC_METHOD_MGMT_CHK_STOP
	// MethodCheckerQuery defines a method for getting the checker status
	MethodCheckerQuery MgmtMethod = C.DRPC_METHOD_MGMT_CHK_QUERY
	// MethodCheckerProp defines a method for getting the checker properties
	MethodCheckerProp MgmtMethod = C.DRPC_METHOD_MGMT_CHK_PROP
	// MethodCheckerAction defines a method for specifying a checker action
	MethodCheckerAction MgmtMethod = C.DRPC_METHOD_MGMT_CHK_ACT
	// MethodPoolUpgrade defines a method for upgrade pool
	MethodPoolUpgrade MgmtMethod = C.DRPC_METHOD_MGMT_POOL_UPGRADE
	// MethodLedManage defines a method to manage a VMD device LED state
	MethodLedManage MgmtMethod = C.DRPC_METHOD_MGMT_LED_MANAGE
	// MethodSetupClientTelemetry defines a method to setup client telemetry
	MethodSetupClientTelemetry MgmtMethod = C.DRPC_METHOD_MGMT_SETUP_CLIENT_TELEM
)

type srvMethod int32

func (m srvMethod) Module() ModuleID {
	return ModuleSrv
}

func (m srvMethod) ID() int32 {
	return int32(m)
}

func (m srvMethod) String() string {
	if s, ok := map[srvMethod]string{
		MethodNotifyReady:         "notify ready",
		MethodClusterEvent:        "cluster event",
		MethodGetPoolServiceRanks: "get pool service ranks",
		MethodPoolFindByLabel:     "find pool by label",
		MethodListPools:           "list pools",
	}[m]; ok {
		return s
	}

	return fmt.Sprintf("%s:%d", m.Module(), m.ID())
}

// IsValid sanity checks the Method ID is within expected bounds.
func (m srvMethod) IsValid() bool {
	startMethodID := int32(m.Module()) * moduleMethodOffset

	if m.ID() <= startMethodID || m.ID() >= int32(C.NUM_DRPC_SRV_METHODS) {
		return false
	}

	return true
}

const (
	// MethodNotifyReady is a ModuleSrv method
	MethodNotifyReady srvMethod = C.DRPC_METHOD_SRV_NOTIFY_READY
	// MethodGetPoolServiceRanks requests the service ranks for a pool
	MethodGetPoolServiceRanks srvMethod = C.DRPC_METHOD_SRV_GET_POOL_SVC
	// MethodPoolFindByLabel requests the service ranks and UUID for a pool
	MethodPoolFindByLabel srvMethod = C.DRPC_METHOD_SRV_POOL_FIND_BYLABEL
	// MethodClusterEvent notifies of a cluster event in the I/O Engine.
	MethodClusterEvent srvMethod = C.DRPC_METHOD_SRV_CLUSTER_EVENT
	// MethodCheckerListPools requests the list of pools from the MS
	MethodCheckerListPools srvMethod = C.DRPC_METHOD_CHK_LIST_POOL // TODO (DAOS-16126): Merge with MethodListPools
	// MethodCheckerRegisterPool registers a pool with the MS
	MethodCheckerRegisterPool srvMethod = C.DRPC_METHOD_CHK_REG_POOL
	// MethodCheckerDeregisterPool deregisters a pool with the MS
	MethodCheckerDeregisterPool srvMethod = C.DRPC_METHOD_CHK_DEREG_POOL
	// MethodCheckerReport reports a checker finding to the MS
	MethodCheckerReport srvMethod = C.DRPC_METHOD_CHK_REPORT
	// MethodListPools requests the list of pools in the system
	MethodListPools srvMethod = C.DRPC_METHOD_SRV_LIST_POOLS
)

type securityMethod int32

func (m securityMethod) Module() ModuleID {
	return ModuleSecurity
}

func (m securityMethod) ID() int32 {
	return int32(m)
}

func (m securityMethod) String() string {
	if s, ok := map[securityMethod]string{
		MethodValidateCredentials: "validate credentials",
	}[m]; ok {
		return s
	}

	return fmt.Sprintf("%s:%d", m.Module(), m.ID())
}

// IsValid sanity checks the Method ID is within expected bounds.
func (m securityMethod) IsValid() bool {
	startMethodID := int32(m.Module()) * moduleMethodOffset

	if m.ID() <= startMethodID || m.ID() >= int32(C.NUM_DRPC_SEC_METHODS) {
		return false
	}

	return true
}

const (
	// MethodValidateCredentials is a ModuleSecurity method
	MethodValidateCredentials securityMethod = C.DRPC_METHOD_SEC_VALIDATE_CREDS
)

// Marshal is a utility function that can be used by dRPC method handlers to
// marshal their method-specific response to be passed back to the ModuleService.
func Marshal(message proto.Message) ([]byte, error) {
	msgBytes, err := proto.Marshal(message)
	if err != nil {
		return nil, MarshalingFailure()
	}
	return msgBytes, nil
}
