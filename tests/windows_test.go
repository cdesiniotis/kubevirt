/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2018 Red Hat, Inc.
 *
 */

package tests_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kubevirt.io/kubevirt/tests/framework/checks"
	"kubevirt.io/kubevirt/tests/libstorage"

	"k8s.io/utils/pointer"

	"kubevirt.io/kubevirt/pkg/controller"
	"kubevirt.io/kubevirt/tests/libnode"

	"kubevirt.io/kubevirt/pkg/virt-controller/watch/topology"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"

	"kubevirt.io/kubevirt/pkg/network/dns"
	"kubevirt.io/kubevirt/tests"
	"kubevirt.io/kubevirt/tests/flags"
	"kubevirt.io/kubevirt/tests/libnet"
	"kubevirt.io/kubevirt/tests/libvmi"
	"kubevirt.io/kubevirt/tests/util"
)

const (
	windowsDisk        = "windows-disk"
	windowsFirmware    = "5d307ca9-b3ef-428c-8861-06e72d69f223"
	windowsVMIUser     = "Administrator"
	windowsVMIPassword = "Heslo123"
)

const (
	winrmCli    = "winrmcli"
	winrmCliCmd = "winrm-cli"
)

var _ = Describe("[Serial][sig-compute]Windows VirtualMachineInstance", func() {
	var virtClient kubecli.KubevirtClient
	var windowsVMI *v1.VirtualMachineInstance

	BeforeEach(func() {
		const OSWindows = "windows"
		var err error
		virtClient, err = kubecli.GetKubevirtClient()
		util.PanicOnError(err)
		checks.SkipIfMissingRequiredImage(virtClient, tests.DiskWindows)
		libstorage.CreatePVC(OSWindows, "30Gi", libstorage.Config.StorageClassWindows, true)
		windowsVMI = tests.NewRandomVMI()
		windowsVMI.Spec = getWindowsVMISpec()
		tests.AddExplicitPodNetworkInterface(windowsVMI)
		windowsVMI.Spec.Domain.Devices.Interfaces[0].Model = "e1000"
	})

	Context("VMI with HyperV reenlightenment enabled", func() {
		BeforeEach(func() {
			windowsVMI.Spec.Domain.Features.Hyperv.Reenlightenment = &v1.FeatureState{Enabled: pointer.Bool(true)}
		})

		When("TSC frequency is exposed on the cluster", func() {
			It("should be able to migrate", func() {
				if !isTSCFrequencyExposed(virtClient) {
					Skip("TSC frequency is not exposed on the cluster")
				}

				var err error
				By("Creating a windows VM")
				windowsVMI, err = virtClient.VirtualMachineInstance(util.NamespaceTestDefault).Create(windowsVMI)
				Expect(err).ToNot(HaveOccurred())
				tests.WaitForSuccessfulVMIStartWithTimeout(windowsVMI, 360)

				By("Migrating the VM")
				migration := tests.NewRandomMigration(windowsVMI.Name, windowsVMI.Namespace)
				migrationUID := tests.RunMigrationAndExpectCompletion(virtClient, migration, tests.MigrationWaitTime)

				By("Checking VMI, confirm migration state")
				tests.ConfirmVMIPostMigration(virtClient, windowsVMI, migrationUID)
			})
		})

		When("TSC frequency is not exposed on the cluster", func() {

			BeforeEach(func() {
				if isTSCFrequencyExposed(virtClient) {
					nodeList, err := virtClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
					Expect(err).ToNot(HaveOccurred())

					for _, node := range nodeList.Items {
						stopNodeLabeller(node.Name, virtClient)
						removeTSCFrequencyFromNode(node)
					}
				}
			})

			AfterEach(func() {
				nodeList, err := virtClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
				Expect(err).ToNot(HaveOccurred())

				for _, node := range nodeList.Items {
					_, isNodeLabellerStopped := node.Annotations[v1.LabellerSkipNodeAnnotation]
					Expect(isNodeLabellerStopped).To(BeTrue())

					updatedNode := resumeNodeLabeller(node.Name, virtClient)
					_, isNodeLabellerStopped = updatedNode.Annotations[v1.LabellerSkipNodeAnnotation]
					Expect(isNodeLabellerStopped).To(BeFalse(), "after node labeller is resumed, %s annotation is expected to disappear from node", v1.LabellerSkipNodeAnnotation)
				}
			})

			It("should be able to start successfully", func() {
				var err error
				By("Creating a windows VM")
				windowsVMI, err = virtClient.VirtualMachineInstance(util.NamespaceTestDefault).Create(windowsVMI)
				Expect(err).ToNot(HaveOccurred())
				tests.WaitForSuccessfulVMIStartWithTimeout(windowsVMI, 360)
				winrnLoginCommand(virtClient, windowsVMI)
			})

			It("should be marked as non-migratable", func() {
				var err error
				By("Creating a windows VM")
				windowsVMI, err = virtClient.VirtualMachineInstance(util.NamespaceTestDefault).Create(windowsVMI)
				Expect(err).ToNot(HaveOccurred())
				tests.WaitForSuccessfulVMIStartWithTimeout(windowsVMI, 360)

				conditionManager := controller.NewVirtualMachineInstanceConditionManager()
				isNonMigratable := func() error {
					windowsVMI, err = virtClient.VirtualMachineInstance(util.NamespaceTestDefault).Get(windowsVMI.Name, &metav1.GetOptions{})
					Expect(err).ToNot(HaveOccurred())

					cond := conditionManager.GetCondition(windowsVMI, v1.VirtualMachineInstanceIsMigratable)
					const errFmt = "condition " + string(v1.VirtualMachineInstanceIsMigratable) + " is expected to be %s %s"

					if statusFalse := k8sv1.ConditionFalse; cond.Status != statusFalse {
						return fmt.Errorf(errFmt, "of status", string(statusFalse))
					}
					if notMigratableNoTscReason := v1.VirtualMachineInstanceReasonNoTSCFrequencyMigratable; cond.Reason != notMigratableNoTscReason {
						return fmt.Errorf(errFmt, "of reason", notMigratableNoTscReason)
					}
					if !strings.Contains(cond.Message, "HyperV Reenlightenment") {
						return fmt.Errorf(errFmt, "with message that contains", "HyperV Reenlightenment")
					}
					return nil
				}

				Eventually(isNonMigratable, 30*time.Second, time.Second).ShouldNot(HaveOccurred())
				Consistently(isNonMigratable, 15*time.Second, 3*time.Second).ShouldNot(HaveOccurred())
			})
		})
	})

	Context("with winrm connection", func() {
		var winrmcliPod *k8sv1.Pod
		var cli []string

		BeforeEach(func() {
			By("Creating winrm-cli pod for the future use")
			winrmcliPod = &k8sv1.Pod{
				ObjectMeta: metav1.ObjectMeta{GenerateName: winrmCli},
				Spec: k8sv1.PodSpec{
					Containers: []k8sv1.Container{
						{
							Name:    winrmCli,
							Image:   fmt.Sprintf("%s/%s:%s", flags.KubeVirtUtilityRepoPrefix, winrmCli, flags.KubeVirtUtilityVersionTag),
							Command: []string{"sleep"},
							Args:    []string{"3600"},
						},
					},
				},
			}

			var err error
			winrmcliPod, err = virtClient.CoreV1().Pods(util.NamespaceTestDefault).Create(context.Background(), winrmcliPod, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())
		})

		Context("[ref_id:139]VMI is created", func() {

			BeforeEach(func() {
				By("Starting the windows VirtualMachineInstance")
				var err error
				windowsVMI, err = virtClient.VirtualMachineInstance(util.NamespaceTestDefault).Create(windowsVMI)
				Expect(err).ToNot(HaveOccurred())
				tests.WaitForSuccessfulVMIStartWithTimeout(windowsVMI, 360)

				cli = winrnLoginCommand(virtClient, windowsVMI)
			})

			It("[test_id:240]should have correct UUID", func() {
				command := append(cli, "wmic csproduct get \"UUID\"")
				By(fmt.Sprintf("Running \"%s\" command via winrm-cli", command))
				var output string
				Eventually(func() error {
					var err error
					output, err = tests.ExecuteCommandOnPod(
						virtClient,
						winrmcliPod,
						winrmcliPod.Spec.Containers[0].Name,
						command,
					)
					return err
				}, time.Minute*5, time.Second*15).ShouldNot(HaveOccurred())
				By("Checking that the Windows VirtualMachineInstance has expected UUID")
				Expect(output).Should(ContainSubstring(strings.ToUpper(windowsFirmware)))
			})

			It("[test_id:3159]should have default masquerade IP", func() {
				command := append(cli, "ipconfig /all")
				By(fmt.Sprintf("Running \"%s\" command via winrm-cli", command))
				var output string
				Eventually(func() error {
					var err error
					output, err = tests.ExecuteCommandOnPod(
						virtClient,
						winrmcliPod,
						winrmcliPod.Spec.Containers[0].Name,
						command,
					)
					return err
				}, time.Minute*5, time.Second*15).ShouldNot(HaveOccurred())

				By("Checking that the Windows VirtualMachineInstance has expected IP address")
				Expect(output).Should(ContainSubstring("10.0.2.2"))
			})

			It("[test_id:3160]should have the domain set properly", func() {
				searchDomain := getPodSearchDomain(windowsVMI)
				Expect(searchDomain).To(HavePrefix(windowsVMI.Namespace), "should contain a searchdomain with the namespace of the VMI")

				runCommandAndExpectOutput(virtClient,
					winrmcliPod,
					cli,
					"wmic nicconfig get dnsdomain",
					`DNSDomain[\n\r\t ]+`+searchDomain+`[\n\r\t ]+`)
			})
		})

		Context("VMI with subdomain is created", func() {
			BeforeEach(func() {
				windowsVMI.Spec.Subdomain = "subdomain"

				By("Starting the windows VirtualMachineInstance with subdomain")
				var err error
				windowsVMI, err = virtClient.VirtualMachineInstance(util.NamespaceTestDefault).Create(windowsVMI)
				Expect(err).ToNot(HaveOccurred())
				tests.WaitForSuccessfulVMIStartWithTimeout(windowsVMI, 360)

				cli = winrnLoginCommand(virtClient, windowsVMI)
			})

			It("should have the domain set properly with subdomain", func() {
				searchDomain := getPodSearchDomain(windowsVMI)
				Expect(searchDomain).To(HavePrefix(windowsVMI.Namespace), "should contain a searchdomain with the namespace of the VMI")

				expectedSearchDomain := windowsVMI.Spec.Subdomain + "." + searchDomain
				runCommandAndExpectOutput(virtClient,
					winrmcliPod,
					cli,
					"wmic nicconfig get dnsdomain",
					`DNSDomain[\n\r\t ]+`+expectedSearchDomain+`[\n\r\t ]+`)
			})
		})

		Context("with bridge binding", func() {
			BeforeEach(func() {
				By("Starting Windows VirtualMachineInstance with bridge binding")
				windowsVMI.Spec.Domain.Devices.Interfaces = []v1.Interface{libvmi.InterfaceDeviceWithBridgeBinding(libvmi.DefaultInterfaceName)}
				var err error
				windowsVMI, err = virtClient.VirtualMachineInstance(util.NamespaceTestDefault).Create(windowsVMI)
				Expect(err).ToNot(HaveOccurred())
				tests.WaitForSuccessfulVMIStartWithTimeout(windowsVMI, 420)

				cli = winrnLoginCommand(virtClient, windowsVMI)
			})

			It("should be recognized by other pods in cluster", func() {

				By("Pinging virt-handler Pod from Windows VMI")

				var err error
				windowsVMI, err = virtClient.VirtualMachineInstance(windowsVMI.Namespace).Get(windowsVMI.Name, &metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())

				getVirtHandlerPod := func() (*k8sv1.Pod, error) {
					winVmiPod := tests.GetRunningPodByVirtualMachineInstance(windowsVMI, windowsVMI.Namespace)
					nodeName := winVmiPod.Spec.NodeName

					pod, err := kubecli.NewVirtHandlerClient(virtClient).Namespace(flags.KubeVirtInstallNamespace).ForNode(nodeName).Pod()
					if err != nil {
						return nil, fmt.Errorf("failed to get virt-handler pod on node %s: %v", nodeName, err)
					}
					return pod, nil
				}

				virtHandlerPod, err := getVirtHandlerPod()
				Expect(err).ToNot(HaveOccurred())

				virtHandlerPodIP := libnet.GetPodIPByFamily(virtHandlerPod, k8sv1.IPv4Protocol)

				command := append(cli, fmt.Sprintf("ping %s", virtHandlerPodIP))

				By(fmt.Sprintf("Running \"%s\" command via winrm-cli", command))
				Eventually(func() error {
					_, err = tests.ExecuteCommandOnPod(
						virtClient,
						winrmcliPod,
						winrmcliPod.Spec.Containers[0].Name,
						command,
					)
					return err
				}, time.Minute*1, time.Second*15).Should(Succeed())
			})
		})
	})
})

func getWindowsVMISpec() v1.VirtualMachineInstanceSpec {
	gracePeriod := int64(0)
	spinlocks := uint32(8191)
	firmware := types.UID(windowsFirmware)
	_false := false
	return v1.VirtualMachineInstanceSpec{
		TerminationGracePeriodSeconds: &gracePeriod,
		Domain: v1.DomainSpec{
			CPU: &v1.CPU{Cores: 2},
			Features: &v1.Features{
				ACPI: v1.FeatureState{},
				APIC: &v1.FeatureAPIC{},
				Hyperv: &v1.FeatureHyperv{
					Relaxed:    &v1.FeatureState{},
					SyNICTimer: &v1.SyNICTimer{Direct: &v1.FeatureState{}},
					VAPIC:      &v1.FeatureState{},
					Spinlocks:  &v1.FeatureSpinlocks{Retries: &spinlocks},
				},
			},
			Clock: &v1.Clock{
				ClockOffset: v1.ClockOffset{UTC: &v1.ClockOffsetUTC{}},
				Timer: &v1.Timer{
					HPET:   &v1.HPETTimer{Enabled: &_false},
					PIT:    &v1.PITTimer{TickPolicy: v1.PITTickPolicyDelay},
					RTC:    &v1.RTCTimer{TickPolicy: v1.RTCTickPolicyCatchup},
					Hyperv: &v1.HypervTimer{},
				},
			},
			Firmware: &v1.Firmware{UUID: firmware},
			Resources: v1.ResourceRequirements{
				Requests: k8sv1.ResourceList{
					k8sv1.ResourceMemory: resource.MustParse("2048Mi"),
				},
			},
			Devices: v1.Devices{
				Disks: []v1.Disk{
					{
						Name: windowsDisk,
						DiskDevice: v1.DiskDevice{
							Disk: &v1.DiskTarget{
								Bus: v1.DiskBusSATA,
							},
						},
					},
				},
			},
		},
		Volumes: []v1.Volume{
			{
				Name: windowsDisk,
				VolumeSource: v1.VolumeSource{
					Ephemeral: &v1.EphemeralVolumeSource{
						PersistentVolumeClaim: &k8sv1.PersistentVolumeClaimVolumeSource{
							ClaimName: tests.DiskWindows,
						},
					},
				},
			},
		},
	}
}

func winrnLoginCommand(virtClient kubecli.KubevirtClient, windowsVMI *v1.VirtualMachineInstance) []string {
	var err error
	windowsVMI, err = virtClient.VirtualMachineInstance(windowsVMI.Namespace).Get(windowsVMI.Name, &metav1.GetOptions{})
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	vmiIp := windowsVMI.Status.Interfaces[0].IP
	cli := []string{
		winrmCliCmd,
		"-hostname",
		vmiIp,
		"-username",
		windowsVMIUser,
		"-password",
		windowsVMIPassword,
	}

	return cli
}

func getPodSearchDomain(windowsVMI *v1.VirtualMachineInstance) string {
	By("fetching /etc/resolv.conf from the VMI Pod")
	resolvConf := tests.RunCommandOnVmiPod(windowsVMI, []string{"cat", "/etc/resolv.conf"})

	By("extracting the search domain of the VMI")
	searchDomains, err := dns.ParseSearchDomains(resolvConf)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	searchDomain := ""
	for _, s := range searchDomains {
		if len(searchDomain) < len(s) {
			searchDomain = s
		}
	}

	return searchDomain
}

func runCommandAndExpectOutput(virtClient kubecli.KubevirtClient, winrmcliPod *k8sv1.Pod, cli []string, command, expectedOutputRegex string) {
	cliCmd := append(cli, command)
	By(fmt.Sprintf("Running \"%s\" command via winrm-cli", cliCmd))
	By("first making sure that we can execute VMI commands")
	EventuallyWithOffset(1, func() error {
		_, err := tests.ExecuteCommandOnPod(
			virtClient,
			winrmcliPod,
			winrmcliPod.Spec.Containers[0].Name,
			cliCmd,
		)
		return err
	}, time.Minute*5, time.Second*15).ShouldNot(HaveOccurred())

	By("repeatedly trying to get the search domain, since it may take some time until the domain is set")
	EventuallyWithOffset(1, func() string {
		output, err := tests.ExecuteCommandOnPod(
			virtClient,
			winrmcliPod,
			winrmcliPod.Spec.Containers[0].Name,
			cliCmd,
		)
		Expect(err).ToNot(HaveOccurred())
		return output
	}, time.Minute*1, time.Second*10).Should(MatchRegexp(expectedOutputRegex))
}

func isTSCFrequencyExposed(virtClient kubecli.KubevirtClient) bool {
	nodeList, err := virtClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())

	for _, node := range nodeList.Items {
		if _, isExposed := node.Labels[topology.TSCFrequencyLabel]; isExposed {
			return true
		}
	}

	return false
}

func removeTSCFrequencyFromNode(node k8sv1.Node) {
	for _, baseLabelToRemove := range []string{topology.TSCFrequencyLabel, topology.TSCFrequencySchedulingLabel} {
		for key, _ := range node.Labels {
			if strings.HasPrefix(key, baseLabelToRemove) {
				libnode.RemoveLabelFromNode(node.Name, key)
			}
		}
	}
}
