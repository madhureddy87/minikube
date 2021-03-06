/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	core "k8s.io/api/core/v1"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	pkg_config "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/service"
)

var serviceListNamespace string

// serviceListCmd represents the service list command
var serviceListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "Lists the URLs for the services in your local cluster",
	Long:  `Lists the URLs for the services in your local cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()
		profileName := viper.GetString(pkg_config.ProfileName)
		cfg, err := config.Load(profileName)
		if err != nil {
			exit.WithError("Error getting config", err)
		}
		cp, err := config.PrimaryControlPlane(*cfg)
		if err != nil {
			exit.WithError("Error getting primary control plane", err)
		}
		if !machine.IsHostRunning(api, driver.MachineName(*cfg, cp)) {
			exit.WithCodeT(exit.Unavailable, "profile {{.name}} is not running.", out.V{"name": profileName})
		}
		serviceURLs, err := service.GetServiceURLs(api, serviceListNamespace, serviceURLTemplate)
		if err != nil {
			out.FatalT("Failed to get service URL: {{.error}}", out.V{"error": err})
			out.ErrT(out.Notice, "Check that minikube is running and that you have specified the correct namespace (-n flag) if required.")
			os.Exit(exit.Unavailable)
		}

		var data [][]string
		for _, serviceURL := range serviceURLs {
			if len(serviceURL.URLs) == 0 {
				data = append(data, []string{serviceURL.Namespace, serviceURL.Name, "No node port"})
			} else {
				serviceURLs := strings.Join(serviceURL.URLs, "\n")

				// if we are running Docker on OSX we empty the internal service URLs
				if runtime.GOOS == "darwin" && cfg.Driver == oci.Docker {
					serviceURLs = ""
				}

				data = append(data, []string{serviceURL.Namespace, serviceURL.Name, "", serviceURLs})
			}
		}

		service.PrintServiceList(os.Stdout, data)
	},
}

func init() {
	serviceListCmd.Flags().StringVarP(&serviceListNamespace, "namespace", "n", core.NamespaceAll, "The services namespace")
	serviceCmd.AddCommand(serviceListCmd)
}
