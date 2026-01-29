package net

type LeaderSelector struct {
    NodeIDs []string
}

func NewLeaderSelector(ids []string) *LeaderSelector {
    return &LeaderSelector{NodeIDs: ids}
}

func (ls *LeaderSelector) LeaderForTick(tick uint64) string {
    if len(ls.NodeIDs) == 0 {
        return ""
    }
    idx := int(tick % uint64(len(ls.NodeIDs)))
    return ls.NodeIDs[idx]
}
