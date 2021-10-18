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

package file_store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baidu/conf-agent/xfile"
	"github.com/baidu/conf-agent/xlog"
)

type FileStore struct {
	// ConfDir is the root dir of conf file
	ConfDir string
	// CoypFiles is list of files and directories copied from default dir to tmp dir
	CopyFiles []string
}

// compose path of tempory directory to store files
func (fileStore *FileStore) tmpDir(version string) string {
	return fileStore.ConfDir + "_" + version
}

func NewFileStore(confDir string, copyFiles []string) (*FileStore, error) {
	return &FileStore{
		ConfDir:   confDir,
		CopyFiles: copyFiles,
	}, nil
}

// UpdateDefaultConfDir updates default config directory with config files in tempory directory.
func (fileStore *FileStore) UpdateDefaultConfDir(ctx context.Context, version string) error {
	dest, err := filepath.EvalSymlinks(fileStore.ConfDir)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(fileStore.ConfDir); err != nil {
		err = fmt.Errorf("file: %s, err: %v", fileStore.ConfDir, err)
		xlog.Default.Error(xlog.ErrLogFormat(ctx, "UpdateDefaultConfDir.Remove", err))

		return err
	}

	// delete the target file if it's a link file
	if dest != fileStore.ConfDir {
		if err := os.RemoveAll(dest); err != nil {
			err = fmt.Errorf("file: %s, err: %v", dest, err)
			xlog.Default.Error(xlog.ErrLogFormat(ctx, "UpdateDefaultConfDir.Remove", err))

			return err
		}
	}

	// ln -sf ModDemo_{version} ModDemo
	// NOTICE: if link fail, bfe can't restart automatically !!!
	if err := xfile.FileLink(fileStore.tmpDir(version), fileStore.ConfDir); err != nil {
		xlog.Default.Error(xlog.ErrLogFormat(ctx, "UpdateDefaultConfDir.FileLink", err))

		return err
	}

	return nil
}

// StoreFile2TmpDir store all file to tempory directory
// it will create new file or overwrite old file
func (fileStore *FileStore) StoreFile2TmpDir(ctx context.Context, version string, files map[string][]byte) error {
	tmpDir := fileStore.tmpDir(version)

	// delete tmp directory if exist
	if err := os.RemoveAll(tmpDir); err != nil && !xfile.IsFileNotExistError(err) {
		err = fmt.Errorf("RemoveAll fail, dir: %s, err: %v", tmpDir, err)
		xlog.Default.Error(xlog.ErrLogFormat(ctx, "fileStore.RemoveAll", err))

		return err
	}

	// create tmp directory
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		err = fmt.Errorf("MkDirAll fail, dir: %s, err: %v", tmpDir, err)
		xlog.Default.Error(xlog.ErrLogFormat(ctx, "fileStore.MkdirAll", err))

		return err
	}

	// copy config files (listed in fileStore.CopyFiles) from default dir to tmp dir
	for _, copyFile := range fileStore.CopyFiles {
		file := filepath.Join(fileStore.ConfDir, copyFile)
		if err := xfile.FileCopyRecursive(file, tmpDir); err != nil {
			if xfile.IsFileNotExistError(err) {
				xlog.Default.Info(xlog.ErrLogFormat(ctx, "fileStore.CopyFiles", err))
				continue
			}

			err = fmt.Errorf("keepFile fail, file: %s, err: %v", file, err)
			xlog.Default.Error(xlog.ErrLogFormat(ctx, "fileStore.CopyFiles", err))

			return err
		}
	}

	// write content to file
	for fileName, fileContent := range files {
		if err := xfile.FileOverwrite(filepath.Join(tmpDir, fileName), fileContent); err != nil {
			xlog.Default.Error(xlog.ErrLogFormat(ctx, "fileStore.FileOverwrite", err))
			return err
		}

		// xlog.Default.Debug(xlog.InfoLogFormat(ctx, "fileStore.FileOverwrite", "fileName: ", fileName,
		// 	" fileContent: ", string(fileContent)))
	}

	return nil
}
