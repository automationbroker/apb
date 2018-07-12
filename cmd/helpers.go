//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Creates a hidden copy of a cobra.Command with optional deprecation text
func createHiddenCmd(cmd *cobra.Command, deprecatedText string) *cobra.Command {
	newCmd := &cobra.Command{
		Use:        cmd.Use,
		Short:      cmd.Short,
		Long:       cmd.Long,
		Args:       cmd.Args,
		Run:        cmd.Run,
		Hidden:     true,
		Deprecated: deprecatedText,
	}
	newCmd.Flags().AddFlagSet(cmd.Flags())
	return newCmd
}

func handleBearerTokenErr() {
	log.Error("Bearer token not found for current 'oc' user. Log in as a different user and retry.")
	log.Info("Some users don't have a token, including 'system:admin'")
	log.Info("View current token with 'oc whoami -t'")
}

func handleResourceInaccessibleErr(resourceType string, namespace string, restateErr bool) {
	errMsg := ""
	if restateErr {
		if namespace != "" {
			errMsg += fmt.Sprintf("Current 'oc' user unable to get '%s' in namespace [%s]. ", resourceType, namespace)
		} else {
			errMsg += fmt.Sprintf("Current 'oc' user unable to get '%s'. ", resourceType)
		}
	}
	log.Errorf(errMsg + "Try again with a more privileged user.")
	log.Info("Administrators can grant 'cluster-admin' privileges with:\n   oc adm policy add-cluster-role-to-user cluster-admin <oc-user>")
}
