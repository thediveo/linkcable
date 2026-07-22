// Copyright 2026 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package privy

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/thediveo/linkcable/privy/internal/nsops"
)

// BindmountedNetns returns a file descriptor (number) referencing an isolated
// network namespace; it either picks up an already existing bind-mounted
// network namespace at path, or creates a new network namespace and bind-mount,
// if necessary.
func BindmountedNetns(path string) (int, error) {
	netnsfd, err := unix.Open(path, os.O_RDONLY, 0)
	if err == nil {
		nstype, err := unix.IoctlRetInt(netnsfd, nsops.NS_GET_NSTYPE)
		if err == nil && nstype == unix.CLONE_NEWNET {
			return netnsfd, nil
		}
		_ = unix.Close(netnsfd)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return 0, fmt.Errorf("cannot create directory for bind-mount; reason: %w", err)
	}
	_ = os.Remove(path)
	if err := os.WriteFile(path, nil, 0644); err != nil {
		return 0, fmt.Errorf("cannot create file for bind-mount; reason: %w", err)
	}
	netnsfd, err = newNetns("/proc")
	if err != nil {
		return 0, fmt.Errorf("cannot create new netns, reason: %w", err)
	}
	if err := unix.Mount(fmt.Sprintf("/proc/self/fd/%d", netnsfd), path, "", unix.MS_BIND|unix.MS_RDONLY, ""); err != nil {
		_ = unix.Close(netnsfd)
		return 0, fmt.Errorf("cannot bind-mount newly created netns, reason: %w", err)
	}
	return netnsfd, nil
}

// RemoveBindmountedNetns removes the bind-mount and its underlying regular file
// at path, returning nil if it succeeds; otherwise, if it fails, it returns an
// error.
//
// Please note that this will not automatically destroy the originally
// bind-mounted network namespace until any other references to it cease to
// exist.
func RemoveBindmountedNetns(path string) error {
	switch err := unix.Unlink(path); err {
	case nil, syscall.ENOENT:
		return nil
	case syscall.EBUSY:
		if err := unix.Unmount(path, 0); err != nil {
			return err
		}
		return unix.Unlink(path)
	default:
		return err
	}
}

// newNetns returns a file descriptor (number) referencing a newly created
// (“unshared”) network namespace; otherwise, it returns an error. The caller is
// responsible to manage the returned file descriptor.
func newNetns(procpath string) (int, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// The usual dance: first, get a reference to the network namespace our
	// OS-level thread currently is attached to; we will later need it in order
	// to restore this thread's original network namespace attachment.
	netnsfd, err := unix.Open(procpath+"/thread-self/ns/net", os.O_RDONLY, 0)
	if err != nil {
		return 0, fmt.Errorf("cannot determine current netns, reason: %w", err)
	}
	// Next, attach our OS-level thread to a newly created network namespace.
	// Note that there isn't any syscall to create an unattached network
	// namespace.
	if err := unix.Unshare(unix.CLONE_NEWNET); err != nil {
		_ = unix.Close(netnsfd)
		return 0, fmt.Errorf("cannot unshare new netns, reason: %w", err)
	}
	// Pick up a reference to the newly created and currently attached network
	// namespace; we will later return this reference and thus keep the network
	// namespace alive by keeping it referenced by an open file descriptor.
	newnetnsfd, err := unix.Open(procpath+"/thread-self/ns/net", os.O_RDONLY, 0)
	if err != nil {
		_ = unix.Close(netnsfd)
		return 0, fmt.Errorf("cannot determine new netns, reason: %w", err)
	}
	// Finally, attach our OS-level thread back to its original network
	// namespace and be done.
	if err := unix.Setns(netnsfd, unix.CLONE_NEWNET); err != nil {
		_ = unix.Close(netnsfd)
		_ = unix.Close(newnetnsfd)
		return 0, fmt.Errorf("cannot return to previous netns, reason: %w", err)
	}
	return newnetnsfd, nil
}
