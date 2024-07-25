package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/terui-ryota/admin/internal/app/admin-web/dto"
	"github.com/terui-ryota/admin/internal/domain/adapter"
	"github.com/terui-ryota/admin/internal/domain/model"
	"github.com/terui-ryota/admin/internal/domain/repository"
	"github.com/terui-ryota/protofiles/go/offer_item"
)

type ListAssigneesParameter struct {
	Stage                   *offer_item.Stage
	WithOfferItem           bool
	WithPersonalInformation bool
	WithQuestionAnswers     bool
	WithBlogger             bool
	WithAsID                bool
	EntryType               *offer_item.EntryType
	WithoutFutureEntry      bool
}

type Assignees []Assignee

type OfferItemUsecase interface {
	ListOfferItems(ctx context.Context, limit, offset uint) ([]OfferItem, uint, error)
	SearchOfferItems(ctx context.Context, nameQuery *string, itemID *model.ItemID, dfItemID *model.DfItemID, limit, offset uint) ([]OfferItem, uint, error)
	GetOfferItem(ctx context.Context, id model.OfferItemID) (*OfferItem, error)
	GetOfferItemForm(ctx context.Context, id model.OfferItemID) (*dto.OfferItemForm, error)
	ListAllAssignees(
		ctx context.Context,
		id model.OfferItemID,
		parameter ListAssigneesParameter,
	) (Assignees, error)
	ListAssigneesUnderExamination(ctx context.Context, limit int) ([]Assignee, error)
	CreateOfferItem(ctx context.Context, form dto.OfferItemForm) (*model.OfferItemID, error)
	UpdateOfferItem(ctx context.Context, id model.OfferItemID, form dto.OfferItemForm) error
	DeleteOfferItem(ctx context.Context, id model.OfferItemID) error
	// 参加を募集する
	Invite(ctx context.Context, id model.OfferItemID) error
	// 選考結果を保存します
	SaveLotteryResults(ctx context.Context, id model.OfferItemID, results []dto.LotteryResultForm) error
	// 下書き審査結果を保存します
	SavePreExaminationResults(ctx context.Context, id model.OfferItemID, results []dto.ExaminationResultForm) error
	// 審査結果を保存します
	SaveExaminationResults(ctx context.Context, id model.OfferItemID, results []dto.ExaminationResultForm) error
	// 支払いを完了に保存します
	SavePaymentResults(ctx context.Context, id model.OfferItemID, amebaIDs []model.AmebaID) error
	// 案件を終了します
	Close(ctx context.Context, id model.OfferItemID) error
	// ステージごとのアサイニー数を取得します
	GetStageAssigneeCountMap(ctx context.Context, id model.OfferItemID) (map[offer_item.Stage]int, error)
	// アンケート取得
	GetQuestionnaire(ctx context.Context, offerItemID model.OfferItemID) (*offer_item.Questionnaire, error)
	// 発送完了します
	FinishShipment(ctx context.Context, offerItemID model.OfferItemID) error

	SendRemindMail(ctx context.Context, offerItemID model.OfferItemID, stage offer_item.Stage) error
}

type offerItemUsecaseImpl struct {
	offerItemAdapter adapter.OfferItemAdapter
}

func NewOfferItemUsecaseImpl(
	offerItemAdapter adapter.OfferItemAdapter,
) OfferItemUsecase {
	return &offerItemUsecaseImpl{
		offerItemAdapter: offerItemAdapter,
	}
}

// GetOfferItem implements OfferItemUsecase.
func (u *offerItemUsecaseImpl) GetOfferItem(ctx context.Context, id model.OfferItemID) (*OfferItem, error) {
	res, err := u.offerItemAdapter.GetOfferItem(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("u.offerItemAdapter.GetOfferItem: %w", err)
	}
	i, l, d, pe, p, e, ap, s := convertSchedules(res.GetSchedules())
	//questionnaire, err := u.offerItemAdapter.GetQuestionnaire(ctx, id)
	//if err != nil {
	//	if !errors.Is(err, adapter.ErrAdapterNotFound) {
	//		return nil, fmt.Errorf("u.offerItemAdapter.GetQuestionnaire: %w", err)
	//	}
	//}
	return &OfferItem{
		OfferItem:               res,
		InvitationSchedule:      i,
		LotterySchedule:         l,
		ShipmentSchedule:        s,
		DraftSubmissionSchedule: d,
		PreExaminationSchedule:  pe,
		ArticlePostingSchedule:  p,
		ExaminationSchedule:     e,
		PaymentSchedule:         ap,
		//Questionnaire:           questionnaire,
	}, nil
}

// SearchOfferItems implements OfferItemUsecase.
func (u *offerItemUsecaseImpl) SearchOfferItems(ctx context.Context, nameQuery *string, itemID *model.ItemID, dfItemID *model.DfItemID, limit, offset uint) ([]OfferItem, uint, error) {
	fmt.Println("======SearchOfferItems======")
	res, paging, err := u.offerItemAdapter.SearchOfferItems(ctx, nameQuery, itemID, dfItemID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("u.offerItemAdapter.SearchOfferItems: %w", err)
	}
	os := make([]OfferItem, 0, len(res))
	for _, o := range res {
		i, l, d, pe, p, e, ap, s := convertSchedules(o.GetSchedules())
		countMap, err := u.offerItemAdapter.GetStageAssigneeCountMap(ctx, model.OfferItemID(o.Id))
		if err != nil {
			return nil, 0, fmt.Errorf("u.offerItemAdapter.GetStageAssigneeCountMap: %w", err)
		}
		oi := OfferItem{
			OfferItem:               o,
			InvitationSchedule:      i,
			LotterySchedule:         l,
			ShipmentSchedule:        s,
			DraftSubmissionSchedule: d,
			PreExaminationSchedule:  pe,
			ArticlePostingSchedule:  p,
			ExaminationSchedule:     e,
			PaymentSchedule:         ap,
			StageAssigneeCountMap:   countMap,
		}
		os = append(os, oi)
	}
	return os, uint(paging.TotalCount), nil
}

func (u *offerItemUsecaseImpl) ListOfferItems(ctx context.Context, limit, offset uint) ([]OfferItem, uint, error) {
	fmt.Println("======ListOfferItems======")
	res, paging, err := u.offerItemAdapter.ListOfferItems(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("u.offerItemAdapter.ListOfferItems: %w", err)
	}

	fmt.Println("==========res==========: ", res)
	os := make([]OfferItem, 0, len(res))
	for _, o := range res {
		i, l, d, pe, p, e, ap, s := convertSchedules(o.GetSchedules())
		countMap, err := u.offerItemAdapter.GetStageAssigneeCountMap(ctx, model.OfferItemID(o.Id))
		if err != nil {
			return nil, 0, fmt.Errorf("u.offerItemAdapter.GetStageAssigneeCountMap: %w", err)
		}
		oi := OfferItem{
			OfferItem:               o,
			InvitationSchedule:      i,
			LotterySchedule:         l,
			ShipmentSchedule:        s,
			DraftSubmissionSchedule: d,
			PreExaminationSchedule:  pe,
			ArticlePostingSchedule:  p,
			ExaminationSchedule:     e,
			PaymentSchedule:         ap,
			StageAssigneeCountMap:   countMap,
		}
		os = append(os, oi)
	}
	return os, uint(paging.TotalCount), nil
}

// DeleteOfferItem implements OfferItemUsecase.
func (a *offerItemUsecaseImpl) DeleteOfferItem(ctx context.Context, id model.OfferItemID) error {
	if err := a.offerItemAdapter.DeleteOfferItem(ctx, id); err != nil {
		return fmt.Errorf("a.offerItemAdapter.DeleteOfferItem: %w", err)
	}
	return nil
}

// GetStageAssigneeCountMap implements OfferItemUsecase.
func (a *offerItemUsecaseImpl) GetStageAssigneeCountMap(ctx context.Context, id model.OfferItemID) (map[offer_item.Stage]int, error) {
	res, err := a.offerItemAdapter.GetStageAssigneeCountMap(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("a.offerItemAdapter.GetStageAssigneeCountMap: %w", err)
	}
	return res, nil
}

// Invite implements OfferItemUsecase.
func (u *offerItemUsecaseImpl) Invite(ctx context.Context, id model.OfferItemID) error {
	if err := u.offerItemAdapter.InviteOffer(ctx, id); err != nil {
		return fmt.Errorf("u.offerItemAdapter.InviteOffer: %w", err)
	}
	return nil
}

// SaveLotteryResults implements OfferItemUsecase.
func (u *offerItemUsecaseImpl) SaveLotteryResults(ctx context.Context, id model.OfferItemID, results []dto.LotteryResultForm) error {
	if err := u.offerItemAdapter.SaveLotteryResults(ctx, id, results); err != nil {
		return fmt.Errorf("u.offerItemAdapter.SaveLotteryResults: %w", err)
	}
	return nil
}

// SavePaymentResults implements OfferItemUsecase.
func (u *offerItemUsecaseImpl) SavePaymentResults(ctx context.Context, id model.OfferItemID, amebaIDs []model.AmebaID) error {
	if err := u.offerItemAdapter.SavePaymentResults(ctx, id, amebaIDs); err != nil {
		return fmt.Errorf("u.offerItemAdapter.SavePaymentResults: %w", err)
	}
	return nil
}

// SavePreExaminationResults implements OfferItemUsecase.
func (u *offerItemUsecaseImpl) SavePreExaminationResults(ctx context.Context, id model.OfferItemID, results []dto.ExaminationResultForm) error {
	if err := u.offerItemAdapter.SaveExaminationResults(ctx, id, offer_item.EntryType_ENTRY_TYPE_DRAFT, results); err != nil {
		return fmt.Errorf("u.offerItemAdapter.SaveExaminationResults: %w", err)
	}
	return nil
}

// SaveExaminationResults implements OfferItemUsecase.
func (u *offerItemUsecaseImpl) SaveExaminationResults(ctx context.Context, id model.OfferItemID, results []dto.ExaminationResultForm) error {
	if err := u.offerItemAdapter.SaveExaminationResults(ctx, id, offer_item.EntryType_ENTRY_TYPE_ENTRY, results); err != nil {
		return fmt.Errorf("u.offerItemAdapter.SaveExaminationResults: %w", err)
	}
	return nil
}

// Close implements OfferItemUsecase.
func (u *offerItemUsecaseImpl) Close(ctx context.Context, id model.OfferItemID) error {
	if err := u.offerItemAdapter.CloseOffer(ctx, id); err != nil {
		return fmt.Errorf("u.offerItemAdapter.CloseOffer: %w", err)
	}
	return nil
}

// SendRemindMail implements OfferItemUsecase.
func (u *offerItemUsecaseImpl) SendRemindMail(ctx context.Context, id model.OfferItemID, stage offer_item.Stage) error {
	if err := u.offerItemAdapter.SendRemindMail(ctx, id, stage); err != nil {
		return fmt.Errorf("u.offerItemAdapter.SendRemindMail: %w", err)
	}
	return nil
}

// validateAmebaIDs は入力された Ameba ID が Pick のアフィリエイター登録されているかの検証を行います。
//func (u *offerItemUsecaseImpl) validateAmebaIDs(ctx context.Context, amebaIDs []model.AmebaID) error {
//	affs, err := u.affiliatorAdapter.BulkGetByAmebaIDs(ctx, amebaIDs)
//	if err != nil {
//		return fmt.Errorf("u.affiliatorAdapter.BulkGetByAmebaIDs: %w", err)
//	}
//	if len(amebaIDs) == len(affs) {
//		return nil
//	}
//	invalidIDs := make([]string, 0, len(affs))
//	for _, id := range amebaIDs {
//		if _, ok := affs[id]; !ok {
//			invalidIDs = append(invalidIDs, id.String())
//		}
//	}
//	return fmt.Errorf("%w %s", ErrAffiliatorNotFound, strings.Join(invalidIDs, ", "))
//}

// CreateOfferItem implements OfferItemUsecase.
func (u *offerItemUsecaseImpl) CreateOfferItem(ctx context.Context, form dto.OfferItemForm) (*model.OfferItemID, error) {
	//amebaIDs := dto.AssigneeForms(form.Assignees).AmebaIDs()
	//if err := u.validateAmebaIDs(ctx, amebaIDs); err != nil {
	//	return nil, fmt.Errorf("u.validateAmebaIDs: %w", err)
	//}
	form.ID = ""
	if id, err := u.offerItemAdapter.CreateOfferItem(ctx, form); err != nil {
		return nil, fmt.Errorf("u.offerItemAdapter.CreateOfferItem: %w", err)
	} else {
		return id, nil
	}
}

// UpdateOfferItem implements OfferItemUsecase.
func (u *offerItemUsecaseImpl) UpdateOfferItem(ctx context.Context, id model.OfferItemID, form dto.OfferItemForm) error {
	//amebaIDs := dto.AssigneeForms(form.Assignees).AmebaIDs()
	//if err := u.validateAmebaIDs(ctx, amebaIDs); err != nil {
	//	return fmt.Errorf("u.validateAmebaIDs: %w", err)
	//}
	if err := u.offerItemAdapter.UpdateOfferItem(ctx, id, form); err != nil {
		return fmt.Errorf("u.offerItemAdapter.UpdateOfferItem: %w", err)
	}
	return nil
}

type OfferItem struct {
	OfferItem *offer_item.OfferItem
	// スケジュール
	// 参加募集スケジュール
	InvitationSchedule OfferItemSchedule
	// 選考スケジュール
	LotterySchedule OfferItemSchedule
	// 発送スケジュール
	ShipmentSchedule *OfferItemSchedule
	// 下書き提出スケジュール
	DraftSubmissionSchedule OfferItemSchedule
	// 下書き審査
	PreExaminationSchedule OfferItemSchedule
	// 記事投稿
	ArticlePostingSchedule OfferItemSchedule
	// 審査
	ExaminationSchedule OfferItemSchedule
	// 支払い中
	PaymentSchedule       OfferItemSchedule
	Questionnaire         *offer_item.Questionnaire
	Assignees             []Assignee
	StageAssigneeCountMap map[offer_item.Stage]int
}

type OfferItemSchedule struct {
	ID        string
	StartDate time.Time
	EndDate   time.Time
}

type Assignee struct {
	Assignee  *offer_item.Assignee
	OfferItem *offer_item.OfferItem
	//QuestionAnswers []*offer_item.QuestionAnswer
	Examination *offer_item.Examination
	//ExaminationEntry    *externalapi.PrivateEntry
	//Blogger             *externalapi.Blogger
	//AsID                *model.AsID
}

// func convertAssignee(
//
//	assignee *offer_item.Assignee,
//	offerItem *offer_item.OfferItem,
//	personalInformation *personal_information.PersonalInformation,
//	//questionAnswers []*offer_item.QuestionAnswer,
//	examination *offer_item.Examination,
//	//privateEntry *externalapi.PrivateEntry,
//	//blogger *externalapi.Blogger,
//	asID *model.AsID,
//
//	) Assignee {
//		return Assignee{
//			Assignee:  assignee,
//			OfferItem: offerItem,
//			//PersonalInformation: personalInformation,
//			//QuestionAnswers: questionAnswers,
//			Examination: examination,
//			//ExaminationEntry:    privateEntry,
//			//Blogger:             blogger,
//			//AsID:                asID,
//		}
//	}
func convertAssignee(
	assignee *offer_item.Assignee,
	offerItem *offer_item.OfferItem,
	examination *offer_item.Examination,
) Assignee {
	return Assignee{
		Assignee:    assignee,
		OfferItem:   offerItem,
		Examination: examination,
	}
}

// ListAssignees implements OfferItemUsecase.
func (u *offerItemUsecaseImpl) ListAllAssignees(
	ctx context.Context,
	id model.OfferItemID,
	parameter ListAssigneesParameter,
) (Assignees, error) {
	res := make([]Assignee, 0)
	var offerItem *offer_item.OfferItem
	var err error
	if parameter.WithOfferItem {
		offerItem, err = u.offerItemAdapter.GetOfferItem(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("u.offerItemAdapter.GetOfferItem: %w", err)
		}
	}
	var stages []offer_item.Stage
	if parameter.Stage == nil {
		stages = func() []offer_item.Stage {
			res := make([]offer_item.Stage, 0)
			for _, v := range offer_item.Stage_value {
				if v == 0 {
					continue
				}
				res = append(res, offer_item.Stage(v))
			}
			return res
		}()
	} else {
		stages = []offer_item.Stage{*parameter.Stage}
	}
	for _, stage := range stages {
		as, err := u.offerItemAdapter.ListAssignees(ctx, id, stage)
		if err != nil {
			return nil, fmt.Errorf("u.offerItemAdapter.ListAssignees: %w", err)
		}
		if len(as) == 0 {
			continue
		}
		assigneeAmebaIDs := make([]model.AmebaID, 0, len(as))
		for _, a := range as {
			assigneeAmebaIDs = append(assigneeAmebaIDs, model.AmebaID(a.AmebaId))
		}

		examinationMap := make(map[model.AmebaID]*offer_item.Examination)
		if parameter.EntryType != nil {
			examinationMap, err = u.getExaminationMap(ctx, id, *parameter.EntryType)
			if err != nil {
				return nil, fmt.Errorf("u.getExaminationMap: %w", err)
			}
		}

		for _, a := range as {
			res = append(
				res,
				convertAssignee(
					a,
					offerItem,
					examinationMap[model.AmebaID(a.GetAmebaId())],
				),
			)
		}
	}
	return res, nil
}

type assigneeID struct {
	offerItemID model.OfferItemID
	amebaID     model.AmebaID
}

//func (u *offerItemUsecaseImpl) ListAssigneesUnderExamination(ctx context.Context, limit int) ([]Assignee, error) {
//	assignees, err := u.offerItemAdapter.ListAssigneesUnderExamination(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("u.offerItemAdapter.ListAssigneesUnderExamination: %w", err)
//	}
//
//	offerItemIDs := make([]model.OfferItemID, 0)
//	amebaIDs := make([]model.AmebaID, 0)
//	for _, a := range assignees {
//		offerItemIDs = append(offerItemIDs, model.OfferItemID(a.GetOfferItemId()))
//		amebaIDs = append(amebaIDs, model.AmebaID(a.GetAmebaId()))
//	}
//	// Sort, Compact で重複削除
//	slices.Sort(offerItemIDs)
//	offerItemIDs = slices.Compact(offerItemIDs)
//	slices.Sort(amebaIDs)
//	amebaIDs = slices.Compact(amebaIDs)
//
//	offerItemMap := make(map[model.OfferItemID]*offer_item.OfferItem)
//	for _, offerItemID := range offerItemIDs {
//		o, err := u.offerItemAdapter.GetOfferItem(ctx, offerItemID)
//		if err != nil {
//			return nil, fmt.Errorf("u.offerItemAdapter.GetOfferItem: %w", err)
//		}
//		offerItemMap[offerItemID] = o
//	}
//
//	assigneeIDExaminationMap := make(map[assigneeID]*offer_item.Examination)
//	for _, offerItemID := range offerItemIDs {
//		draftExaminationMap, err := u.offerItemAdapter.BulkGetExaminations(ctx, offerItemID, offer_item.EntryType_ENTRY_TYPE_DRAFT)
//		if err != nil {
//			return nil, fmt.Errorf("u.offerItemAdapter.BulkGetExaminations: %w", err)
//		}
//		for amebaID, draftExamination := range draftExaminationMap {
//			assigneeIDExaminationMap[assigneeID{
//				offerItemID: offerItemID,
//				amebaID:     amebaID,
//			}] = draftExamination
//		}
//		examinationMap, err := u.offerItemAdapter.BulkGetExaminations(ctx, offerItemID, offer_item.EntryType_ENTRY_TYPE_ENTRY)
//		if err != nil {
//			return nil, fmt.Errorf("u.offerItemAdapter.BulkGetExaminations: %w", err)
//		}
//		for amebaID, examination := range examinationMap {
//			// 下書き審査のデータがあっても公開審査のデータがある場合は上書きします。
//			// リクエスト時点で審査すべきデータを抽出するため公開記事を優先（公開審査がある場合は下書き審査は終了している）します。
//			assigneeIDExaminationMap[assigneeID{
//				offerItemID: offerItemID,
//				amebaID:     amebaID,
//			}] = examination
//		}
//	}
//
//	entryIDs := make([]ddto.AmebaIDEntryID, 0)
//	for _, e := range assigneeIDExaminationMap {
//		if e.GetOptionalEntryId() == nil {
//			continue
//		}
//		entryIDs = append(entryIDs, ddto.AmebaIDEntryID{
//			AmebaID: e.GetAmebaId(),
//			EntryID: e.GetEntryId(),
//		})
//	}
//	// Sort, Compact で重複削除
//	slices.SortFunc(entryIDs, func(a, b ddto.AmebaIDEntryID) int {
//		if n := cmp.Compare(a.AmebaID, b.AmebaID); n != 0 {
//			return n
//		}
//		return cmp.Compare(a.EntryID, b.EntryID)
//	})
//	entryIDs = slices.Compact(entryIDs)
//	entryMap, err := u.amebaBlogAdapter.BulkGetEntries(ctx, entryIDs)
//	if err != nil {
//		return nil, fmt.Errorf("u.amebaBlogAdapter.BulkGetEntries: %w", err)
//	}
//
//	bloggerMap, err := u.amebaBlogAdapter.BulkGetBloggerByAmebaIDs(ctx, amebaIDs)
//	if err != nil {
//		return nil, fmt.Errorf("u.amebaBlogAdapter.BulkGetBloggerByAmebaIDs: %w", err)
//	}
//
//	res := make([]Assignee, 0)
//	for _, a := range assignees {
//		e, isFutureEntry := func() (*externalapi.PrivateEntry, bool) {
//			ex, ok := assigneeIDExaminationMap[assigneeID{
//				offerItemID: model.OfferItemID(a.GetOfferItemId()),
//				amebaID:     model.AmebaID(a.GetAmebaId()),
//			}]
//			if !ok {
//				return nil, false
//			}
//			e, ok := entryMap[ex.GetEntryId()]
//			if !ok {
//				return nil, false
//			}
//			// 未来記事判定
//			if e.PublishedTime.After(time.Now()) {
//				return &e, true
//			}
//			return &e, false
//		}()
//		// 未来記事かつ記事審査中ステージであれば除外
//		if isFutureEntry && a.GetStage() == offer_item.Stage_STAGE_EXAMINATION {
//			continue
//		}
//
//		res = append(res, Assignee{
//			Assignee:  a,
//			OfferItem: offerItemMap[model.OfferItemID(a.GetOfferItemId())],
//			Examination: assigneeIDExaminationMap[assigneeID{
//				offerItemID: model.OfferItemID(a.GetOfferItemId()),
//				amebaID:     model.AmebaID(a.GetAmebaId()),
//			}],
//		})
//	}
//	return res, nil
//}

func (u *offerItemUsecaseImpl) ListAssigneesUnderExamination(ctx context.Context, limit int) ([]Assignee, error) {
	assignees, err := u.offerItemAdapter.ListAssigneesUnderExamination(ctx)
	if err != nil {
		return nil, fmt.Errorf("u.offerItemAdapter.ListAssigneesUnderExamination: %w", err)
	}

	offerItemIDs := make([]model.OfferItemID, 0)
	for _, a := range assignees {
		offerItemIDs = append(offerItemIDs, model.OfferItemID(a.GetOfferItemId()))
	}
	// Sort, Compact で重複削除
	slices.Sort(offerItemIDs)
	offerItemIDs = slices.Compact(offerItemIDs)

	offerItemMap := make(map[model.OfferItemID]*offer_item.OfferItem)
	for _, offerItemID := range offerItemIDs {
		o, err := u.offerItemAdapter.GetOfferItem(ctx, offerItemID)
		if err != nil {
			return nil, fmt.Errorf("u.offerItemAdapter.GetOfferItem: %w", err)
		}
		offerItemMap[offerItemID] = o
	}

	assigneeIDExaminationMap := make(map[assigneeID]*offer_item.Examination)
	for _, offerItemID := range offerItemIDs {
		draftExaminationMap, err := u.offerItemAdapter.BulkGetExaminations(ctx, offerItemID, offer_item.EntryType_ENTRY_TYPE_DRAFT)
		if err != nil {
			return nil, fmt.Errorf("u.offerItemAdapter.BulkGetExaminations: %w", err)
		}
		for amebaID, draftExamination := range draftExaminationMap {
			assigneeIDExaminationMap[assigneeID{
				offerItemID: offerItemID,
				amebaID:     amebaID,
			}] = draftExamination
		}
		examinationMap, err := u.offerItemAdapter.BulkGetExaminations(ctx, offerItemID, offer_item.EntryType_ENTRY_TYPE_ENTRY)
		if err != nil {
			return nil, fmt.Errorf("u.offerItemAdapter.BulkGetExaminations: %w", err)
		}
		for amebaID, examination := range examinationMap {
			assigneeIDExaminationMap[assigneeID{
				offerItemID: offerItemID,
				amebaID:     amebaID,
			}] = examination
		}
	}

	res := make([]Assignee, 0)
	for _, a := range assignees {
		res = append(res, Assignee{
			Assignee:  a,
			OfferItem: offerItemMap[model.OfferItemID(a.GetOfferItemId())],
			Examination: assigneeIDExaminationMap[assigneeID{
				offerItemID: model.OfferItemID(a.GetOfferItemId()),
				amebaID:     model.AmebaID(a.GetAmebaId()),
			}],
		})
	}
	return res, nil
}

func (u *offerItemUsecaseImpl) getExaminationMap(ctx context.Context, id model.OfferItemID, entryType offer_item.EntryType) (map[model.AmebaID]*offer_item.Examination, error) {
	res, err := u.offerItemAdapter.BulkGetExaminations(ctx, id, entryType)
	if err != nil {
		return nil, fmt.Errorf("u.offerItemAdapter.BulkGetExaminations: %w", err)
	}
	return res, nil
}

func (u *offerItemUsecaseImpl) getQuestionAnswersMap(ctx context.Context, id model.OfferItemID, assigneeAmebaIDs []model.AmebaID) (map[model.AmebaID][]*offer_item.QuestionAnswer, error) {
	res, err := u.offerItemAdapter.BulkGetQuestionnaireQuestionAnswers(ctx, id, assigneeAmebaIDs)
	if err != nil {
		return nil, fmt.Errorf("u.offerItemAdapter.BulkGetQuestionnaireQuestionAnswers: %w", err)
	}
	m := make(map[model.AmebaID][]*offer_item.QuestionAnswer)
	for k, v := range res {
		answers := make([]*offer_item.QuestionAnswer, 0, len(v.GetQuestionIdAnswerMap()))
		for _, v2 := range v.GetQuestionIdAnswerMap() {
			answers = append(answers, v2)
		}
		m[k] = answers
	}
	return m, nil
}

func convertOfferItemForm(o OfferItem) dto.OfferItemForm {
	return dto.OfferItemForm{
		ID:     o.OfferItem.Id,
		Name:   o.OfferItem.GetName(),
		ItemID: model.ItemID(o.OfferItem.GetItem().GetId()),
		DfItemID: func() *model.DfItemID {
			if o.OfferItem.GetOptionalDfItem() == nil {
				return nil
			}
			id := o.OfferItem.GetDfItem().GetId()
			return (*model.DfItemID)(&id)
		}(),
		CouponBannerID: func() *model.BannerID {
			if o.OfferItem.GetOptionalCouponBannerId() == nil {
				return nil
			}
			id := o.OfferItem.GetCouponBannerId()
			return (*model.BannerID)(&id)
		}(),
		SpecialAmount: o.OfferItem.GetSpecialAmount(),
		SpecialRate:   o.OfferItem.GetSpecialRate(),
		HasSpecialCommission: func() bool {
			if o.OfferItem.GetSpecialAmount() > 0 {
				return true
			}
			if o.OfferItem.GetSpecialRate() > 0 {
				return true
			}
			return false
		}(),
		HasSample:                         o.OfferItem.GetHasSample(),
		NeedsPreliminaryReview:            o.OfferItem.GetNeedsPreliminaryReview(),
		NeedsAfterReview:                  o.OfferItem.GetNeedsAfterReview(),
		NeedsPRMark:                       o.OfferItem.GetNeedsPrMark(),
		PostRequired:                      o.OfferItem.GetPostRequired(),
		PostTarget:                        o.OfferItem.GetPostTarget(),
		HasCoupon:                         o.OfferItem.GetHasCoupon(),
		HasLottery:                        o.OfferItem.GetHasLottery(),
		ProductFeatures:                   o.OfferItem.GetProductFeatures(),
		CautionaryPoints:                  o.OfferItem.GetCautionaryPoints(),
		ReferenceInfo:                     o.OfferItem.GetReferenceInfo(),
		OtherInfo:                         o.OfferItem.GetOtherInfo(),
		IsInvitationMailSent:              o.OfferItem.GetIsInvitationMailSent(),
		IsOfferDetailMailSent:             o.OfferItem.GetIsOfferDetailMailSent(),
		IsPassedPreliminaryReviewMailSent: o.OfferItem.GetIsPassedPreliminaryReviewMailSent(),
		IsFailedPreliminaryReviewMailSent: o.OfferItem.GetIsFailedPreliminaryReviewMailSent(),
		IsArticlePostMailSent:             o.OfferItem.GetIsArticlePostMailSent(),
		IsPassedAfterReviewMailSent:       o.OfferItem.GetIsPassedAfterReviewMailSent(),
		IsFailedAfterReviewMailSent:       o.OfferItem.GetIsFailedAfterReviewMailSent(),
		InvitationSchedule:                dto.NewScheduleForm(o.InvitationSchedule.ID, o.InvitationSchedule.StartDate, o.InvitationSchedule.EndDate),
		LotterySchedule:                   dto.NewScheduleForm(o.LotterySchedule.ID, o.LotterySchedule.StartDate, o.LotterySchedule.EndDate),
		ShipmentSchedule:                  dto.NewScheduleForm(o.ShipmentSchedule.ID, o.ShipmentSchedule.StartDate, o.ShipmentSchedule.EndDate),
		DraftSubmissionSchedule:           dto.NewScheduleForm(o.DraftSubmissionSchedule.ID, o.DraftSubmissionSchedule.StartDate, o.DraftSubmissionSchedule.EndDate),
		PreExaminationSchedule:            dto.NewScheduleForm(o.PreExaminationSchedule.ID, o.PreExaminationSchedule.StartDate, o.PreExaminationSchedule.EndDate),
		ArticlePostingSchedule:            dto.NewScheduleForm(o.ArticlePostingSchedule.ID, o.ArticlePostingSchedule.StartDate, o.ArticlePostingSchedule.EndDate),
		ExaminationSchedule:               dto.NewScheduleForm(o.ExaminationSchedule.ID, o.ExaminationSchedule.StartDate, o.ExaminationSchedule.EndDate),
		PaymentSchedule:                   dto.NewScheduleForm(o.PaymentSchedule.ID, o.PaymentSchedule.EndDate, o.PaymentSchedule.EndDate), // StartDate はゼロ値のため EndDate を設定
		HasQuestionnaire:                  o.Questionnaire != nil,
		Questionnaire: func() dto.QuestionnaireForm {
			if o.Questionnaire == nil {
				return dto.QuestionnaireForm{
					Detail:    "",
					Questions: []dto.QuestionForm{},
				}
			}
			questionnaire := o.Questionnaire
			return dto.QuestionnaireForm{
				Detail: questionnaire.GetDescription(),
				Questions: func() []dto.QuestionForm {
					res := make([]dto.QuestionForm, 0, len(questionnaire.GetQuestions()))
					for _, q := range questionnaire.GetQuestions() {
						res = append(res, dto.QuestionForm{
							ID:        q.GetId(),
							Type:      int(q.GetQuestionType()),
							Title:     q.GetTitle(),
							ImageData: q.GetImageUrl(),
							Options:   q.GetOptions(),
						})
					}
					return res
				}(),
			}
		}(),
		Assignees: func() []dto.AssigneeForm {
			res := make([]dto.AssigneeForm, 0, len(o.Assignees))
			for _, a := range o.Assignees {
				res = append(res, dto.AssigneeForm{
					AmebaID:    a.Assignee.GetAmebaId(),
					WritingFee: int(a.Assignee.GetWritingFee()),
				})
			}
			return res
		}(),
		DraftedItemInfoName:        o.OfferItem.GetDraftedItemInfo().GetName(),
		DraftedItemInfoContentName: o.OfferItem.GetDraftedItemInfo().GetContentName(),
		DraftedItemInfoURL:         o.OfferItem.GetDraftedItemInfo().GetUrl(),
		DraftedItemInfoImageURL:    o.OfferItem.GetDraftedItemInfo().GetImageUrl(),
		DraftedItemInfoCommissionType: func() offer_item.ItemCommissionType {
			return o.OfferItem.GetDraftedItemInfo().GetMinCommission().GetCommissionType()
		}(),
		DraftedItemInfoMinCommission: float64(o.OfferItem.GetDraftedItemInfo().GetMinCommission().GetCalculatedRate()),
		DraftedItemInfoMaxCommission: float64(o.OfferItem.GetDraftedItemInfo().GetMaxCommission().GetCalculatedRate()),
	}
}

func convertSchedules(schedules []*offer_item.Schedule) (invitation, lottery, draftSubmission, preExamination, articlePosting, examination, payment OfferItemSchedule, shipment *OfferItemSchedule) {
	for i := range schedules {
		schedule := schedules[i]
		switch schedule.GetScheduleType() {
		case offer_item.ScheduleType_SCHEDULE_TYPE_INVITATION:
			invitation = OfferItemSchedule{
				ID:        schedule.GetId(),
				StartDate: schedule.GetStartDate().AsTime(),
				EndDate:   schedule.GetEndDate().AsTime(),
			}
		case offer_item.ScheduleType_SCHEDULE_TYPE_LOTTERY:
			lottery = OfferItemSchedule{
				ID:        schedule.GetId(),
				StartDate: schedule.GetStartDate().AsTime(),
				EndDate:   schedule.GetEndDate().AsTime(),
			}
		case offer_item.ScheduleType_SCHEDULE_TYPE_SHIPMENT:
			shipment = &OfferItemSchedule{
				ID:        schedule.GetId(),
				StartDate: schedule.GetStartDate().AsTime(),
				EndDate:   schedule.GetEndDate().AsTime(),
			}
		case offer_item.ScheduleType_SCHEDULE_TYPE_DRAFT_SUBMISSION:
			draftSubmission = OfferItemSchedule{
				ID:        schedule.GetId(),
				StartDate: schedule.GetStartDate().AsTime(),
				EndDate:   schedule.GetEndDate().AsTime(),
			}
		case offer_item.ScheduleType_SCHEDULE_TYPE_PRE_EXAMINATION:
			preExamination = OfferItemSchedule{
				ID:        schedule.GetId(),
				StartDate: schedule.GetStartDate().AsTime(),
				EndDate:   schedule.GetEndDate().AsTime(),
			}
		case offer_item.ScheduleType_SCHEDULE_TYPE_ARTICLE_POSTING:
			articlePosting = OfferItemSchedule{
				ID:        schedule.GetId(),
				StartDate: schedule.GetStartDate().AsTime(),
				EndDate:   schedule.GetEndDate().AsTime(),
			}
		case offer_item.ScheduleType_SCHEDULE_TYPE_EXAMINATION:
			examination = OfferItemSchedule{
				ID:        schedule.GetId(),
				StartDate: schedule.GetStartDate().AsTime(),
				EndDate:   schedule.GetEndDate().AsTime(),
			}
		case offer_item.ScheduleType_SCHEDULE_TYPE_PAYMENT:
			payment = OfferItemSchedule{
				ID:        schedule.GetId(),
				StartDate: schedule.GetStartDate().AsTime(),
				EndDate:   schedule.GetEndDate().AsTime(),
			}
		default:
		}
	}
	return //nolint: nakedret
}

func (u *offerItemUsecaseImpl) GetOfferItemForm(ctx context.Context, id model.OfferItemID) (*dto.OfferItemForm, error) {
	o, err := u.GetOfferItem(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("u.GetOfferItem: %w", err)
	}
	f := convertOfferItemForm(*o)
	return &f, nil
}

func (u *offerItemUsecaseImpl) GetQuestionnaire(ctx context.Context, offerItemID model.OfferItemID) (*offer_item.Questionnaire, error) {
	res, err := u.offerItemAdapter.GetQuestionnaire(ctx, offerItemID)
	if err != nil {
		if errors.Is(err, adapter.ErrAdapterNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("u.offerItemAdapter.GetQuestionnaire: %w", err)
	}
	return res, nil
}

func (u *offerItemUsecaseImpl) FinishShipment(ctx context.Context, offerItemID model.OfferItemID) error {
	if err := u.offerItemAdapter.FinishShipment(ctx, offerItemID); err != nil {
		return fmt.Errorf("u.offerItemAdapter.FinishShipment: %w", err)
	}
	return nil
}
