package services

import (
	"fmt"
	"path/filepath"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/tools/cache"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/log"

	"kubevirt.io/kubevirt/pkg/config"
	"kubevirt.io/kubevirt/pkg/hooks"
	hostdisk "kubevirt.io/kubevirt/pkg/host-disk"
	"kubevirt.io/kubevirt/pkg/network/sriov"
	"kubevirt.io/kubevirt/pkg/storage/types"
	"kubevirt.io/kubevirt/pkg/util"
)

type VolumeRendererOption func(renderer *VolumeRenderer) error

type VolumeRenderer struct {
	containerDiskDir string
	ephemeralDiskDir string
	virtShareDir     string
	namespace        string
	vmiVolumes       []v1.Volume
	podVolumes       []k8sv1.Volume
	podVolumeMounts  []k8sv1.VolumeMount
	volumeDevices    []k8sv1.VolumeDevice
}

func NewVolumeRenderer(namespace string, ephemeralDisk string, containerDiskDir string, virtShareDir string, volumeOptions ...VolumeRendererOption) (*VolumeRenderer, error) {
	volumeRenderer := &VolumeRenderer{
		containerDiskDir: containerDiskDir,
		ephemeralDiskDir: ephemeralDisk,
		namespace:        namespace,
		virtShareDir:     virtShareDir,
	}
	for _, volumeOption := range volumeOptions {
		if err := volumeOption(volumeRenderer); err != nil {
			return nil, err
		}
	}
	return volumeRenderer, nil
}

func (vr *VolumeRenderer) Mounts() []k8sv1.VolumeMount {
	volumeMounts := []k8sv1.VolumeMount{
		mountPath("private", util.VirtPrivateDir),
		mountPath("public", util.VirtShareDir),
		mountPath("ephemeral-disks", vr.ephemeralDiskDir),
		mountPathWithPropagation(containerDisks, vr.containerDiskDir, k8sv1.MountPropagationHostToContainer),
		mountPath("libvirt-runtime", "/var/run/libvirt"),
		mountPath("sockets", filepath.Join(vr.virtShareDir, "sockets")),
	}
	return append(volumeMounts, vr.podVolumeMounts...)
}

func (vr *VolumeRenderer) Volumes() []k8sv1.Volume {
	volumes := []k8sv1.Volume{
		emptyDirVolume("private"),
		emptyDirVolume("public"),
		emptyDirVolume("sockets"),
		emptyDirVolume(virtBinDir),
		emptyDirVolume("libvirt-runtime"),
		emptyDirVolume("ephemeral-disks"),
		emptyDirVolume(containerDisks),
	}
	return append(volumes, vr.podVolumes...)
}

func (vr *VolumeRenderer) VolumeDevices() []k8sv1.VolumeDevice {
	return vr.volumeDevices
}

func mountPath(name string, path string) k8sv1.VolumeMount {
	return k8sv1.VolumeMount{
		Name:      name,
		MountPath: path,
	}
}

func mountPathWithPropagation(name string, path string, propagation k8sv1.MountPropagationMode) k8sv1.VolumeMount {
	return k8sv1.VolumeMount{
		Name:             name,
		MountPath:        path,
		MountPropagation: &propagation,
	}
}

func emptyDirVolume(name string) k8sv1.Volume {
	return k8sv1.Volume{
		Name: name,
		VolumeSource: k8sv1.VolumeSource{
			EmptyDir: &k8sv1.EmptyDirVolumeSource{}},
	}
}

func downwardAPIDirVolume(name, path, fieldPath string) k8sv1.Volume {
	return k8sv1.Volume{
		Name: name,
		VolumeSource: k8sv1.VolumeSource{
			DownwardAPI: &k8sv1.DownwardAPIVolumeSource{
				Items: []k8sv1.DownwardAPIVolumeFile{
					{
						Path: path,
						FieldRef: &k8sv1.ObjectFieldSelector{
							FieldPath: fieldPath,
						},
					},
				},
			},
		},
	}
}

func withVMIVolumes(pvcStore cache.Store, vmiSpecVolumes []v1.Volume, vmiVolumeStatus []v1.VolumeStatus) VolumeRendererOption {
	return func(renderer *VolumeRenderer) error {
		hotplugVolumesByName := hotplugVolumes(vmiVolumeStatus, vmiSpecVolumes)
		for _, volume := range vmiSpecVolumes {
			if _, isHotplugVolume := hotplugVolumesByName[volume.Name]; isHotplugVolume {
				continue
			}

			if volume.PersistentVolumeClaim != nil {
				if err := renderer.handlePVCVolume(volume, pvcStore); err != nil {
					return err
				}
			}

			if volume.Ephemeral != nil {
				if err := renderer.handleEphemeralVolume(volume, pvcStore); err != nil {
					return err
				}
			}

			if volume.HostDisk != nil {
				renderer.handleHostDisk(volume)
			}

			if volume.DataVolume != nil {
				if err := renderer.handleDataVolume(volume, pvcStore); err != nil {
					return err
				}
			}

			if volume.ConfigMap != nil {
				renderer.handleConfigMap(volume)
			}

			if volume.Secret != nil {
				renderer.handleSecret(volume)
			}

			if volume.DownwardAPI != nil {
				renderer.handleDownwardAPI(volume)
			}

			if volume.DownwardMetrics != nil {
				renderer.handleDownwardMetrics(volume)
			}

			if volume.CloudInitNoCloud != nil {
				renderer.handleCloudInitNoCloud(volume)
			}

			if volume.Sysprep != nil {
				renderer.handleSysprep(volume)
			}

			if volume.CloudInitConfigDrive != nil {
				renderer.handleCloudInitConfigDrive(volume)
			}
		}
		return nil
	}
}

func (vr *VolumeRenderer) handleCloudInitConfigDrive(volume v1.Volume) {
	if volume.CloudInitConfigDrive != nil {
		if volume.CloudInitConfigDrive.UserDataSecretRef != nil {
			// attach a secret referenced by the user
			volumeName := volume.Name + "-udata"
			vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
				Name: volumeName,
				VolumeSource: k8sv1.VolumeSource{
					Secret: &k8sv1.SecretVolumeSource{
						SecretName: volume.CloudInitConfigDrive.UserDataSecretRef.Name,
					},
				},
			})
			vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
				Name:      volumeName,
				MountPath: filepath.Join(config.SecretSourceDir, volume.Name, "userdata"),
				SubPath:   "userdata",
				ReadOnly:  true,
			})
			vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
				Name:      volumeName,
				MountPath: filepath.Join(config.SecretSourceDir, volume.Name, "userData"),
				SubPath:   "userData",
				ReadOnly:  true,
			})
		}
		if volume.CloudInitConfigDrive.NetworkDataSecretRef != nil {
			// attach a secret referenced by the networkdata
			volumeName := volume.Name + "-ndata"
			vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
				Name: volumeName,
				VolumeSource: k8sv1.VolumeSource{
					Secret: &k8sv1.SecretVolumeSource{
						SecretName: volume.CloudInitConfigDrive.NetworkDataSecretRef.Name,
					},
				},
			})
			vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
				Name:      volumeName,
				MountPath: filepath.Join(config.SecretSourceDir, volume.Name, "networkdata"),
				SubPath:   "networkdata",
				ReadOnly:  true,
			})
			vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
				Name:      volumeName,
				MountPath: filepath.Join(config.SecretSourceDir, volume.Name, "networkData"),
				SubPath:   "networkData",
				ReadOnly:  true,
			})
		}
	}
}

func (vr *VolumeRenderer) handleSysprep(volume v1.Volume) {
	if volume.Sysprep != nil {
		var volumeSource k8sv1.VolumeSource
		// attach a Secret or ConfigMap referenced by the user
		volumeSource, err := sysprepVolumeSource(*volume.Sysprep)
		if err != nil {
			//return nil, err
		}
		vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
			Name:         volume.Name,
			VolumeSource: volumeSource,
		})
		vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
			Name:      volume.Name,
			MountPath: filepath.Join(config.SysprepSourceDir, volume.Name),
			ReadOnly:  true,
		})
	}
}

func hotplugVolumes(vmiVolumeStatus []v1.VolumeStatus, vmiSpecVolumes []v1.Volume) map[string]struct{} {
	hotplugVolumeSet := map[string]struct{}{}
	for _, volumeStatus := range vmiVolumeStatus {
		if volumeStatus.HotplugVolume != nil {
			hotplugVolumeSet[volumeStatus.Name] = struct{}{}
		}
	}
	// This detects hotplug volumes for a started but not ready VMI
	for _, volume := range vmiSpecVolumes {
		if (volume.DataVolume != nil && volume.DataVolume.Hotpluggable) || (volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.Hotpluggable) {
			hotplugVolumeSet[volume.Name] = struct{}{}
		}
	}
	return hotplugVolumeSet
}

func withAccessCredentials(accessCredentials []v1.AccessCredential) VolumeRendererOption {
	return func(renderer *VolumeRenderer) error {
		for _, accessCred := range accessCredentials {
			secretName := ""
			if accessCred.SSHPublicKey != nil && accessCred.SSHPublicKey.Source.Secret != nil {
				secretName = accessCred.SSHPublicKey.Source.Secret.SecretName
			} else if accessCred.UserPassword != nil && accessCred.UserPassword.Source.Secret != nil {
				secretName = accessCred.UserPassword.Source.Secret.SecretName
			}

			if secretName == "" {
				continue
			}
			volumeName := secretName + "-access-cred"
			renderer.podVolumes = append(renderer.podVolumes, k8sv1.Volume{
				Name: volumeName,
				VolumeSource: k8sv1.VolumeSource{
					Secret: &k8sv1.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			})
			renderer.podVolumeMounts = append(renderer.podVolumeMounts, k8sv1.VolumeMount{
				Name:      volumeName,
				MountPath: filepath.Join(config.SecretSourceDir, volumeName),
				ReadOnly:  true,
			})
		}
		return nil
	}
}

func withSidecarVolumes(hookSidecars hooks.HookSidecarList) VolumeRendererOption {
	return func(renderer *VolumeRenderer) error {
		if len(hookSidecars) != 0 {
			renderer.podVolumes = append(renderer.podVolumes, emptyDirVolume(hookSidecarSocks))
			renderer.podVolumeMounts = append(renderer.podVolumeMounts, k8sv1.VolumeMount{
				Name:      hookSidecarSocks,
				MountPath: hooks.HookSocketsSharedDirectory,
			})
		}
		return nil
	}
}

func withHugepages() VolumeRendererOption {
	return func(renderer *VolumeRenderer) error {
		renderer.podVolumes = append(renderer.podVolumes, k8sv1.Volume{
			Name: "hugepages",
			VolumeSource: k8sv1.VolumeSource{
				EmptyDir: &k8sv1.EmptyDirVolumeSource{
					Medium: k8sv1.StorageMediumHugePages,
				},
			},
		})
		renderer.podVolumeMounts = append(renderer.podVolumeMounts, k8sv1.VolumeMount{
			Name:      "hugepages",
			MountPath: filepath.Join("/dev/hugepages"),
		})
		return nil
	}
}

func withHotplugSupport(hotplugDiskDir string) VolumeRendererOption {
	return func(renderer *VolumeRenderer) error {
		prop := k8sv1.MountPropagationHostToContainer
		renderer.podVolumeMounts = append(renderer.podVolumeMounts, k8sv1.VolumeMount{
			Name:             hotplugDisks,
			MountPath:        hotplugDiskDir,
			MountPropagation: &prop,
		})
		renderer.podVolumes = append(renderer.podVolumes, emptyDirVolume(hotplugDisks))
		return nil
	}
}

func withSRIOVPciMapAnnotation() VolumeRendererOption {
	return func(renderer *VolumeRenderer) error {
		renderer.podVolumeMounts = append(renderer.podVolumeMounts, mountPath(sriov.VolumeName, sriov.MountPath))
		renderer.podVolumes = append(renderer.podVolumes,
			downwardAPIDirVolume(
				sriov.VolumeName, sriov.VolumePath, fmt.Sprintf("metadata.annotations['%s']", sriov.NetworkPCIMapAnnot)),
		)
		return nil
	}
}

func imgPullSecrets(volumes ...v1.Volume) []k8sv1.LocalObjectReference {
	var imagePullSecrets []k8sv1.LocalObjectReference
	for _, volume := range volumes {
		if volume.ContainerDisk != nil && volume.ContainerDisk.ImagePullSecret != "" {
			imagePullSecrets = appendUniqueImagePullSecret(imagePullSecrets, k8sv1.LocalObjectReference{
				Name: volume.ContainerDisk.ImagePullSecret,
			})
		}
	}
	return imagePullSecrets
}

func serviceAccount(volumes ...v1.Volume) string {
	for _, volume := range volumes {
		if volume.ServiceAccount != nil {
			return volume.ServiceAccount.ServiceAccountName
		}
	}
	return ""
}

func (vr *VolumeRenderer) addPVCToLaunchManifest(pvcStore cache.Store, volume v1.Volume, claimName string) error {
	logger := log.DefaultLogger()
	_, exists, isBlock, err := types.IsPVCBlockFromStore(pvcStore, vr.namespace, claimName)
	if err != nil {
		logger.Errorf("error getting PVC: %v", claimName)
		return err
	} else if !exists {
		logger.Errorf("didn't find PVC %v", claimName)
		return PvcNotFoundError{Reason: fmt.Sprintf("didn't find PVC %v", claimName)}
	} else if isBlock {
		devicePath := filepath.Join(string(filepath.Separator), "dev", volume.Name)
		device := k8sv1.VolumeDevice{
			Name:       volume.Name,
			DevicePath: devicePath,
		}
		vr.volumeDevices = append(vr.volumeDevices, device)
	} else {
		volumeMount := k8sv1.VolumeMount{
			Name:      volume.Name,
			MountPath: hostdisk.GetMountedHostDiskDir(volume.Name),
		}
		vr.podVolumeMounts = append(vr.podVolumeMounts, volumeMount)
	}
	return nil
}

func (vr *VolumeRenderer) handlePVCVolume(volume v1.Volume, pvcStore cache.Store) error {
	claimName := volume.PersistentVolumeClaim.ClaimName
	if err := vr.addPVCToLaunchManifest(pvcStore, volume, claimName); err != nil {
		return err
	}
	vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
		Name: volume.Name,
		VolumeSource: k8sv1.VolumeSource{
			PersistentVolumeClaim: &k8sv1.PersistentVolumeClaimVolumeSource{
				ClaimName: volume.PersistentVolumeClaim.ClaimName,
				ReadOnly:  volume.PersistentVolumeClaim.ReadOnly,
			},
		},
	})
	return nil
}

func (vr *VolumeRenderer) handleEphemeralVolume(volume v1.Volume, pvcStore cache.Store) error {
	claimName := volume.Ephemeral.PersistentVolumeClaim.ClaimName
	if err := vr.addPVCToLaunchManifest(pvcStore, volume, claimName); err != nil {
		return err
	}
	vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
		Name: volume.Name,
		VolumeSource: k8sv1.VolumeSource{
			PersistentVolumeClaim: volume.Ephemeral.PersistentVolumeClaim,
		},
	})
	return nil
}

func (vr *VolumeRenderer) handleDataVolume(volume v1.Volume, pvcStore cache.Store) error {
	claimName := volume.DataVolume.Name
	if err := vr.addPVCToLaunchManifest(pvcStore, volume, claimName); err != nil {
		return err
	}
	vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
		Name: volume.Name,
		VolumeSource: k8sv1.VolumeSource{
			PersistentVolumeClaim: &k8sv1.PersistentVolumeClaimVolumeSource{
				ClaimName: claimName,
			},
		},
	})
	return nil
}

func (vr *VolumeRenderer) handleHostDisk(volume v1.Volume) {
	var hostPathType k8sv1.HostPathType

	switch hostType := volume.HostDisk.Type; hostType {
	case v1.HostDiskExists:
		hostPathType = k8sv1.HostPathDirectory
	case v1.HostDiskExistsOrCreate:
		hostPathType = k8sv1.HostPathDirectoryOrCreate
	}

	vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
		Name:      volume.Name,
		MountPath: hostdisk.GetMountedHostDiskDir(volume.Name),
	})
	vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
		Name: volume.Name,
		VolumeSource: k8sv1.VolumeSource{
			HostPath: &k8sv1.HostPathVolumeSource{
				Path: filepath.Dir(volume.HostDisk.Path),
				Type: &hostPathType,
			},
		},
	})
}

func (vr *VolumeRenderer) handleSecret(volume v1.Volume) {
	vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
		Name:      volume.Name,
		MountPath: filepath.Join(config.SecretSourceDir, volume.Name),
		ReadOnly:  true,
	})
	vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
		Name: volume.Name,
		VolumeSource: k8sv1.VolumeSource{
			Secret: &k8sv1.SecretVolumeSource{
				SecretName: volume.Secret.SecretName,
				Optional:   volume.Secret.Optional,
			},
		},
	})
}

func (vr *VolumeRenderer) handleConfigMap(volume v1.Volume) {
	vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
		Name:      volume.Name,
		MountPath: filepath.Join(config.ConfigMapSourceDir, volume.Name),
		ReadOnly:  true,
	})
	vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
		Name: volume.Name,
		VolumeSource: k8sv1.VolumeSource{
			ConfigMap: &k8sv1.ConfigMapVolumeSource{
				LocalObjectReference: volume.ConfigMap.LocalObjectReference,
				Optional:             volume.ConfigMap.Optional,
			},
		},
	})
}

func (vr *VolumeRenderer) handleDownwardAPI(volume v1.Volume) {
	// attach a Secret to the pod
	vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
		Name:      volume.Name,
		MountPath: filepath.Join(config.DownwardAPISourceDir, volume.Name),
		ReadOnly:  true,
	})
	vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
		Name: volume.Name,
		VolumeSource: k8sv1.VolumeSource{
			DownwardAPI: &k8sv1.DownwardAPIVolumeSource{
				Items: volume.DownwardAPI.Fields,
			},
		},
	})
}

func (vr *VolumeRenderer) handleCloudInitNoCloud(volume v1.Volume) {
	if volume.CloudInitNoCloud.UserDataSecretRef != nil {
		// attach a secret referenced by the user
		volumeName := volume.Name + "-udata"
		vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
			Name: volumeName,
			VolumeSource: k8sv1.VolumeSource{
				Secret: &k8sv1.SecretVolumeSource{
					SecretName: volume.CloudInitNoCloud.UserDataSecretRef.Name,
				},
			},
		})
		vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
			Name:      volumeName,
			MountPath: filepath.Join(config.SecretSourceDir, volume.Name, "userdata"),
			SubPath:   "userdata",
			ReadOnly:  true,
		})
		vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
			Name:      volumeName,
			MountPath: filepath.Join(config.SecretSourceDir, volume.Name, "userData"),
			SubPath:   "userData",
			ReadOnly:  true,
		})
	}
	if volume.CloudInitNoCloud.NetworkDataSecretRef != nil {
		// attach a secret referenced by the networkdata
		volumeName := volume.Name + "-ndata"
		vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
			Name: volumeName,
			VolumeSource: k8sv1.VolumeSource{
				Secret: &k8sv1.SecretVolumeSource{
					SecretName: volume.CloudInitNoCloud.NetworkDataSecretRef.Name,
				},
			},
		})
		vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
			Name:      volumeName,
			MountPath: filepath.Join(config.SecretSourceDir, volume.Name, "networkdata"),
			SubPath:   "networkdata",
			ReadOnly:  true,
		})
		vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
			Name:      volumeName,
			MountPath: filepath.Join(config.SecretSourceDir, volume.Name, "networkData"),
			SubPath:   "networkData",
			ReadOnly:  true,
		})
	}
}

func (vr *VolumeRenderer) handleDownwardMetrics(volume v1.Volume) {
	sizeLimit := resource.MustParse("1Mi")
	vr.podVolumes = append(vr.podVolumes, k8sv1.Volume{
		Name: volume.Name,
		VolumeSource: k8sv1.VolumeSource{
			EmptyDir: &k8sv1.EmptyDirVolumeSource{
				Medium:    "Memory",
				SizeLimit: &sizeLimit,
			},
		},
	})
	vr.podVolumeMounts = append(vr.podVolumeMounts, k8sv1.VolumeMount{
		Name:      volume.Name,
		MountPath: config.DownwardMetricDisksDir,
	})
}
