package documents

import "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"

type PollVoteDocument struct {
	ID     string `bson:"id" json:"id"`
	Choice string `bson:"choise" json:"choise"`
	Time   string `bson:"time" json:"time"`
}

type PollDocument struct {
	Guild           string             `bson:"guild" json:"guild"`
	MessageID       string             `bson:"messageid" json:"messageid"`
	Question        string             `bson:"question" json:"question"`
	CreateMemberID  string             `bson:"create_member_id" json:"create_member_id"`
	ManyChoose      int                `bson:"many_choose" json:"many_choose"`
	CanChangeChoose bool               `bson:"can_change_choose" json:"can_change_choose"`
	CanSeeResult    bool               `bson:"can_see_result" json:"can_see_result"`
	End             bool               `bson:"end" json:"end"`
	Anonymous       bool               `bson:"anonymous" json:"anonymous"`
	ChooseData      []string           `bson:"choose_data" json:"choose_data"`
	JoinMember      []PollVoteDocument `bson:"join_member" json:"join_member"`
}

func PollDocumentFromDomain(poll domain.Poll) PollDocument {
	votes := make([]PollVoteDocument, 0, len(poll.Votes))
	for _, vote := range poll.Votes {
		votes = append(votes, PollVoteDocument{
			ID:     vote.UserID,
			Choice: vote.Choice,
			Time:   vote.Time,
		})
	}
	return PollDocument{
		Guild:           poll.GuildID,
		MessageID:       poll.MessageID,
		Question:        poll.Question,
		CreateMemberID:  poll.CreatorID,
		ManyChoose:      poll.MaxChoices,
		CanChangeChoose: poll.CanChangeChoice,
		CanSeeResult:    poll.CanSeeResult,
		End:             poll.Ended,
		Anonymous:       poll.Anonymous,
		ChooseData:      append([]string(nil), poll.Choices...),
		JoinMember:      votes,
	}
}

func (d PollDocument) ToDomain() domain.Poll {
	votes := make([]domain.PollVote, 0, len(d.JoinMember))
	for _, vote := range d.JoinMember {
		votes = append(votes, domain.PollVote{
			UserID: vote.ID,
			Choice: vote.Choice,
			Time:   vote.Time,
		})
	}
	maxChoices := d.ManyChoose
	if maxChoices == 0 {
		maxChoices = 1
	}
	return domain.Poll{
		GuildID:         d.Guild,
		MessageID:       d.MessageID,
		Question:        d.Question,
		CreatorID:       d.CreateMemberID,
		MaxChoices:      maxChoices,
		CanChangeChoice: d.CanChangeChoose,
		CanSeeResult:    d.CanSeeResult,
		Ended:           d.End,
		Anonymous:       d.Anonymous,
		Choices:         append([]string(nil), d.ChooseData...),
		Votes:           votes,
	}
}
