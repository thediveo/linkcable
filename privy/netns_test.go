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
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/thediveo/spacetest/netns"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("netkitty network namespaces", Ordered, func() {

	BeforeEach(func() {
		goodgos := Goroutines()
		goodfds := Filedescriptors()
		Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Second).
			ShouldNot(HaveLeaked(goodgos))
		Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
	})

	When("creating network namespaces", func() {

		It("reports when it can't determine the current netns", func() {
			currentNetnsIno := netns.CurrentIno()
			Expect(newNetns("/tmp/nowhere")).Error().To(HaveOccurred())
			Expect(netns.CurrentIno()).To(Equal(currentNetnsIno))
		})

		It("reports when it can't unshare a new netns", func() {
			if os.Getuid() == 0 {
				Skip("needs non-root")
			}
			currentNetnsIno := netns.CurrentIno()
			Expect(newNetns("/proc")).Error().To(HaveOccurred())
			Expect(netns.CurrentIno()).To(Equal(currentNetnsIno))
		})

		It("successfully creates a new network namespace", func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
			currentNetnsIno := netns.CurrentIno()
			netnsfd := Successful(newNetns("/proc"))
			DeferCleanup(unix.Close, netnsfd)
			Expect(netns.Ino(netnsfd)).NotTo(Equal(currentNetnsIno))
			Expect(netns.CurrentIno()).To(Equal(currentNetnsIno))
		})

	})

	When("removing bind-mounted network namespaces", func() {

		It("succeeds when there is nothing to be removed", func() {
			Expect(RemoveBindmountedNetns("/foo/bar")).To(Succeed())
		})

		It("removes an ordinary file", func() {
			const path = "/tmp/linkcable-bindmounted-netns"
			Expect(os.WriteFile(path, nil, 0755)).To(Succeed())
			Expect(RemoveBindmountedNetns(path)).To(Succeed())
			Expect(path).NotTo(BeAnExistingFile())
		})

		It("reports when it cannot unbind", func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}

			const path = "/tmp/linkcable-bindmounted-netns"
			runtime.LockOSThread()
			defer func() {
				Expect(syscall.Seteuid(0)).To(Succeed())
				Expect(RemoveBindmountedNetns(path)).To(Succeed())
				runtime.UnlockOSThread()
			}()

			Expect(os.WriteFile(path, nil, 0755)).To(Succeed())
			Expect(unix.Mount("/proc/thread-self/ns/net", path, "", unix.MS_BIND|unix.MS_RDONLY, "")).To(Succeed())
			Expect(netns.Ino(path)).NotTo(BeZero()) // will fail if not a bind-mounted netns

			Expect(syscall.Seteuid(65534)).To(Succeed())
			Expect(os.Geteuid()).To(Equal(65534))
			Expect(RemoveBindmountedNetns(path)).NotTo(Succeed())
		})

		It("removes a bind-mount as well as the ordinary file beneath it", func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}

			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			const path = "/tmp/linkcable-bindmounted-netns"
			Expect(os.WriteFile(path, nil, 0755)).To(Succeed())
			Expect(unix.Mount("/proc/thread-self/ns/net", path, "", unix.MS_BIND|unix.MS_RDONLY, "")).To(Succeed())
			Expect(netns.Ino(path)).NotTo(BeZero()) // will fail if not a bind-mounted netns
			Expect(RemoveBindmountedNetns(path)).To(Succeed())
			Expect(path).NotTo(BeAnExistingFile())
		})

	})

	When("ensuring to (re)use its own network namespace", Ordered, func() {

		const path = "/tmp/linkcable-bindmounted-netns"

		When("not being rude", func() {

			BeforeEach(func() {
				if os.Getuid() == 0 {
					Skip("don't be root")
				}
			})

			It("reports when in the expected place is something that's not a bind-mounted netns", func() {
				Expect(BindmountedNetns("/")).Error().To(HaveOccurred())
			})

			It("reports when the path cannot be created", func() {
				Expect(BindmountedNetns("/foo/bar")).Error().To(HaveOccurred())
			})

		})

		When("running as root", func() {

			BeforeEach(func() {
				if os.Getuid() != 0 {
					Skip("needs root")
				}

				// Because of the preceeding (ordered) specs we know at this point
				// that removeBindmountedNetns behaves (mostly) correct, so let's
				// simply build upon this understanding to simply things...
				Expect(RemoveBindmountedNetns(path)).To(Succeed())
				DeferCleanup(RemoveBindmountedNetns, path)
			})

			It("creates one the first time", func() {
				netnsfd := Successful(BindmountedNetns(path))
				DeferCleanup(unix.Close, netnsfd)
				Expect(netns.Ino(netnsfd)).NotTo(Equal(netns.CurrentIno()))
			})

			It("reuses an existing one", func() {
				netnsfd := Successful(BindmountedNetns(path))
				DeferCleanup(unix.Close, netnsfd)
				netnsIno := netns.Ino(netnsfd)
				Expect(unix.Close(netnsfd)).To(Succeed())

				netnsfd = Successful(BindmountedNetns(path))
				Expect(netns.Ino(netnsfd)).To(Equal(netnsIno))
			})

		})

	})

})
