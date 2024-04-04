//
// Copyright Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/devfile/library/v2/pkg/util"
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/assert"
)

const (
	RawGitHubHost string = "raw.githubusercontent.com"
)

func TestDownloadInMemoryClient(t *testing.T) {
	const downloadErr = "failed to retrieve %s, 404: Not Found"
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		_, err := rw.Write([]byte("OK"))
		if err != nil {
			t.Error(err)
		}
	}))

	// Close the server when test finishes
	defer server.Close()

	devfileUtilsClient := NewDevfileUtilsClient()

	tests := []struct {
		name       string
		url        string
		token      string
		client     DevfileUtils
		want       []byte
		wantParent []byte
		wantErr    string
	}{
		{
			name:   "Case 1: Input url is valid",
			client: devfileUtilsClient,
			url:    server.URL,
			want:   []byte{79, 75},
		},
		{
			name:    "Case 2: Input url is invalid",
			client:  devfileUtilsClient,
			url:     "invalid",
			wantErr: "unsupported protocol scheme",
		},
		{
			name:    "Case 3: Git provider with invalid url",
			client:  devfileUtilsClient,
			url:     "github.com/mike-hoang/invalid-repo",
			token:   "",
			want:    []byte(nil),
			wantErr: "failed to parse git repo. error:*",
		},
		{
			name:    "Case 4: Public Github repo with missing blob",
			client:  devfileUtilsClient,
			url:     "https://github.com/devfile/library/main/README.md",
			wantErr: "failed to parse git repo. error: url path to directory or file should contain 'tree' or 'blob'*",
		},
		{
			name:    "Case 5: Public Github repo, with invalid token ",
			client:  devfileUtilsClient,
			url:     "https://github.com/devfile/library/blob/main/devfile.yaml",
			token:   "fake-token",
			wantErr: fmt.Sprintf(downloadErr, "https://"+RawGitHubHost+"/devfile/library/main/devfile.yaml"),
		},
		{
			name:   "Case 6: Input url is valid with a mock client, dont use mock data during invocation",
			client: &MockDevfileUtilsClient{},
			url:    server.URL,
			want:   []byte{79, 75},
		},
		{
			name:   "Case 7: Input url is valid with a mock client and mock token",
			client: &MockDevfileUtilsClient{MockGitURL: util.MockGitUrl{Host: "https://github.com/devfile/library/blob/main/devfile.yaml"}, GitTestToken: "valid-token", DownloadOptions: util.MockDownloadOptions{MockFile: "OK"}},
			url:    "https://github.com/devfile/library/blob/main/devfile.yaml",
			want:   []byte{79, 75},
		},
		{
			name:    "Case 8: Public Github repo, with invalid token ",
			client:  &MockDevfileUtilsClient{MockGitURL: util.MockGitUrl{Host: "https://github.com/devfile/library/blob/main/devfile.yaml"}, GitTestToken: "invalid-token"},
			url:     "https://github.com/devfile/library/blob/main/devfile.yaml",
			wantErr: "failed to retrieve https://github.com/devfile/library/blob/main/devfile.yaml",
		},
		{
			name:   "Case 9: Input github url is valid with a mock client, dont use mock data during invocation",
			client: &MockDevfileUtilsClient{},
			url:    "https://raw.githubusercontent.com/maysunfaisal/OK/main/OK.txt",
			want:   []byte{79, 75},
		},
		{
			name:       "Case 10: Test devfile with private parent",
			client:     &MockDevfileUtilsClient{},
			url:        "https://github.com/devfile/library/blob/main/devfile.yaml",
			token:      "parent-devfile",
			want:       []byte(util.MockDevfileWithParentRef),
			wantParent: []byte(util.MockParentDevfile),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.client.DownloadInMemory(util.HTTPRequestParams{URL: tt.url, Token: tt.token})
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("Failed to download file with error: %s", err)
			} else if err == nil && !reflect.DeepEqual(data, tt.want) {
				t.Errorf("Expected: %v, received: %v, difference at %v", string(tt.want), string(data[:]), pretty.Compare(tt.want, data))
			} else if err != nil {
				assert.Regexp(t, tt.wantErr, err.Error(), "Error message should match")
			}

			if len(tt.wantParent) > 0 {
				data, err := tt.client.DownloadInMemory(util.HTTPRequestParams{URL: tt.url, Token: tt.token})
				if (err != nil) != (tt.wantErr != "") {
					t.Errorf("Failed to download file with error: %s", err)
				} else if err == nil && !reflect.DeepEqual(data, tt.wantParent) {
					t.Errorf("Expected: %v, received: %v, difference at %v", string(tt.wantParent), string(data[:]), pretty.Compare(tt.wantParent, data))
				} else if err != nil {
					assert.Regexp(t, tt.wantErr, err.Error(), "Error message should match")
				}
			}
		})
	}
}

func TestValidateDevfileExistence(t *testing.T) {

	tests := []struct {
		name          string
		url           string
		wantErr       bool
		expectedValue bool
	}{
		{
			name:          "recognizes devfile.yaml",
			url:           "https://dummyurlpath/devfile/registry/main/stacks/python/3.0.0/devfile.yaml",
			wantErr:       false,
			expectedValue: true,
		},
		{
			name:          "recognizes devfile.yml",
			url:           "https://dummyurlpath/devfile/registry/main/stacks/python/3.0.0/devfile.yml",
			wantErr:       false,
			expectedValue: true,
		},
		{
			name:          "recognizes .devfile.yaml",
			url:           "https://dummyurlpath/devfile/registry/main/stacks/python/3.0.0/.devfile.yaml",
			wantErr:       false,
			expectedValue: true,
		},
		{
			name:          "recognizes .devfile.yml",
			url:           "https://dummyurlpath/devfile/registry/main/stacks/python/3.0.0/.devfile.yml",
			wantErr:       false,
			expectedValue: true,
		},
		{
			name:          "no valid devfile in path",
			url:           "https://dummyurlpath/devfile/registry/main/stacks/python/3.0.0/deploy.yaml",
			wantErr:       true,
			expectedValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ValidateDevfileExistence(tt.url)
			assert.EqualValues(t, tt.expectedValue, res, "expected res = %t, got %t", tt.expectedValue, res)
		})
	}
}
