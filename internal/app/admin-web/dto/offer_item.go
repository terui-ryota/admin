package dto

import (
	"slices"
	"time"

	"github.com/terui-ryota/admin/internal/domain/model"
	"github.com/terui-ryota/protofiles/go/offer_item"
)

type OfferItemForm struct {
	ID string `json:"id"`
	// 基本設定
	// PR投稿案件名
	Name string `json:"name"`
	// 案件 ID
	ItemID model.ItemID `json:"item_id"`
	// DF 案件 ID
	DfItemID *model.DfItemID `json:"df_item_id"`

	DraftedItemInfoName           string                        `json:"drafted_item_info_name"`
	DraftedItemInfoContentName    string                        `json:"drafted_item_info_content_name"`
	DraftedItemInfoURL            string                        `json:"drafted_item_info_url"`
	DraftedItemInfoImageURL       string                        `json:"drafted_item_info_image_url"`
	DraftedItemInfoCommissionType offer_item.ItemCommissionType `json:"drafted_item_info_commission_type"`
	DraftedItemInfoMinCommission  float64                       `json:"drafted_item_info_min_commission"`
	DraftedItemInfoMaxCommission  float64                       `json:"drafted_item_info_max_commission"`

	CouponBannerID         *model.BannerID       `json:"coupon_banner_id"`
	HasSpecialCommission   bool                  `json:"has_special_commission"`
	SpecialAmount          int64                 `json:"special_amount"`
	SpecialRate            float64               `json:"special_rate"`
	HasSample              bool                  `json:"has_sample"`
	NeedsPreliminaryReview bool                  `json:"needs_preliminary_review"`
	NeedsAfterReview       bool                  `json:"needs_after_review"`
	NeedsPRMark            bool                  `json:"needs_pr_mark"`
	PostRequired           bool                  `json:"post_required"`
	PostTarget             offer_item.PostTarget `json:"post_target"`
	HasCoupon              bool                  `json:"has_coupon"`
	HasLottery             bool                  `json:"has_lottery"`
	ProductFeatures        string                `json:"product_features"`
	CautionaryPoints       string                `json:"cautionary_points"`
	ReferenceInfo          string                `json:"reference_info"`
	OtherInfo              string                `json:"other_info"`

	IsInvitationMailSent              bool `json:"is_invitation_mail_sent"`
	IsOfferDetailMailSent             bool `json:"is_offer_detail_mail_sent"`
	IsPassedPreliminaryReviewMailSent bool `json:"is_passed_preliminary_review_mail_sent"`
	IsFailedPreliminaryReviewMailSent bool `json:"is_failed_preliminary_review_mail_sent"`
	IsArticlePostMailSent             bool `json:"is_article_post_mail_sent"`
	IsPassedAfterReviewMailSent       bool `json:"is_passed_after_review_mail_sent"`
	IsFailedAfterReviewMailSent       bool `json:"is_failed_after_review_mail_sent"`

	// スケジュール
	// 参加募集
	InvitationSchedule ScheduleForm `json:"invitation_schedule"`
	// 選考
	LotterySchedule ScheduleForm `json:"lottery_schedule"`
	// 発送
	ShipmentSchedule ScheduleForm `json:"shipment_schedule"`
	// 下書き提出
	DraftSubmissionSchedule ScheduleForm `json:"draft_submission_schedule"`
	// 下書き審査
	PreExaminationSchedule ScheduleForm `json:"pre_examination_schedule"`
	// 記事投稿
	ArticlePostingSchedule ScheduleForm `json:"article_posting_schedule"`
	// 審査
	ExaminationSchedule ScheduleForm `json:"examination_schedule"`
	// 支払い中（記事投稿後）
	PaymentSchedule ScheduleForm `json:"payment_schedule"`

	// アンケート有無
	HasQuestionnaire bool `json:"has_questionnaire"`
	// アンケート
	Questionnaire QuestionnaireForm `json:"questionnaire"`

	// アサイニー
	Assignees []AssigneeForm `json:"assignees"`
}

type AssigneeForms []AssigneeForm

func (fs AssigneeForms) AmebaIDs() []model.AmebaID {
	res := make([]model.AmebaID, 0)
	for _, f := range fs {
		res = append(res, model.AmebaID(f.AmebaID))
	}
	slices.Sort(res)
	return slices.Compact(res)
}

type AssigneeForm struct {
	AmebaID    string `json:"ameba_id"`
	StageName  string `json:"stage"`
	WritingFee int    `json:"writing_fee"`
}

func (f *AssigneeForm) Stage() offer_item.Stage {
	s, ok := offer_item.Stage_value[f.StageName]
	if !ok {
		return 0
	}
	return offer_item.Stage(s)
}

type ScheduleForm struct {
	ID        string    `json:"id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

func NewScheduleForm(id string, s, e time.Time) ScheduleForm {
	return ScheduleForm{
		ID:        id,
		StartDate: s.Local(),
		EndDate:   e.Local(),
	}
}

type QuestionnaireForm struct {
	// 詳細
	Detail    string         `json:"detail"`
	Questions []QuestionForm `json:"questions"`
}

type QuestionForm struct {
	ID        string   `json:"id"`
	Type      int      `json:"type"` // offer_item.Questionnaire_QuestionType
	Title     string   `json:"title"`
	ImageData string   `json:"image_data"`
	Options   []string `json:"options"`
}

type LotteryResultForm struct {
	AmebaID      string   `json:"ameba_id"`
	IsPassed     bool     `json:"is_passed"`
	ShippingData []string `json:"shipping_data"`
	JANCode      string   `json:"jan_code"`
}

type ExaminationResultForm struct {
	AmebaID      string  `json:"ameba_id"`
	IsPassed     bool    `json:"is_passed"`
	ExaminerName string  `json:"examiner_name"`
	Reason       *string `json:"reason"`
}
