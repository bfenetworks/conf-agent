// Copyright (c) 2021 The BFE Authors.
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

package xfile

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRenameFileIfNotLinkFile(t *testing.T) {
	dir := "./testdata/link"
	defer FileOverwrite(dir+"/.gitkeep", []byte{})

	tests := []struct {
		name string

		oldPath string
		newPath string

		beforeCmd func() error
		wantErr   bool
	}{
		{
			name:    "case_old_path_no_exist",
			oldPath: "no_exsit",
			wantErr: true,
		},
		{
			name:    "case_link_file",
			oldPath: "link_file",
			// newPath: "link_file_01",
			beforeCmd: func() error {
				if err := FileOverwrite(filepath.Join(dir, "link_file_real/xxxx.data"), []byte("xxx")); err != nil {
					return err
				}

				cmd := exec.Command("ln", "-sf", "link_file_real", "link_file")
				cmd.Dir = dir

				return cmd.Run()
			},
			wantErr: false,
		},
		{
			name:    "case_normal_file",
			oldPath: "normal_file",
			newPath: "normal_file_01",
			beforeCmd: func() error {
				if err := FileOverwrite(filepath.Join(dir, "normal_file/xxxx.data"), []byte("xxx")); err != nil {
					return err
				}

				return nil
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		if tt.beforeCmd != nil {
			if err := tt.beforeCmd(); err != nil {
				t.Errorf("%s: %v", tt.name, err)
				return
			}
		}

		if err := RenameFileIfNotLinkFile(filepath.Join(dir, tt.oldPath), filepath.Join(dir, tt.newPath)); (err != nil) != tt.wantErr {
			t.Errorf("name: %s RenameFileIfNotLinkFile() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}

		if err := exec.Command("rm", "-rf", dir).Run(); err != nil {
			t.Error(err)
			return
		}
	}
}
