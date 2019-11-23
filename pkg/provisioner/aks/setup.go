package aks

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2015-12-01/features"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/terraform/states"

	"github.com/kubekit/azure"
	"github.com/kubekit/kubekit/pkg/configurator/ssh"
	"github.com/kubekit/kubekit/pkg/provisioner/state"
)

const (
	numRetries      = 30
	intervalSeconds = 60

	jumpboxNumRetries      = 15
	jumpboxIntervalSeconds = 20

	msAzureExtensionsPublisher = "Microsoft.Azure.Extensions"
	customScriptExtension      = "CustomScript"
	vmssCustomScriptExtension  = "vmssCSE"
	vmssNamePrefix             = "aks"
	vmssNameSuffix             = "vmss"

	errRegisterFeatureFmt = "unable to register feature: %s"
)

func (p *Platform) registerProvider(session *azure.Session, resoureProviderNamespace string) error {
	providersClient, err := azure.ProvidersClientByEnvStr(p.config.Environment, session)
	if err != nil {
		return fmt.Errorf("issues connecting to Azure via ProvidersClient: %s", err)
	}

	p.ui.Log.Infof("Registering provider %s.", resoureProviderNamespace)
	registerResult, err := azure.RegisterProvider(providersClient, resoureProviderNamespace)
	if err != nil {
		return fmt.Errorf("failed to register provider %s: %s", resoureProviderNamespace, err)
	} else if *registerResult.RegistrationState != azure.ProviderRegisteredState {
		isRegistered, err := azure.IsProviderRegistered(providersClient, resoureProviderNamespace)
		if err != nil {
			return err
		}
		for i := 0; i < numRetries && !isRegistered; i++ {
			p.ui.Log.Infof("Provider is still not registered, checking again in %d seconds [%d retries left]", intervalSeconds, numRetries-i)
			time.Sleep(time.Duration(intervalSeconds) * time.Second)
			isRegistered, err = azure.IsProviderRegistered(providersClient, resoureProviderNamespace)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// enablePreviewFeature enables the preview feature in Azure
func (p *Platform) enablePreviewFeature(client *features.Client, session *azure.Session, featureNamespace, featureName string) error {
	// register provider if neeeded
	err := p.registerProvider(session, featureNamespace)
	if err != nil {
		return err
	}

	// enable preview feature in Azure since Terraform doesnt provide a way for us to do so
	registerResult, err := azure.RegisterFeature(client, featureNamespace, featureName)
	if err != nil {
		return err
	}
	if *registerResult.Properties.State == azure.FeatureRegisteredState {
		return nil // p.registerProvider(session, featureNamespace)
	}

	p.ui.Log.Infof("Registering feature %s/%s for the first time, this can take 15+ minutes...", featureNamespace, featureName)

	var isRegistered bool
	for i := 0; i < numRetries; i++ {
		time.Sleep(time.Duration(intervalSeconds) * time.Second)
		isRegistered, err = azure.IsFeatureRegistered(client, featureNamespace, featureName)
		if err != nil {
			return err
		}
		if isRegistered {
			return nil // p.registerProvider(session, featureNamespace)
		}

		p.ui.Log.Infof("Feature is still not registered, checking again in %d seconds [%d retries left]", intervalSeconds, numRetries-i)
	}

	return fmt.Errorf(errRegisterFeatureFmt, err)
}

func (p *Platform) setupPreviewFeatures() error {
	if p.config.PreviewFeatures == nil {
		return nil
	}

	authInfo := &azure.AuthInfo{
		SubscriptionID: p.config.SubscriptionID,
		TenantID:       p.config.TenantID,
		ClientID:       p.config.ClientID,
		ClientSecret:   p.config.ClientSecret,
	}

	session, err := azure.NewSession(authInfo, false)
	if err != nil {
		return fmt.Errorf("issues connecting to Azure: %s", err)
	}

	featuresClient, err := azure.FeaturesClientByEnvStr(p.config.Environment, session)
	if err != nil {
		return fmt.Errorf("issues connecting to Azure via Client: %s", err)
	}

	for _, previewFeature := range p.config.PreviewFeatures {
		p.ui.Log.Debugf("Enabling preview feature: %s/%s", previewFeature.Namespace, previewFeature.Name)
		err := p.enablePreviewFeature(featuresClient, session, previewFeature.Namespace, previewFeature.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Platform) setupJumpbox(session *azure.Session, output map[string]*states.OutputValue, nodeResourceGroup string) error {
	if p.config.Jumpbox == nil {
		return nil
	}

	privKey := p.config.Jumpbox.PrivateKey
	privKeyFile := p.config.Jumpbox.PrivateKeyFile

	// if both private key and file is missing from jumpbox definition, then take from the cluster
	if privKey == "" && privKeyFile == "" {
		privKey = p.config.PrivateKey
		privKeyFile = p.config.PrivateKeyFile
	}

	if p.config.Jumpbox != nil && (privKey != "" || privKeyFile != "") {
		nicsClient, err := azure.NicsClientByEnvStr(p.config.Environment, session)
		if err != nil {
			return fmt.Errorf("issues connecting to Azure via Interface Client: %s", err)
		}
		publicIPsClient, err := azure.PublicIPAddressesClientByEnvStr(p.config.Environment, session)
		if err != nil {
			return fmt.Errorf("issues connecting to Azure via Public IP Addresses Client: %s", err)
		}

		if jumpbox, err := state.OutputKeysValueAsString(output, "jumpbox"); err == nil {
			nicsList, err := azure.ListPrimaryIPsInfo(nicsClient, publicIPsClient, p.config.ClusterName)
			if err != nil {
				if strings.Contains(err.Error(), `Code="ResourceGroupNotFound"`) {
					p.ui.Log.Infof("The '%s' resource group was not found. Skipping listing of NICs.", p.config.ClusterName)
				} else {
					return fmt.Errorf("issues retrieving network interface info from Azure via Interface Client: %s", err)
				}
			}

			// filter jumpbox in case there are other VMs in the resource group
			hostIP := jumpbox
			for _, n := range p.getNodeStates(nicsList, "", "jumpbox") {
				if n.PrivateIP == jumpbox {
					if n.PublicIP != "" {
						hostIP = n.PublicIP
					}

					break
				}
			}

			// get public ip and upload private key of aks cluster to vm
			// for some reason the file provisioner in terraform is acting wonky
			// may revisit in the future if i ever remember to
			if privKey == "" {
				privateKey, err := ioutil.ReadFile(privKeyFile)
				if err != nil {
					return fmt.Errorf("failed to read private key from %s: %s", privKeyFile, err)
				}
				// set the private key and file under the cluster config
				p.config.Jumpbox.PrivateKeyFile = privKeyFile
				p.config.Jumpbox.PrivateKey = string(privateKey)
			}

			sshConfig, err := ssh.New(p.config.Jumpbox.AdminUsername, hostIP, p.config.Jumpbox.PrivateKey, "")
			if err != nil {
				return fmt.Errorf("unable to create ssh config: %s", err)
			}
			clusterPrivKey, err := decryptKey(p.config.PrivateKey)
			if err != nil {
				return fmt.Errorf("failed to decrypt key: %s", err)
			}

			readyStatus := "not ready"
			for i := 0; i < jumpboxNumRetries; i++ {
				p.ui.Log.Infof("Waiting for jumpbox to be ready...")
				execOutput, _, _, _ := sshConfig.ExecAndWait("echo ready")
				if execOutput != "" {
					readyStatus = execOutput
					break
				}
				time.Sleep(jumpboxIntervalSeconds * time.Second)
			}

			if readyStatus == "not ready" {
				p.ui.Log.Errorf("Jumpbox readiness: %s", readyStatus)
				return fmt.Errorf("jumpbox did not become ready in the alotted time of %d seconds", jumpboxIntervalSeconds*jumpboxNumRetries)
			}
			p.ui.Log.Infof("Jumpbox readiness: %s", readyStatus)

			// delete previous file since the file is read only chmodded
			// and will throw a permission error if we try to reupload
			sshConfig.Exec(fmt.Sprintf("rm -f /home/%s/.ssh/id_rsa", p.config.Jumpbox.AdminUsername))

			err = sshConfig.CreateFile(fmt.Sprintf("/home/%s/.ssh/id_rsa", p.config.Jumpbox.AdminUsername), clusterPrivKey, 0400)
			if err != nil {
				return fmt.Errorf("failed to upload private key to %s@%s:/home/%s/.ssh/id_rsa: %s", p.config.Jumpbox.AdminUsername, sshConfig.Address, p.config.Jumpbox.AdminUsername, err)
			}

			if p.config.Jumpbox.UploadKubeconfg {
				if kubeconfig, err := state.OutputKeysValueAsString(output, "kubeconfig"); err == nil {
					err = sshConfig.CreateFile(fmt.Sprintf("/home/%s/kubeconfig", p.config.Jumpbox.AdminUsername), kubeconfig, 0644)
					if err != nil {
						return fmt.Errorf("failed to upload kubeconfig to %s@%s:/home/%s/kubeconfig: %s", p.config.Jumpbox.AdminUsername, sshConfig.Address, p.config.Jumpbox.AdminUsername, err)
					}

					sshConfig.Exec(fmt.Sprintf("if ! (grep -q 'export KUBECONFIG=' /home/%s/.bashrc); then echo 'export KUBECONFIG=/home/%s/kubeconfig' >> /home/%s/.bashrc; fi",
						p.config.Jumpbox.AdminUsername, p.config.Jumpbox.AdminUsername, p.config.Jumpbox.AdminUsername))
				}
			}

			if p.config.Jumpbox.Commands != nil {
				for _, c := range p.config.Jumpbox.Commands {
					sshConfig.Exec(c)
				}
			}

			if p.config.Jumpbox.FileUploads != nil {
				for _, f := range p.config.Jumpbox.FileUploads {
					fileUpload, err := ioutil.ReadFile(f)
					if err != nil {
						return fmt.Errorf("failed to read file from %s: %s", f, err)
					}

					err = sshConfig.CreateFile(fmt.Sprintf("/home/%s", p.config.Jumpbox.AdminUsername), string(fileUpload), 0644)
					if err != nil {
						return fmt.Errorf("failed to upload %s to %s@%s:/home/%s: %s", f, p.config.Jumpbox.AdminUsername, sshConfig.Address, p.config.Jumpbox.AdminUsername, err)
					}
				}
			}
		}
	}

	return nil
}

func getAvailabilitySetOfVM(vmID string, availSetVMIDMap map[string][]string) string {
	for k, v := range availSetVMIDMap {
		for _, i := range v {
			if i == vmID {
				return k
			}
		}
	}
	return ""
}

func (p *Platform) setupAvailabilitySets(session *azure.Session, output map[string]*states.OutputValue, nodeResourceGroup string) error {
	csePublicSettingsMap := make(map[string]publicSettings, len(p.config.NodePools))
	vmDataDisksMap := make(map[string][]compute.DataDisk, len(p.config.NodePools))
	timestamp := 1

	// if availability set, add customscript extension to each VM (doenst work on the AvailabilitySet level)
	availSetClient, err := azure.AvailabilitySetClientByEnvStr(p.config.Environment, session)
	if err != nil {
		return fmt.Errorf("issues connecting to Azure via AvailabilitySet Client: %s", err)
	}

	var availSetNames []string
	for availSetName, nodePool := range p.config.NodePools {
		if nodePool.Type == nodePoolTypeAS {
			availSetNames = append(availSetNames, availSetName)
		}
	}

	if len(availSetNames) > 0 {
		vmIDs, err := azure.GetVMIDsFromAvailabilitySets(availSetClient, nodeResourceGroup, availSetNames)
		if err != nil && !strings.Contains(err.Error(), `Code="ResourceGroupNotFound"`) {
			return err
		}

		vmsClient, err := azure.VirtualMachinesClientByEnvStr(p.config.Environment, session)
		if err != nil {
			return fmt.Errorf("issues connecting to Azure via VMs Client: %s", err)
		}

		if len(vmIDs) > 0 {
			var vmIDFilter []string
			for _, availSetVMIDs := range vmIDs {
				vmIDFilter = append(vmIDFilter, availSetVMIDs...)
			}

			vms, err := azure.GetVMs(vmsClient, nodeResourceGroup, vmIDFilter)
			if err != nil {
				return err
			}

			for _, vm := range vms {
				vmAvailSet := getAvailabilitySetOfVM(*vm.ID, vmIDs)
				if len(vmAvailSet) == 0 {
					continue
				}

				nodePool, ok := p.config.NodePools[vmAvailSet]
				if !ok {
					continue
				}

				if nodePool.DataDisks == nil && p.config.DefaultNodePool.DataDisks != nil && len(*p.config.DefaultNodePool.DataDisks) > 0 {
					nodePool.DataDisks = p.config.DefaultNodePool.DataDisks
				}

				var scriptMetadata map[string]string
				if nodePool.DataDisks == nil || len(*nodePool.DataDisks) == 0 {
					scriptMetadata = make(map[string]string, 1)
				} else {
					scriptMetadata = make(map[string]string, len(*nodePool.DataDisks)+1)
					scriptMetadata["NUM_DATA_DISKS"] = strconv.Itoa(len(*nodePool.DataDisks))
				}

				// docker root
				scriptMetadata["DOCKER_ROOT"] = nodePool.DockerRoot

				// ephemeral disk
				if nodePool.EphemeralMountPoint != "" {
					scriptMetadata["EPHEMERAL_MNT_PT"] = nodePool.EphemeralMountPoint
				} else {
					scriptMetadata["EPHEMERAL_MNT_PT"] = p.config.DefaultNodePool.EphemeralMountPoint
				}
				// data disks
				var vmDataDisks []compute.DataDisk
				if nodePool.DataDisks != nil {
					for di, dv := range *nodePool.DataDisks {
						scriptMetadata["DATA_DISK_"+strconv.Itoa(di)+"_MNT_PT"] = dv.MountPoint
						scriptMetadata["DATA_DISK_"+strconv.Itoa(di)+"_VOL_BYTES"] = strconv.Itoa(dv.VolumeSize * (1024 ^ 3)) // GiB to bytes

						dd := compute.DataDisk{
							// Azure: Parameter 'dataDisk.name' is not allowed.
							//Name:                    to.StringPtr("data"),
							Lun:                     to.Int32Ptr(int32(di)), // azure doesnt support many data disks so we should be ok going from int to int32
							Caching:                 compute.CachingTypesNone,
							WriteAcceleratorEnabled: to.BoolPtr(false),
							CreateOption:            compute.DiskCreateOptionTypesEmpty,
							DiskSizeGB:              to.Int32Ptr(int32(dv.VolumeSize)),
						}
						vmDataDisks = append(vmDataDisks, dd)
					}
				}
				if len(vmDataDisks) > 0 {
					vm.StorageProfile.DataDisks = &vmDataDisks
				} else {
					vm.StorageProfile.DataDisks = nil
				}

				// add/update CustomScript extension if there is anything to be applied
				if len(scriptMetadata) == 0 {
					continue
				}

				var scriptMetadataList []string
				for k, v := range scriptMetadata {
					scriptMetadataList = append(scriptMetadataList, fmt.Sprintf("%s='%s'", k, v))
				}

				var buff bytes.Buffer
				gz := gzip.NewWriter(&buff)
				//if _, err = gz.Write([]byte(fmt.Sprintf(PrepareDataDiskScript, strings.Join(scriptMetadataList, "\n")))); err != nil {
				//	return fmt.Errorf("issues compressing prepare data disk script: %s", err)
				//}
				if _, err = gz.Write([]byte(customScript)); err != nil {
					return fmt.Errorf("issues compressing prepare temp disk script: %s", err)
				}

				if err = gz.Flush(); err != nil {
					return fmt.Errorf("issues compressing prepare disk script: %s", err)
				}
				if err = gz.Close(); err != nil {
					return fmt.Errorf("issues compressing prepare disk script: %s", err)
				}
				vmssCSEPublicSettings := publicSettings{
					Script: base64.StdEncoding.EncodeToString(buff.Bytes()),
				}

				if scriptMetadata["EPHEMERAL_MNT_PT"] != "" || len(vmDataDisks) > 0 {
					csePublicSettingsMap[vmAvailSet] = vmssCSEPublicSettings
					vmDataDisksMap[vmAvailSet] = vmDataDisks
				}

				if vm.Resources == nil {
					var extensions []compute.VirtualMachineExtension
					vm.Resources = &extensions
				}

				// check if the CustomScript extension is already in the list
				existingExtIndex := -1
				for i, e := range *vm.Resources {
					if *e.Name == vmssCustomScriptExtension && *e.Publisher == msAzureExtensionsPublisher && *e.Type == customScriptExtension {
						existingExtIndex = i
					}
				}

				if _, ok := csePublicSettingsMap[vmAvailSet]; ok {
					if existingExtIndex == -1 {
						// add CustomScript extension
						*vm.Resources = append(*vm.Resources, compute.VirtualMachineExtension{})
						existingExtIndex = 0
					}

					cse := (*vm.Resources)[existingExtIndex]
					cse.ProtectedSettings = nil
					//*cse.Name = PrepareDataDiskScriptName
					*cse.Name = customScriptName

					// if there is an existing timestamp, then just increment
					if settings, ok := cse.Settings.(map[string]interface{}); ok {
						for k, v := range settings {
							if k == "timestamp" {
								timestamp = int(v.(float64)) + 1
							}
						}
					}

					vmssCSEPublicSettings.Timestamp = timestamp
					cse.Settings = vmssCSEPublicSettings

					// update VM
					vmUpdate := compute.VirtualMachineUpdate{
						Plan:                     vm.Plan,
						Identity:                 vm.Identity,
						Zones:                    vm.Zones,
						Tags:                     vm.Tags,
						VirtualMachineProperties: vm.VirtualMachineProperties, // should include updates already
					}
					err = azure.UpdateVM(vmsClient, nodeResourceGroup, *vm.Name, vmUpdate)
					if err != nil {
						return fmt.Errorf("issues updating '%s' VM in '%s' AvailabilitySet: %s", *vm.Name, vmAvailSet, err)
					}
					p.ui.Log.Debugf("Updated '%s' VM in '%s' AvailabilitySet.", *vm.Name, vmAvailSet)

					// TODO: check if we need to update script extension twice like in the VMSS
				}
			}
		}
	}

	return nil
}

func (p *Platform) setupScaleSets(session *azure.Session, output map[string]*states.OutputValue, nodeResourceGroup string) error {
	csePublicSettingsMap := make(map[string]publicSettings, len(p.config.NodePools))
	ssDataDisksMap := make(map[string][]compute.VirtualMachineScaleSetDataDisk, len(p.config.NodePools))
	timestamp := 1

	ctx := context.Background()

	vmssClient, err := azure.VMSSClientByEnvStr(p.config.Environment, session)
	if err != nil {
		return fmt.Errorf("issues connecting to Azure via VMSS Client: %s", err)
	}

	vmssVMsClient, err := azure.VMSSVMsClientByEnvStr(p.config.Environment, session)
	if err != nil {
		return fmt.Errorf("issues connecting to Azure via VMSS VMs Client: %s", err)
	}

	vmssList, err := azure.ListVMSSsWithContext(vmssClient, ctx, nodeResourceGroup)
	if err != nil && !strings.Contains(err.Error(), `Code="ResourceGroupNotFound"`) {
		return err
	}

	for _, vmss := range vmssList.Values() {
		// ex: aks-fastcompute-30320017-vmss
		nameSplit := strings.Split(*vmss.Name, "-")
		if len(nameSplit) > 3 && nameSplit[0] == vmssNamePrefix && nameSplit[len(nameSplit)-1] == vmssNameSuffix {
			// check to see if the vmss matches with one of our node pools
			// if so, update
			if nodePool, ok := p.config.NodePools[nameSplit[1]]; ok {

				// if in the future we plan on updating more than just the storage, then we may want to move some code around
				if nodePool.Type == nodePoolTypeVMSS {
					// inherit the default node pool data disks settings if none was given
					if nodePool.DataDisks == nil && p.config.DefaultNodePool.DataDisks != nil && len(*p.config.DefaultNodePool.DataDisks) > 0 {
						nodePool.DataDisks = p.config.DefaultNodePool.DataDisks
					}

					var scriptMetadata map[string]string
					if nodePool.DataDisks == nil || len(*nodePool.DataDisks) == 0 {
						scriptMetadata = make(map[string]string, 1)
					} else {
						scriptMetadata = make(map[string]string, len(*nodePool.DataDisks)+1)
						scriptMetadata["NUM_DATA_DISKS"] = strconv.Itoa(len(*nodePool.DataDisks))
					}

					// docker root
					scriptMetadata["DOCKER_ROOT"] = nodePool.DockerRoot

					// ephemeral disk
					if nodePool.EphemeralMountPoint != "" {
						scriptMetadata["EPHEMERAL_MNT_PT"] = nodePool.EphemeralMountPoint
					} else {
						scriptMetadata["EPHEMERAL_MNT_PT"] = p.config.DefaultNodePool.EphemeralMountPoint
					}
					// data disks
					var vmDataDisks []compute.VirtualMachineScaleSetDataDisk
					if nodePool.DataDisks != nil {
						for di, dv := range *nodePool.DataDisks {
							scriptMetadata["DATA_DISK_"+strconv.Itoa(di)+"_MNT_PT"] = dv.MountPoint
							scriptMetadata["DATA_DISK_"+strconv.Itoa(di)+"_VOL_BYTES"] = strconv.Itoa(dv.VolumeSize * (1024 ^ 3)) // GiB to bytes

							dd := compute.VirtualMachineScaleSetDataDisk{
								// Azure: Parameter 'dataDisk.name' is not allowed.
								//Name:                    to.StringPtr("data"),
								Lun:                     to.Int32Ptr(int32(di)), // azure doesnt support many data disks so we should be ok going from int to int32
								Caching:                 compute.CachingTypesNone,
								WriteAcceleratorEnabled: to.BoolPtr(false),
								CreateOption:            compute.DiskCreateOptionTypesEmpty,
								DiskSizeGB:              to.Int32Ptr(int32(dv.VolumeSize)),
							}
							vmDataDisks = append(vmDataDisks, dd)
						}
					}

					var scriptMetadataList []string
					for k, v := range scriptMetadata {
						scriptMetadataList = append(scriptMetadataList, fmt.Sprintf("%s='%s'", k, v))
					}

					var buff bytes.Buffer
					gz := gzip.NewWriter(&buff)
					//if _, err = gz.Write([]byte(fmt.Sprintf(PrepareDataDiskScript, strings.Join(scriptMetadataList, "\n")))); err != nil {
					//	return fmt.Errorf("issues compressing prepare data disk script: %s", err)
					//}
					if _, err = gz.Write([]byte(customScript)); err != nil {
						return fmt.Errorf("issues compressing prepare temp disk script: %s", err)
					}

					if err = gz.Flush(); err != nil {
						return fmt.Errorf("issues compressing prepare disk script: %s", err)
					}
					if err = gz.Close(); err != nil {
						return fmt.Errorf("issues compressing prepare disk script: %s", err)
					}
					vmssCSEPublicSettings := publicSettings{
						Script: base64.StdEncoding.EncodeToString(buff.Bytes()),
					}

					if scriptMetadata["EPHEMERAL_MNT_PT"] != "" || len(vmDataDisks) > 0 {
						csePublicSettingsMap[nodePool.Name] = vmssCSEPublicSettings
						ssDataDisksMap[nodePool.Name] = vmDataDisks
					}
					// add data disk(s) to vmss storage profile
					if len(vmDataDisks) > 0 {
						vmss.VirtualMachineScaleSetProperties.VirtualMachineProfile.StorageProfile.DataDisks = &vmDataDisks
					} else {
						vmss.VirtualMachineScaleSetProperties.VirtualMachineProfile.StorageProfile.DataDisks = nil
					}

					if _, ok := csePublicSettingsMap[nodePool.Name]; ok {
						// NOTE: we can't use CustomData (cloud-init) since AKS scale sets seems to overwrite it to prep the nodes
						for _, e := range *vmss.VirtualMachineScaleSetProperties.VirtualMachineProfile.ExtensionProfile.Extensions {
							if *e.Name == vmssCustomScriptExtension && *e.VirtualMachineScaleSetExtensionProperties.Publisher == msAzureExtensionsPublisher && *e.VirtualMachineScaleSetExtensionProperties.Type == customScriptExtension {
								// NOTE: passing in a map[string]interface{} throws an error about commandToExecute existing
								// in both public and protected, even though when you marshal the output it doesn't show it
								// so its something weird going on in the Azure side
								e.Settings = nil
								e.ProtectedSettings = nil
								//*e.Name = PrepareDataDiskScriptName
								*e.Name = customScriptName

								// if there is an existing timestamp, then just increment
								if settings, ok := e.VirtualMachineScaleSetExtensionProperties.Settings.(map[string]interface{}); ok {
									for k, v := range settings {
										if k == "timestamp" {
											timestamp = int(v.(float64)) + 1
										}
									}
								}

								vmssCSEPublicSettings.Timestamp = timestamp
								e.VirtualMachineScaleSetExtensionProperties.Settings = vmssCSEPublicSettings
							}
						}

						// update vmss
						err = azure.UpdateVMSSWithContext(vmssClient, ctx, nodeResourceGroup, *vmss.Name, &vmss)
						if err != nil {
							return fmt.Errorf("issues updating '%s' Virtual Machine Scale Set: %s", *vmss.Name, err)
						}
						p.ui.Log.Debugf("Updated '%s' Virtual Machine Scale Set.", *vmss.Name)

						// upgrade vmss
						err = azure.UpgradeVMSSWithContext(vmssClient, vmssVMsClient, ctx, nodeResourceGroup, *vmss.Name, "", "", "")
						if err != nil {
							return fmt.Errorf("issues upgrading '%s Virtual Machine Scale Set: %s", *vmss.Name, err)
						}
						p.ui.Log.Debugf("Upgraded '%s' Virtual Machine Scale Set.", *vmss.Name)
					}
				}
			}
		}
	}

	// update custom script again, because azure does stupid things and it doesn't take the first time
	// this could be because aks doesn't allow for customscript without this hack? (the ui allows it the first time, but not the cli)
	vmssList, err = azure.ListVMSSsWithContext(vmssClient, ctx, nodeResourceGroup)
	if err != nil {
		return err
	}

	timestamp++
	for _, vmss := range vmssList.Values() {
		// ex: aks-fastcompute-30320017-vmss
		nameSplit := strings.Split(*vmss.Name, "-")
		if len(nameSplit) > 3 && nameSplit[0] == vmssNamePrefix && nameSplit[len(nameSplit)-1] == vmssNameSuffix {
			// check to see if the vmss matches with one of our node pools
			// if so, update
			if nodePool, ok := p.config.NodePools[nameSplit[1]]; ok {
				if nodePool.Type == nodePoolTypeVMSS {
					if _, ok := csePublicSettingsMap[nodePool.Name]; ok {
						for _, e := range *vmss.VirtualMachineScaleSetProperties.VirtualMachineProfile.ExtensionProfile.Extensions {
							//if *e.Name == PrepareDataDiskScriptName && *e.VirtualMachineScaleSetExtensionProperties.Publisher == msAzureExtensionsPublisher && *e.VirtualMachineScaleSetExtensionProperties.Type == customScriptExtension {
							if *e.Name == customScriptName && *e.VirtualMachineScaleSetExtensionProperties.Publisher == msAzureExtensionsPublisher && *e.VirtualMachineScaleSetExtensionProperties.Type == customScriptExtension {

								// if there is an existing timestamp, then just increment
								if settings, ok := e.VirtualMachineScaleSetExtensionProperties.Settings.(map[string]interface{}); ok {
									for k, v := range settings {
										if k == "timestamp" {
											timestamp = int(v.(float64)) + 1
										}
									}
								}

								tempSettings := csePublicSettingsMap[nodePool.Name]
								tempSettings.Timestamp = timestamp
								e.VirtualMachineScaleSetExtensionProperties.Settings = tempSettings
							}
						}

						// update vmss
						err = azure.UpdateVMSSWithContext(vmssClient, ctx, nodeResourceGroup, *vmss.Name, &vmss)
						if err != nil {
							return fmt.Errorf("issues updating '%s' Virtual Machine Scale Set: %s", *vmss.Name, err)
						}
						p.ui.Log.Debugf("Updated '%s' Virtual Machine Scale Set again.", *vmss.Name)

						// upgrade vmss
						err = azure.UpgradeVMSSWithContext(vmssClient, vmssVMsClient, ctx, nodeResourceGroup, *vmss.Name, "", "", "")
						if err != nil {
							return fmt.Errorf("issues upgrading '%s Virtual Machine Scale Set: %s", *vmss.Name, err)
						}
						p.ui.Log.Debugf("Upgraded '%s' Virtual Machine Scale Set again.", *vmss.Name)
					}
				}
			}
		}
	}

	return nil
}
