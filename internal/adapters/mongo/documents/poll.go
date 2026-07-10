package documents

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

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

type PollReadDocument struct {
	Guild           bson.RawValue `bson:"guild" json:"guild"`
	MessageID       bson.RawValue `bson:"messageid" json:"messageid"`
	Question        bson.RawValue `bson:"question" json:"question"`
	CreateMemberID  bson.RawValue `bson:"create_member_id" json:"create_member_id"`
	ManyChoose      bson.RawValue `bson:"many_choose" json:"many_choose"`
	CanChangeChoose bson.RawValue `bson:"can_change_choose" json:"can_change_choose"`
	CanSeeResult    bson.RawValue `bson:"can_see_result" json:"can_see_result"`
	End             bson.RawValue `bson:"end" json:"end"`
	Anonymous       bson.RawValue `bson:"anonymous" json:"anonymous"`
	ChooseData      bson.RawValue `bson:"choose_data" json:"choose_data"`
	JoinMember      bson.RawValue `bson:"join_member" json:"join_member"`
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

func (d PollReadDocument) ToDomain() domain.Poll {
	guild, _ := legacyMongooseString(d.Guild)
	messageID, _ := legacyMongooseString(d.MessageID)
	question, _ := legacyMongooseString(d.Question)
	creatorID, _ := legacyMongooseString(d.CreateMemberID)
	maxChoices, ok := LegacyExactInt64(d.ManyChoose)
	if !ok || maxChoices < 1 {
		maxChoices = 1
	}
	return domain.Poll{
		GuildID:         guild,
		MessageID:       messageID,
		Question:        question,
		CreatorID:       creatorID,
		MaxChoices:      int(maxChoices),
		CanChangeChoice: legacyMongooseBoolean(d.CanChangeChoose),
		CanSeeResult:    legacyMongooseBoolean(d.CanSeeResult),
		Ended:           legacyMongooseBoolean(d.End),
		Anonymous:       legacyMongooseBoolean(d.Anonymous),
		Choices:         legacyPollChoices(d.ChooseData),
		Votes:           legacyPollVotes(d.JoinMember),
	}
}

func legacyPollChoices(value bson.RawValue) []string {
	values := legacyPollArrayValues(value)
	choices := make([]string, 0, len(values))
	for _, candidate := range values {
		if choice, ok := candidate.StringValueOK(); ok {
			choices = append(choices, choice)
		}
	}
	return choices
}

func legacyPollVotes(value bson.RawValue) []domain.PollVote {
	values := legacyPollArrayValues(value)
	votes := make([]domain.PollVote, 0, len(values))
	for _, candidate := range values {
		document, ok := candidate.DocumentOK()
		if !ok {
			continue
		}
		userID, userOK := document.Lookup("id").StringValueOK()
		choice, choiceOK := document.Lookup("choise").StringValueOK()
		voteTime, timeOK := document.Lookup("time").StringValueOK()
		if !userOK || !choiceOK || !timeOK {
			continue
		}
		votes = append(votes, domain.PollVote{UserID: userID, Choice: choice, Time: voteTime})
	}
	return votes
}

func legacyPollArrayValues(value bson.RawValue) []bson.RawValue {
	array, ok := value.ArrayOK()
	if ok {
		values, err := array.Values()
		if err == nil {
			return values
		}
		return nil
	}
	if value.Type == 0 || value.Type == bson.TypeNull || value.Type == bson.TypeUndefined {
		return nil
	}
	return []bson.RawValue{value}
}
