package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/joeyloman/kubevirt-ip-helper/pkg/util"

	kihclientset "github.com/joeyloman/kubevirt-ip-helper/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"github.com/ghodss/yaml"
	kihv1 "github.com/joeyloman/kubevirt-ip-helper/pkg/apis/kubevirtiphelper.k8s.binbash.org/v1"
)

func checkForVirtualMachineInstance(k8s_clientset *kubernetes.Clientset, vmNamespace string, vmName string) bool {
	_, err := k8s_clientset.RESTClient().Get().AbsPath("/apis/kubevirt.io/v1").Namespace(vmNamespace).Resource("virtualmachines").Name(vmName).DoRaw(context.TODO())
	if err != nil {
		if strings.Contains(err.Error(), "the server could not find the requested resource") {
			return true
		}
	}

	return false
}

func backupVirtualMachineNetworkConfiguration(kih_clientset *kihclientset.Clientset, vmNamespace string, vmName string) (err error) {
	vmNetCfgObj, err := kih_clientset.KubevirtiphelperV1().VirtualMachineNetworkConfigs(vmNamespace).Get(context.TODO(), vmName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error while fetching VirtualMachineNetworkConfig object: %s", err.Error())
	}

	scheme := runtime.NewScheme()
	err = kihv1.AddToScheme(scheme)
	if err != nil {
		return err
	}
	codec := serializer.NewCodecFactory(scheme).LegacyCodec(kihv1.SchemeGroupVersion)
	jsonData, err := runtime.Encode(codec, vmNetCfgObj)
	if err != nil {
		return err
	}
	yamlData, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s_%s.yaml", vmNamespace, vmName)
	fmt.Printf("writing vmnetcfg object backup to: %s\n", filename)

	return os.WriteFile(filename, yamlData, 0644)
}

func gatherVirtualMachineNetworkConfigurations(k8s_clientset *kubernetes.Clientset, kih_clientset *kihclientset.Clientset) (err error) {
	vmNetCfgObjs, err := kih_clientset.KubevirtiphelperV1().VirtualMachineNetworkConfigs(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("cannot fetch VirtualMachineNetworkConfig objects: %s", err.Error())
	}

	for _, vmnetcfg := range vmNetCfgObjs.Items {
		if checkForVirtualMachineInstance(k8s_clientset, vmnetcfg.Namespace, vmnetcfg.Name) {
			fmt.Printf("\n%s/%s has no vm, backup and cleanup the vmnetcfg object? (y/n) ", vmnetcfg.Namespace, vmnetcfg.Name)
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err == nil {
				input = strings.TrimSpace(input)
				if input == "y" {
					if err = backupVirtualMachineNetworkConfiguration(kih_clientset, vmnetcfg.Namespace, vmnetcfg.Name); err != nil {
						fmt.Printf("failed to write backup file: %s\n", err.Error())
					} else {
						fmt.Printf("backup succeeded, removing vmnetcfg object..")
						if err := kih_clientset.KubevirtiphelperV1().VirtualMachineNetworkConfigs(vmnetcfg.Namespace).Delete(context.TODO(), vmnetcfg.Name, metav1.DeleteOptions{}); err != nil {
							fmt.Printf("error while deleting VirtualMachineNetworkConfig object %s/%s: %s",
								vmnetcfg.Namespace, vmnetcfg.Name, err.Error())
						}
						fmt.Printf("done!\n")
					}

				}
			}
		}
	}

	return
}

func main() {
	var kubeconfig_file string
	kubeconfig_file = os.Getenv("KUBECONFIG")
	if kubeconfig_file == "" {
		kubeconfig_file = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	var config *rest.Config
	if util.FileExists(kubeconfig_file) {
		// uses kubeconfig
		kubeconfig := flag.String("kubeconfig", kubeconfig_file, "(optional) absolute path to the kubeconfig file")
		flag.Parse()
		config_kube, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		config = config_kube
	} else {
		// creates the in-cluster config
		config_rest, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		config = config_rest
	}

	k8s_clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	kih_clientset, err := kihclientset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	if err := gatherVirtualMachineNetworkConfigurations(k8s_clientset, kih_clientset); err != nil {
		fmt.Printf("error: %s", err.Error())
	}
}
