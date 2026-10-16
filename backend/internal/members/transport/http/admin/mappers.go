package admin

import "expensebot/internal/members/models"

func mapMemberToOut(m models.Member) memberOut {
	return memberOut{
		UserID:      m.UserID,
		Username:    m.Username,
		FirstName:   m.FirstName,
		Name:        m.Name,
		DisplayName: m.DisplayName(),
		Status:      string(m.Status),
		CreatedAt:   m.CreatedAt,
	}
}

func mapUpdateIn(in updateMemberIn) models.UpdateMemberIn {
	out := models.UpdateMemberIn{Name: in.Name}
	if in.Status != nil {
		s := models.Status(*in.Status)
		out.Status = &s
	}

	return out
}
