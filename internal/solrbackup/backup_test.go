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

var _ = Describe("Backup Methods Tests", func() {
	Context("Backup Tests", func() {

		var config Config
		config.SolrEndpoint = "http://solr.solr-backup-test.svc.cluster.mantam:8983"
		config.Collections = []string{"test", "test1"}
		config.Location = "/"
		config.RetaintionDays = 1

		reqId := time.Now().UnixMilli()

		Describe("Test open/close", func() {
			It("Should be succeed", func() {})
		})

		Describe("Test backup single manually", func() {
			It("StartBackup should be succeed", func() {
				err := StartBackup(config, 0, reqId)
				Expect(err).To(BeNil(), "start backup returns error")
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

		Describe("Test backup single together", func() {
			It("Backup should be succeed", func() {
				err := Backup(config, 0)
				Expect(err).To(BeNil(), "Backup returns error")
			})
		})

		Describe("Test backup all together", func() {
			It("BackupAll should be succeed", func() {
				err := BackupAll(config)
				Expect(err).To(BeNil(), "BackupAll returns error")
			})
		})

		Describe("Test list backup", func() {
			It("BackupList should be succeed", func() {
				err := BackupList(config, 0)
				Expect(err).To(BeNil(), "BackupList returns error")
			})

			It("BackupListAll should be succeed", func() {
				err := BackupListAll(config)
				Expect(err).To(BeNil(), "BackupListAll returns error")
			})
		})

		Describe("Test delete backup", func() {
			It("BackupDelete should be succeed", func() {
				err := BackupDelete(config, 0)
				Expect(err).To(BeNil(), "BackupDelete returns error")
			})

			It("BackupDeleteAll should be succeed", func() {
				err := BackupDeleteAll(config)
				Expect(err).To(BeNil(), "BackupDeleteAll returns error")
			})
		})

	})
})
