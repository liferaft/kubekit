package kluster

import (
	"strings"
)

// Status is used to name the cluster status
type Status int

// All the possible values of a cluster status
// Pending 			: The cluster is not created/provisioned yet
// Provisioned  : The cluster is provisioned and ready to be configured
// Running 			: The cluster has a Kubernetes cluster running
// Stopped  		: The cluster has a Kubernetes cluster but it's not running. It can go to the Running or Terminated status
// Terminated 	: The cluster is destroyed, it does not exists anymore
const (
	AbsentStatus              Status = 1 << iota // 00000000001 : Does not exists, not created/provisioned yet
	CreatingStatus                               // 00000000010 : It's been created, it's been provisioned
	ProvisionedStatus                            // 00000000100 : The provisioning was successfully completed and ready to be configured
	FailedProvisioningStatus                     // 00000001000 : The provisioning failed, some cluster nodes may noe exists
	FailedConfigurationStatus                    // 00000010000 : The configuration started and failed
	FailedCreationStatus                         // 						: The cluster failed to be created
	CreatedStatus                                // 00000100000 : The configuration was successfully completed
	RunningStatus                                // 00001000000 : The cluster has a Kubernetes cluster up & running
	StoppedStatus                                // 00010000000 : The cluster has a Kubernetes cluster but it's not running. It can go to the Running or Terminated status
	TerminatingStatus                            // 00100000000 : It's in the process to be destroyed
	TerminatedStatus                             // 01000000000 : The cluster is destroyed, it does not exists anymore
	FailedTerminationStatus                      // 10000000000 : The termination process failed
	UnknownStatus
)

// AllStatuses contain all the statuses in one variable
var AllStatuses = []Status{
	AbsentStatus,
	CreatingStatus,
	ProvisionedStatus,
	FailedProvisioningStatus,
	FailedConfigurationStatus,
	FailedCreationStatus,
	CreatedStatus,
	RunningStatus,
	StoppedStatus,
	TerminatingStatus,
	TerminatedStatus,
	FailedTerminationStatus,
}

// String returns the name of the received status
func (status Status) String() string {
	switch status {
	case AbsentStatus:
		return "absent"
	case CreatingStatus:
		return "creating"
	case ProvisionedStatus:
		return "provisioned"
	case FailedProvisioningStatus:
		return "failed to provision"
	case FailedConfigurationStatus:
		return "failed to configure"
	case FailedCreationStatus:
		return "failed to create"
	case CreatedStatus:
		return "created"
	case RunningStatus:
		return "running"
	case StoppedStatus:
		return "stopped"
	case TerminatingStatus:
		return "terminating"
	case TerminatedStatus:
		return "terminated"
	case FailedTerminationStatus:
		return "failed to terminate"
	default:
		return "unknown"
	}

}

// ParseStatus returns the status from a status name
func ParseStatus(status string) Status {
	switch strings.ToLower(status) {
	case "absent":
		return AbsentStatus
	case "creating":
		return CreatingStatus
	case "provisioned":
		return ProvisionedStatus
	case "failed to provision":
		return FailedProvisioningStatus
	case "failed to configure":
		return FailedConfigurationStatus
	case "failed to create":
		return FailedCreationStatus
	case "created":
		return CreatedStatus
	case "running":
		return RunningStatus
	case "stopped":
		return StoppedStatus
	case "terminating":
		return TerminatingStatus
	case "terminated":
		return TerminatedStatus
	case "failed to terminate":
		return FailedTerminationStatus
	default:
		return UnknownStatus
	}
}
