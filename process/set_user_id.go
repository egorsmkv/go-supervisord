//go:build !windows

package process

import (
	"os/user"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func setUserID(procAttr *syscall.SysProcAttr, uid uint32, gid uint32) {
	u, err := user.Current()
	if err == nil {
		cuid, uidErr := strconv.ParseUint(u.Uid, 10, 32)
		cgid, gidErr := strconv.ParseUint(u.Gid, 10, 32)
		if uidErr == nil && gidErr == nil && uint32(cuid) == uid && uint32(cgid) == gid {
			log.Info("no need to switch user")
			return
		}
	}
	procAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, NoSetGroups: true}
}
