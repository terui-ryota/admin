package adapter

import (
	"context"

	"github.com/terui-ryota/admin/internal/app/admin-web/dto"
	"github.com/terui-ryota/admin/internal/domain/model"
	"github.com/terui-ryota/protofiles/go/common"
	"github.com/terui-ryota/protofiles/go/offer_item"
)

type OfferItemAdapter interface {
	ListOfferItems(ctx context.Context, limit, offset uint) ([]*offer_item.OfferItem, *common.ListResult, error)
	GetStageAssigneeCountMap(ctx context.Context, offerItemID model.OfferItemID) (map[offer_item.Stage]int, error)
	SearchOfferItems(ctx context.Context, nameQuery *string, itemID *model.ItemID, dfItemID *model.DfItemID, limit, offset uint) ([]*offer_item.OfferItem, *common.ListResult, error)
	GetOfferItem(ctx context.Context, offerItemID model.OfferItemID) (*offer_item.OfferItem, error)
	GetQuestionnaire(ctx context.Context, offerItemID model.OfferItemID) (*offer_item.Questionnaire, error)

	ListAssignees(ctx context.Context, offerItemID model.OfferItemID, stage offer_item.Stage) ([]*offer_item.Assignee, error)
	ListAssigneesUnderExamination(ctx context.Context) ([]*offer_item.Assignee, error)
	CreateOfferItem(ctx context.Context, form dto.OfferItemForm) (*model.OfferItemID, error)
	UpdateOfferItem(ctx context.Context, offerItemID model.OfferItemID, form dto.OfferItemForm) error
	DeleteOfferItem(ctx context.Context, offerItemID model.OfferItemID) error
	InviteOffer(ctx context.Context, offerItemID model.OfferItemID) error
	SaveLotteryResults(ctx context.Context, offerItemID model.OfferItemID, results []dto.LotteryResultForm) error
	CloseOffer(ctx context.Context, offerItemID model.OfferItemID) error
	BulkGetExaminations(ctx context.Context, offerItemID model.OfferItemID, entryType offer_item.EntryType) (map[model.AmebaID]*offer_item.Examination, error)
	SaveExaminationResults(ctx context.Context, offerItemID model.OfferItemID, entryType offer_item.EntryType, forms []dto.ExaminationResultForm) error
	SavePaymentResults(ctx context.Context, offerItemID model.OfferItemID, amebaIDs []model.AmebaID) error
	BulkGetQuestionnaireQuestionAnswers(ctx context.Context, offerItemID model.OfferItemID, amebaIDs []model.AmebaID) (map[model.AmebaID]*offer_item.BulkGetQuestionnaireQuestionAnswersResponse_QuestionAnswerMap, error)
	FinishShipment(ctx context.Context, offerItemID model.OfferItemID) error
	SendRemindMail(ctx context.Context, offerItemID model.OfferItemID, stage offer_item.Stage) error
}
