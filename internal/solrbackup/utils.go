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
	"encoding/json"
	"fmt"
	"io/ioutil"
	klog "k8s.io/klog/v2"
	"net/http"
	"time"
)

const (
	collection_api string = "/solr/admin/collections"
)

func sendRequest(uri string) (map[string]interface{}, error) {
	resp, err := http.Get(uri)

	if err != nil {
		klog.Errorf("error while get reqeust: %v", err)

		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Error("error while reading response: %v", err)

		return nil, err
	}

	klog.V(5).Infof("body: %v", string(body))

	var resp_json map[string]interface{}

	if err := json.Unmarshal(body, &resp_json); err != nil {
		klog.Errorf("error while reading response: %v", err)

		return nil, err
	}

	return resp_json, nil
}

func waitRequestStatus(config Config, reqId int64) error {
	for {
		reqstatus_uri := fmt.Sprintf("%s%s?action=REQUESTSTATUS&requestid=sb-%d", config.SolrEndpoint, collection_api, reqId)
		klog.V(5).Infof("wrs uri: %v", reqstatus_uri)

		resp, err := sendRequest(reqstatus_uri)

		if err != nil {
			klog.Errorf("error: %v", err)

			return err
		}

		klog.V(5).Infof("request status response %v", resp)

		state, _ := resp["status"].(map[string]interface{})["state"]

		if state == "running" || state == "submitted" {
			time.Sleep(time.Second * 5)
			continue
		} else if state == "completed" {
			break
		} else {
			return fmt.Errorf("unknown state: %v", state)
		}
	}

	return nil
}

func deleteRequestId(config Config, reqId int64) error {
	delete_uri := fmt.Sprintf("%s%s?action=DELETESTATUS&requestid=sb-%d", config.SolrEndpoint, collection_api, reqId)
	klog.V(5).Infof("delete uri: %v", delete_uri)

	resp, err := sendRequest(delete_uri)

	if err != nil {
		klog.Errorf("error: %v", err)

		return err
	}

	klog.V(5).Infof("delete response %v", resp)

	return nil
}
