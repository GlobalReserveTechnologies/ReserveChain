package econ

import "sync"

type ProfileMode string

const (
    ProfileStable   ProfileMode = "stable"
    ProfileMMF      ProfileMode = "mmf"
    ProfileCorridor ProfileMode = "corridor"
)

var (
    profileMu     sync.RWMutex
    activeProfile ProfileMode = ProfileStable
)

func GetProfile() ProfileMode {
    profileMu.RLock()
    defer profileMu.RUnlock()
    return activeProfile
}

func SetProfile(p ProfileMode) {
    profileMu.Lock()
    defer profileMu.Unlock()
    activeProfile = p
}
