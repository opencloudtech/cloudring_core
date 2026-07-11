// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type parentRootPath struct {
	root *os.Root
	name string
}

func openParentRootPath(path string) (parentRootPath, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return parentRootPath{}, fmt.Errorf("resolve absolute path: %w", err)
	}
	absolute = filepath.Clean(absolute)
	parent := filepath.Dir(absolute)
	name := filepath.Base(absolute)
	if parent == absolute {
		name = "."
	}
	root, err := os.OpenRoot(parent)
	if err != nil {
		return parentRootPath{}, fmt.Errorf("open parent directory: %w", err)
	}
	return parentRootPath{root: root, name: name}, nil
}

func (path *parentRootPath) close(resultErr *error) {
	if path == nil || path.root == nil {
		return
	}
	if err := path.root.Close(); err != nil {
		*resultErr = errors.Join(*resultErr, fmt.Errorf("close parent directory: %w", err))
	}
	path.root = nil
}

func readOperatorSelectedFile(path string) (data []byte, resultErr error) {
	rooted, err := openParentRootPath(path)
	if err != nil {
		return nil, err
	}
	defer rooted.close(&resultErr)

	data, err = rooted.root.ReadFile(rooted.name)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func lstatWithinParent(path string) (info os.FileInfo, resultErr error) {
	rooted, err := openParentRootPath(path)
	if err != nil {
		return nil, err
	}
	defer rooted.close(&resultErr)

	info, err = rooted.root.Lstat(rooted.name)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func removeWithinParent(path string) (resultErr error) {
	rooted, err := openParentRootPath(path)
	if err != nil {
		return err
	}
	defer rooted.close(&resultErr)
	return rooted.root.Remove(rooted.name)
}

func renameWithinParent(source string, destination string) (resultErr error) {
	sourceAbsolute, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("resolve evidence source path: %w", err)
	}
	destinationAbsolute, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("resolve evidence destination path: %w", err)
	}
	sourceAbsolute = filepath.Clean(sourceAbsolute)
	destinationAbsolute = filepath.Clean(destinationAbsolute)
	if filepath.Dir(sourceAbsolute) != filepath.Dir(destinationAbsolute) {
		return errors.New("evidence replacement source and destination must share a parent directory")
	}

	root, err := os.OpenRoot(filepath.Dir(sourceAbsolute))
	if err != nil {
		return fmt.Errorf("open evidence replacement parent: %w", err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close evidence replacement parent: %w", err))
		}
	}()
	return root.Rename(filepath.Base(sourceAbsolute), filepath.Base(destinationAbsolute))
}

func ensureEvidenceParentDirectory(path string) (resultErr error) {
	absolute, err := canonicalEvidenceDestinationForParentCreation(path)
	if err != nil {
		return fmt.Errorf("resolve evidence destination path: %w", err)
	}
	destinationParent := filepath.Dir(filepath.Clean(absolute))
	volume := filepath.VolumeName(destinationParent)
	rootPath := volume + string(os.PathSeparator)
	if volume == "" {
		rootPath = string(os.PathSeparator)
	}
	relative, err := filepath.Rel(rootPath, destinationParent)
	if err != nil {
		return fmt.Errorf("resolve evidence parent below filesystem root: %w", err)
	}
	if relative == ".." || filepath.IsAbs(relative) || filepathHasParentTraversal(relative) {
		return errors.New("evidence parent escapes the filesystem root")
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return fmt.Errorf("open evidence filesystem root: %w", err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close evidence parent root: %w", err))
		}
	}()
	if relative == "." {
		return nil
	}

	currentPath := rootPath
	for _, component := range splitRelativePath(relative) {
		childPath := filepath.Join(currentPath, component)
		next, err := openOrCreateEvidenceDirectory(root, component, childPath)
		if err != nil {
			return err
		}
		if closeErr := root.Close(); closeErr != nil {
			root = next
			return fmt.Errorf("close evidence ancestor %q: %w", currentPath, closeErr)
		}
		root = next
		currentPath = childPath
	}
	return nil
}

func filepathHasParentTraversal(path string) bool {
	for _, component := range splitRelativePath(path) {
		if component == ".." {
			return true
		}
	}
	return false
}

func splitRelativePath(path string) []string {
	if path == "" || path == "." {
		return nil
	}
	components := make([]string, 0)
	for _, component := range strings.Split(path, string(os.PathSeparator)) {
		if component != "" && component != "." {
			components = append(components, component)
		}
	}
	return components
}

func openOrCreateEvidenceDirectory(root *os.Root, name string, displayPath string) (*os.Root, error) {
	info, err := root.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		if err := root.Mkdir(name, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("create private evidence directory %q: %w", displayPath, err)
		}
		info, err = root.Lstat(name)
	}
	if err != nil {
		return nil, fmt.Errorf("inspect evidence parent component %q: %w", displayPath, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("evidence parent component %q is a symbolic link", displayPath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("evidence parent component %q is not a directory", displayPath)
	}
	next, err := root.OpenRoot(name)
	if err != nil {
		return nil, fmt.Errorf("open evidence parent component %q: %w", displayPath, err)
	}
	openedInfo, err := next.Stat(".")
	if err != nil {
		primary := fmt.Errorf("inspect opened evidence parent component %q: %w", displayPath, err)
		if closeErr := next.Close(); closeErr != nil {
			primary = errors.Join(primary, fmt.Errorf("close unverified evidence parent component: %w", closeErr))
		}
		return nil, primary
	}
	if !os.SameFile(info, openedInfo) {
		primary := fmt.Errorf("evidence parent component %q changed while being opened", displayPath)
		if closeErr := next.Close(); closeErr != nil {
			primary = errors.Join(primary, fmt.Errorf("close changed evidence parent component: %w", closeErr))
		}
		return nil, primary
	}
	return next, nil
}
