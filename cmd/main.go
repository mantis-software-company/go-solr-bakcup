/*
Copyright 2022 Mantis Software
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

package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	cobradoc "github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	klog "k8s.io/klog/v2"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	rootCmd = &cobra.Command{
		Use:   "solr-backup",
		Short: "A solr backup tool",
		Long:  `A solr cloud backup tool for kube cronjobs`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd)
		},
	}
	version   = ""
	buildTime = ""
	goVersion = ""

	readMeHeader = `# Solr Backup Tool
Backups solr cloud resides at kubernetes. This tools runs inside cronjobs to take backups.
Backup configuration should be reside at solr yaml statefulset. It only triggers async backup and checks its status.
# Usage
`

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Backup/Restore tool\n")
			fmt.Printf("Version: %v\n", version)
			fmt.Printf("Build Time: %v\n", buildTime)
			fmt.Printf("%v\n", goVersion)
		},
	}

	readmeCmd = &cobra.Command{
		Use:    "readme",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {

			lh := func(name string) string {
				base := strings.TrimSuffix(name, path.Ext(name))
				base = strings.Replace(base, "_", "-", -1)
				return "#" + base
			}

			var genReadMe func(cmd *cobra.Command, out *bytes.Buffer) error
			genReadMe = func(cmd *cobra.Command, out *bytes.Buffer) error {
				cmd.DisableAutoGenTag = true
				if err := cobradoc.GenMarkdownCustom(cmd, out, lh); err != nil {
					return err
				}
				for _, subcmd := range cmd.Commands() {
					if err := genReadMe(subcmd, out); err != nil {
						return err
					}
				}
				return nil
			}
			out := new(bytes.Buffer)
			genReadMe(rootCmd, out)
			if rm, err := os.Create("README.md"); err == nil {
				rm.Write([]byte(readMeHeader))
				rm.Write(out.Bytes())
				rm.Close()
			} else {
				klog.V(0).Error(err, "cannot generate readme")
			}
		},
	}
)

func init() {

	klog.InitFlags(nil)

	rootCmd.PersistentFlags().StringP("config", "", "", "configuration file")
	pflag.CommandLine.AddGoFlag(flag.CommandLine.Lookup("v"))
	pflag.CommandLine.AddGoFlag(flag.CommandLine.Lookup("logtostderr"))
	pflag.CommandLine.Set("logtostderr", "true")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(readmeCmd)

}

func Execute() error {
	return rootCmd.Execute()
}

func initializeConfig(cmd *cobra.Command) error {
	klog.V(6).Infof("initialize config")
	progName := filepath.Base(os.Args[0])
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/" + progName)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	} else {
		klog.V(6).Infof("config file loaded from one of default locations")
	}

	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	if configFile != "" {
		klog.V(6).Infof("a config file given as parameter: %v", configFile)
		if r, err := os.Open(configFile); err == nil {
			err = v.MergeConfig(r)
			if err != nil {
				klog.V(6).Error(err, "cannot merge config file")
				return err
			}
			r.Close()
		} else {
			klog.V(6).Error(err, "cannot open config file")
			return err
		}
	}

	v.SetEnvPrefix(strings.ToUpper(progName))
	v.AutomaticEnv()

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			v.BindEnv(f.Name, fmt.Sprintf("%s_%s", strings.ToUpper(progName), envVarSuffix))
		}

		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
	klog.V(6).Infof("config initialized")
	return nil
}

func main() {
	if err := Execute(); err != nil {
		klog.Errorf("backup command failed err=%v", err)
	}
}
