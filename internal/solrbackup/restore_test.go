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

package solrbackup

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Restore Methods Tests", func() {
	Context("Backup Tests", func() {

		var config Config
		config.SolrEndpoint = "http://solr.solr-backup-test.svc.cluster.mantam:8983"
		config.Collections = []string{"test", "test1"}
		config.Location = "/"
		config.RetaintionDays = 5

		reqId := time.Now().UnixMilli()

		Describe("Test restore single manually", func() {
			It("StartRestoreInplace should be succeed", func() {
				err := StartRestoreInplace(config, 0, reqId)
				Expect(err).To(BeNil(), "start restore returns error")
			})

			It("waitRequestStatus should be succeed", func() {
				err := waitRequestStatus(config, reqId)
				Expect(err).To(BeNil(), "waitRequestStatus returns error")
			})

			It("deleteRequestId should be succeed", func() {
				err := deleteRequestId(config, reqId)
				Expect(err).To(BeNil(), "deleteRequestId returns error")
			})
		})

		Describe("Test restore single together", func() {
			It("RestoreInplace should be succeed", func() {
				err := RestoreInplace(config, 0)
				Expect(err).To(BeNil(), "Backup returns error")
			})
		})

		Describe("Test restore all together", func() {
			It("RestoreAllInplace should be succeed", func() {
				err := RestoreAllInplace(config)
				Expect(err).To(BeNil(), "BackupAll returns error")
			})
		})

	})
})
