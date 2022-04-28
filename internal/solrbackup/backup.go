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

func startDelete(config Config, colId, backupId, reqId int64) error {
	col := config.Collections[colId]

	var delete_uri string

	if backupId == -1 {
		delete_uri = fmt.Sprintf("%s%s?action=DELETEBACKUP&async=sb-%d&name=%s&location=%s&purgeUnused=true", config.SolrEndpoint, collection_api, reqId, col, config.Location)
	} else {
		delete_uri = fmt.Sprintf("%s%s?action=DELETEBACKUP&async=sb-%d&name=%s&location=%s&backupId=%d", config.SolrEndpoint, collection_api, reqId, col, config.Location, backupId)
	}

	klog.V(5).Infof("delete uri: %v", delete_uri)

	resp, err := sendRequest(delete_uri)

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

func backupPurgeUnused(config Config, colId int64) error {
	reqId := time.Now().UnixMilli()

	if err := startDelete(config, colId, -1, reqId); err != nil {
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

func BackupDeleteWithColIdWithBackupId(config Config, colId, backupId int64) error {
	reqId := time.Now().UnixMilli()

	if err := startDelete(config, colId, backupId, reqId); err != nil {
		return err
	}

	if err := waitRequestStatus(config, reqId); err != nil {
		return err
	}

	if err := deleteRequestId(config, reqId); err != nil {
		return err
	}

	return backupPurgeUnused(config, colId)
}

func BackupDelete(config Config, colId int64) error {
	before := time.Now().AddDate(0, 0, -1*config.RetaintionDays)

	backups, err := backupListRetrive(config, colId)

	if err != nil {
		return err
	}

	backupIds := make([]int, 0)

	for i := 0; i < backups.Len(); i++ {
		backup := backups.Index(i).Interface().(map[string]interface{})

		backupId := int(backup["backupId"].(float64))
		startTimeStr := backup["startTime"].(string)

		date, err := time.Parse("2006-01-02T15:04:05.000000Z", startTimeStr)

		if err != nil {
			klog.Errorf("cannot parse start time %v", err)

			return err
		}

		if date.Before(before) {
			backupIds = append(backupIds, backupId)
		}

	}

	for _, backupId := range backupIds {
		if err := BackupDeleteWithColIdWithBackupId(config, colId, int64(backupId)); err != nil {
			return err
		}
	}

	return nil
}

func BackupDeleteAll(config Config) error {
	for colId, _ := range config.Collections {
		if err := BackupDelete(config, int64(colId)); err != nil {
			return err
		}
	}

	return nil
}

func backupListRetrive(config Config, colId int64) (*reflect.Value, error) {
	col := config.Collections[colId]

	backup_uri := fmt.Sprintf("%s%s?action=LISTBACKUP&name=%s&location=%s", config.SolrEndpoint, collection_api, col, config.Location)
	klog.V(5).Infof("backup uri: %v", backup_uri)

	resp, err := sendRequest(backup_uri)

	if err != nil {
		klog.Errorf("error: %v", err)

		return nil, err
	}

	v, ok := resp["error"]
	if ok && v == "Task with the same requestid already exists." {
		klog.Error("error: %v", v)

		return nil, err
	}

	tmp_backups, ok := resp["backups"]

	if !ok {
		return nil, errors.New("backups key not found")
	}

	backups := reflect.ValueOf(tmp_backups)

	return &backups, nil
}

func BackupList(config Config, colId int64) error {
	backups, err := backupListRetrive(config, colId)

	if err != nil {
		return err
	}

	backupList := make(map[int]map[string]interface{})
	backupIds := make([]int, 0)

	for i := 0; i < backups.Len(); i++ {
		backup := backups.Index(i).Interface().(map[string]interface{})

		backupId := int(backup["backupId"].(float64))

		backupList[backupId] = backup

		backupIds = append(backupIds, backupId)
	}

	sort.Ints(backupIds[:])

	col := config.Collections[colId]

	t := prettytable.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(prettytable.Row{"#", "Collection", "Config Name", "Alias", "Backup Time"})

	for _, i := range backupIds {
		backup := backupList[i]

		t.AppendRow(prettytable.Row{backup["backupId"], col, backup["collection.configName"], backup["collectionAlias"], backup["startTime"]})

	}

	t.Render()

	return nil
}

func BackupListAll(config Config) error {
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
