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
	"errors"
	"fmt"
	prettytable "github.com/jedib0t/go-pretty/v6/table"
	klog "k8s.io/klog/v2"
	"os"
	"reflect"
	"sort"
	"time"
)

func BackupList(config Config, colId int64) error {
	col := config.Collections[colId]

	backup_uri := fmt.Sprintf("%s%s?action=LISTBACKUP&name=%s&location=%s", config.SolrEndpoint, collection_api, col, config.Location)
	klog.V(5).Infof("backup uri: %v", backup_uri)

	resp, err := sendRequest(backup_uri)

	if err != nil {
		klog.Errorf("error: %v", err)

		return err
	}

	v, ok := resp["error"]
	if ok && v == "Task with the same requestid already exists." {
		klog.Error("error: %v", v)

		return err
	}

	tmp_backups, ok := resp["backups"]

	if !ok {
		return errors.New("backups key not found")
	}

	backups := reflect.ValueOf(tmp_backups)

	backupList := make(map[int]map[string]interface{})
	backupIds := make([]int, 0)

	for i := 0; i < backups.Len(); i++ {
		backup := backups.Index(i).Interface().(map[string]interface{})

		backupId := int(backup["backupId"].(float64))

		backupList[backupId] = backup

		backupIds = append(backupIds, backupId)
	}

	sort.Ints(backupIds[:])

	t := prettytable.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(prettytable.Row{"#", "Config Name", "Alias", "Backup Time"})

	for i := 0; i < len(backupIds); i++ {
		backup := backupList[i]

		t.AppendRow(prettytable.Row{backup["backupId"], backup["collection.configName"], backup["collectionAlias"], backup["startTime"]})

	}

	fmt.Sprintf("Backup List of %v", col)
	t.Render()

	return nil
}

func BackupListAll(config Config, colId int64) error {
	for colId, _ := range config.Collections {
		if err := BackupList(config, int64(colId)); err != nil {
			return err
		}
	}

	return nil
}

func StartBackup(config Config, colId, reqId int64) error {

	col := config.Collections[colId]

	backup_uri := fmt.Sprintf("%s%s?action=BACKUP&async=sb-%d&collection=%s&name=%s&location=%s&incremental=true", config.SolrEndpoint, collection_api, reqId, col, col, config.Location)
	klog.V(5).Infof("backup uri: %v", backup_uri)

	resp, err := sendRequest(backup_uri)

	if err != nil {
		klog.Errorf("error: %v", err)

		return err
	}

	v, ok := resp["error"]
	if ok && v == "Task with the same requestid already exists." {
		klog.Error("error: %v", v)

		return err
	}

	return nil
}

func Backup(config Config, colId int64) error {
	reqId := time.Now().UnixMilli()

	if err := StartBackup(config, colId, reqId); err != nil {
		return err
	}

	if err := waitRequestStatus(config, reqId); err != nil {
		return err
	}

	if err := deleteRequestId(config, reqId); err != nil {
		return err
	}

	return nil
}

func BackupAll(config Config) error {
	for colId, _ := range config.Collections {
		if err := Backup(config, int64(colId)); err != nil {
			return err
		}
	}

	return nil
}
