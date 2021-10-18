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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func IsFileNotExistError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(strings.ToLower(err.Error()), "no such file or directory")
}

func FileOverwrite(fileName string, content []byte) error {
	dir := path.Dir(fileName)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("create dir fail, dir: %s, err: %v", dir, err)
		}
	}

	if err := ioutil.WriteFile(fileName, content, os.ModePerm); err != nil {
		return fmt.Errorf("overwrite file fail, file: %s, err: %v", fileName, err)
	}

	return nil
}

func FileCopyRecursive(from, to string) error {
	if bs, err := exec.Command("cp", "-rf", from, to).CombinedOutput(); err != nil {
		return fmt.Errorf("FileCopyRecursive fail, from: %s, to: %s, err: %s", from, to, bytes.Trim(bs, "\r\n"))
	}

	return nil
}

// RenameFileIfNotLinkFile rename oldPath to newPath then link oldPath to newPath if oldPath is not a link file
// if file is link, do nothing
// else rename it by newPath then link it by oldPath
func RenameFileIfNotLinkFile(oldPath, newPath string) error {
	originPath, err := filepath.EvalSymlinks(oldPath)
	if err != nil {
		return err
	}

	// link file, do nothing
	if originPath != oldPath {
		return nil
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("rename fail, oldPath: %s, newPath: %s, err: %v", oldPath, newPath, err)
	}

	return FileLink(newPath, oldPath)
}

func FileLink(target, linkName string) error {
	if err := exec.Command("ln", "-sf", target, linkName).Run(); err != nil {
		return fmt.Errorf("ln -sf %s, %s fail, err: %v", target, linkName, err)
	}

	return nil
}
