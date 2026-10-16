package bot

import (
	"expensebot/internal/app/telegram"
	membersModels "expensebot/internal/members/models"
)

func mapUserToMember(u telegram.User) membersModels.Member {
	return membersModels.Member{UserID: u.ID, Username: u.Username, FirstName: u.FirstName}
}
