//go:build !windows && !darwin && !linux

package main

import "errors"

func verifyUnixNoExtendedACL(_ string) error {
	return errors.New("this Unix platform does not expose a supported extended-ACL verification contract")
}
