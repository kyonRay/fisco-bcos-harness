// Package schema fixes the team smart-sheet layout the spec settled:
// column titles and the sub-requirement status enum. Skills and CLI
// share these strings so the sheet stays the single source of truth.
package schema

// Sub-requirement sheet column titles.
const (
	ColRequirement = "需求名"
	ColParent      = "所属总需求"
	ColMilestone   = "milestone"
	ColOwner       = "认领人"
	ColStatus      = "状态"
	ColPRLink      = "PR链接"
	ColNotes       = "备注"
)

// Milestone sheet column titles.
const (
	ColMilestoneName  = "名称"
	ColMilestoneOwner = "负责人"
	ColGateStatus     = "门禁状态"
	ColGateTime       = "门禁时间"
)

// Statuses is the sub-requirement lifecycle, in order.
var Statuses = []string{
	"待认领", "开发中", "待review", "review循环", "人工review", "已合入", "完成",
}

// ValidStatus reports whether s is one of Statuses.
func ValidStatus(s string) bool {
	for _, v := range Statuses {
		if v == s {
			return true
		}
	}
	return false
}
