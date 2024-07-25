package adapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/terui-ryota/admin/internal/app/admin-web/config"
	"github.com/terui-ryota/admin/internal/app/admin-web/dto"
	"github.com/terui-ryota/admin/internal/domain/adapter"
	"github.com/terui-ryota/admin/internal/domain/model"
	"github.com/terui-ryota/protofiles/go/common"
	"github.com/terui-ryota/protofiles/go/offer_item"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type offerItemAdapterImpl struct {
	offerItemHandlerClient offer_item.OfferItemHandlerClient
}

func NewOfferItemAdapterImpl(config config.OfferItemGRPCServer) (adapter.OfferItemAdapter, error) {
	fmt.Println("======config.Host======: ", config.Host)
	fmt.Println("======config.Port======: ", config.Port)
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", config.Host, config.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//grpc.WithUnaryInterceptor(
		//	grpc_middleware.UnaryClientInterceptor("admin"),
		//),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{
			StartOptions: trace.StartOptions{
				Sampler: trace.ProbabilitySampler(0.5),
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc.Dial: %w", err)
	}
	return &offerItemAdapterImpl{
		offerItemHandlerClient: offer_item.NewOfferItemHandlerClient(conn),
	}, nil
}

func (a *offerItemAdapterImpl) GetQuestionnaire(ctx context.Context, offerItemID model.OfferItemID) (*offer_item.Questionnaire, error) {
	res, err := a.offerItemHandlerClient.GetQuestionnaire(ctx, &offer_item.GetQuestionnaireRequest{
		OfferItemId: offerItemID.String(),
	})
	if err != nil {
		//if errors.Is(err, apperr.OfferItemNotFoundError) {
		if errors.Is(err, errors.New("OfferItemNotFoundError")) {
			return nil, adapter.ErrAdapterNotFound
		}
		return nil, fmt.Errorf("a.offerItemHandlerClient.GetQuestionnaire: %w", err)
	}
	return res.GetQuestionnaire(), nil
}

func (a *offerItemAdapterImpl) GetOfferItem(ctx context.Context, offerItemID model.OfferItemID) (*offer_item.OfferItem, error) {
	res, err := a.offerItemHandlerClient.GetOfferItem(ctx, &offer_item.GetOfferItemRequest{
		OfferItemId: offerItemID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("a.offerItemHandlerClient.GetOfferItem: %w", err)
	}
	return res.GetOfferItem(), nil
}

func (a *offerItemAdapterImpl) SearchOfferItems(ctx context.Context, nameQuery *string, itemID *model.ItemID, dfItemID *model.DfItemID, limit, offset uint) ([]*offer_item.OfferItem, *common.ListResult, error) {
	res, err := a.offerItemHandlerClient.SearchOfferItem(ctx, &offer_item.SearchOfferItemRequest{
		OptionalOfferItemName: func() *offer_item.SearchOfferItemRequest_OfferItemName {
			if nameQuery == nil {
				return nil
			}
			return &offer_item.SearchOfferItemRequest_OfferItemName{
				OfferItemName: *nameQuery,
			}
		}(),
		OptionalItemId: func() *offer_item.SearchOfferItemRequest_ItemId {
			if itemID == nil {
				return nil
			}
			return &offer_item.SearchOfferItemRequest_ItemId{
				ItemId: itemID.String(),
			}
		}(),
		OptionalDfItemId: func() *offer_item.SearchOfferItemRequest_DfItemId {
			if dfItemID == nil {
				return nil
			}
			return &offer_item.SearchOfferItemRequest_DfItemId{
				DfItemId: dfItemID.String(),
			}
		}(),
		Condition: &common.ListCondition{
			Offset: uint32(offset),
			Limit:  uint32(limit),
			Sort: []*common.Sort{
				{
					OrderBy:  "created_at",
					Ordering: common.Ordering_DESC,
				},
			},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("a.offerItemHandlerClient.SearchOfferItem: %w", err)
	}
	if res.GetOfferItems() == nil {
		return make([]*offer_item.OfferItem, 0), res.GetResult(), nil
	}
	return res.GetOfferItems(), res.GetResult(), nil
}

func (a *offerItemAdapterImpl) ListOfferItems(ctx context.Context, limit, offset uint) ([]*offer_item.OfferItem, *common.ListResult, error) {
	res, err := a.offerItemHandlerClient.ListOfferItem(ctx, &offer_item.ListOfferItemRequest{
		Condition: &common.ListCondition{
			Offset: uint32(offset),
			Limit:  uint32(limit),
			Sort: []*common.Sort{
				{
					OrderBy:  "created_at",
					Ordering: common.Ordering_DESC,
				},
			},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("a.offerItemHandlerClient.ListOfferItem: %w", err)
	}
	return res.GetOfferItems(), res.GetResult(), nil
}

// BulkGetStageAssigneeCount implements adapter.OfferItemAdapter.
func (a *offerItemAdapterImpl) GetStageAssigneeCountMap(ctx context.Context, offerItemID model.OfferItemID) (map[offer_item.Stage]int, error) {
	res, err := a.offerItemHandlerClient.ListStageAssigneeCount(ctx, &offer_item.ListStageAssigneeCountRequest{
		OfferItemId: offerItemID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("a.offerItemHandlerClient.ListStageAssigneeCount: %w", err)
	}
	m := make(map[offer_item.Stage]int)
	for _, c := range res.GetAssigneeCounts() {
		m[c.GetStage()] = int(c.GetCount())
	}
	return m, nil
}

// DeleteOfferItem implements adapter.OfferItemV2Adapter.
func (a *offerItemAdapterImpl) DeleteOfferItem(ctx context.Context, offerItemID model.OfferItemID) error {
	if _, err := a.offerItemHandlerClient.DeleteOfferItem(ctx, &offer_item.DeleteOfferItemRequest{
		OfferItemId: offerItemID.String(),
	}); err != nil {
		return fmt.Errorf("a.offerItemHandlerClient.DeleteOfferItem: %w", err)
	}
	return nil
}

func (a *offerItemAdapterImpl) CloseOffer(ctx context.Context, offerItemID model.OfferItemID) error {
	if _, err := a.offerItemHandlerClient.CompletedOfferItem(ctx, &offer_item.CompletedOfferItemRequest{
		OfferItemId: offerItemID.String(),
	}); err != nil {
		return fmt.Errorf("a.offerItemHandlerClient.CompletedOfferItem: %w", err)
	}
	return nil
}

func (a *offerItemAdapterImpl) SaveLotteryResults(ctx context.Context, offerItemID model.OfferItemID, results []dto.LotteryResultForm) error {
	if _, err := a.offerItemHandlerClient.UploadLotteryResults(ctx, &offer_item.UploadLotteryResultsRequest{
		OfferItemId: offerItemID.String(),
		MapLotteryResultWithShippingData: func() map[string]*offer_item.UploadLotteryResultsRequest_LotteryResult {
			res := make(map[string]*offer_item.UploadLotteryResultsRequest_LotteryResult)
			for _, f := range results {
				var janCode *offer_item.UploadLotteryResultsRequest_LotteryResult_JanCode
				if f.JANCode != "" { // 空なら詰めない
					janCode = &offer_item.UploadLotteryResultsRequest_LotteryResult_JanCode{
						JanCode: f.JANCode,
					}
				}
				res[f.AmebaID] = &offer_item.UploadLotteryResultsRequest_LotteryResult{
					IsPassedLottery: f.IsPassed,
					ShippingData:    f.ShippingData,
					OptionalJanCode: janCode,
				}
			}
			return res
		}(),
	}); err != nil {
		return fmt.Errorf("a.offerItemHandlerClient.UploadLotteryResults: %w", err)
	}
	return nil
}

// InviteOffer implements adapter.OfferItemV2Adapter.
func (a *offerItemAdapterImpl) InviteOffer(ctx context.Context, offerItemID model.OfferItemID) error {
	if _, err := a.offerItemHandlerClient.InviteOffer(ctx, &offer_item.InviteOfferRequest{
		OfferItemId: offerItemID.String(),
	}); err != nil {
		return fmt.Errorf("a.offerItemHandlerClient.InviteOffer: %w", err)
	}
	return nil
}

// SendRemindMail implements adapter.OfferItemAdapter.
func (a *offerItemAdapterImpl) SendRemindMail(ctx context.Context, offerItemID model.OfferItemID, stage offer_item.Stage) error {
	if _, err := a.offerItemHandlerClient.SendMail(ctx, &offer_item.SendMailRequest{
		OfferItemId: offerItemID.String(),
		Stage:       stage,
	}); err != nil {
		return fmt.Errorf("a.offerItemHandlerClient.SendMail: %w", err)
	}
	return nil
}

// CreateOfferItem implements adapter.OfferItemAdapter.
func (a *offerItemAdapterImpl) CreateOfferItem(ctx context.Context, form dto.OfferItemForm) (*model.OfferItemID, error) {
	if res, err := a.offerItemHandlerClient.SaveOfferItem(ctx, &offer_item.SaveOfferItemRequest{
		OfferItem: &offer_item.SaveOfferItem{
			Id:     form.ID,
			Name:   form.Name,
			ItemId: form.ItemID.String(),
			OptionalDfItemId: func() *offer_item.SaveOfferItem_DfItemId {
				if form.DfItemID == nil {
					return nil
				}
				return &offer_item.SaveOfferItem_DfItemId{
					DfItemId: form.DfItemID.String(),
				}
			}(),
			OptionalCouponBannerId: func() *offer_item.SaveOfferItem_CouponBannerId {
				if form.CouponBannerID == nil || *form.CouponBannerID == "" {
					return nil
				}
				return &offer_item.SaveOfferItem_CouponBannerId{
					CouponBannerId: form.CouponBannerID.String(),
				}
			}(),
			DraftedItemInfo: &offer_item.ItemInfo{
				Name:        form.DraftedItemInfoName,
				ContentName: form.DraftedItemInfoContentName,
				ImageUrl:    form.DraftedItemInfoImageURL,
				Url:         form.DraftedItemInfoURL,
				MinCommission: &offer_item.Commission{
					CommissionType: form.DraftedItemInfoCommissionType,
					CalculatedRate: float32(form.DraftedItemInfoMinCommission),
				},
				MaxCommission: &offer_item.Commission{
					CommissionType: form.DraftedItemInfoCommissionType,
					CalculatedRate: float32(form.DraftedItemInfoMaxCommission),
				},
			},
			HasSample:                         form.HasSample,
			NeedsPreliminaryReview:            form.NeedsPreliminaryReview,
			NeedsAfterReview:                  form.NeedsAfterReview,
			NeedsPrMark:                       form.NeedsPRMark,
			PostRequired:                      form.PostRequired,
			PostTarget:                        form.PostTarget,
			HasCoupon:                         form.HasCoupon,
			SpecialAmount:                     form.SpecialAmount,
			SpecialRate:                       form.SpecialRate,
			HasSpecialCommission:              form.HasSpecialCommission,
			HasLottery:                        form.HasLottery,
			ProductFeatures:                   form.ProductFeatures,
			CautionaryPoints:                  form.CautionaryPoints,
			ReferenceInfo:                     form.ReferenceInfo,
			OtherInfo:                         form.OtherInfo,
			IsInvitationMailSent:              form.IsInvitationMailSent,
			IsOfferDetailMailSent:             form.IsOfferDetailMailSent,
			IsPassedPreliminaryReviewMailSent: form.IsPassedPreliminaryReviewMailSent,
			IsFailedPreliminaryReviewMailSent: form.IsFailedPreliminaryReviewMailSent,
			IsArticlePostMailSent:             form.IsArticlePostMailSent,
			IsPassedAfterReviewMailSent:       form.IsPassedAfterReviewMailSent,
			IsFailedAfterReviewMailSent:       form.IsFailedAfterReviewMailSent,
			Schedules: func() []*offer_item.SaveSchedule {
				res := make([]*offer_item.SaveSchedule, 0)
				if form.HasSample {
					res = append(res, &offer_item.SaveSchedule{
						Id:           form.ShipmentSchedule.ID,
						ScheduleType: offer_item.ScheduleType_SCHEDULE_TYPE_SHIPMENT,
						OptionalStartDate: &offer_item.SaveSchedule_StartDate{
							StartDate: timestamppb.New(form.ShipmentSchedule.StartDate),
						},
						OptionalEndDate: &offer_item.SaveSchedule_EndDate{
							EndDate: timestamppb.New(form.ShipmentSchedule.EndDate),
						},
					})
				} else {
					res = append(res, &offer_item.SaveSchedule{
						Id:           form.ShipmentSchedule.ID,
						ScheduleType: offer_item.ScheduleType_SCHEDULE_TYPE_SHIPMENT,
					})
				}
				if form.HasLottery {
					res = append(res, &offer_item.SaveSchedule{
						Id:           form.LotterySchedule.ID,
						ScheduleType: offer_item.ScheduleType_SCHEDULE_TYPE_LOTTERY,
						OptionalStartDate: &offer_item.SaveSchedule_StartDate{
							StartDate: timestamppb.New(form.LotterySchedule.StartDate),
						},
						OptionalEndDate: &offer_item.SaveSchedule_EndDate{
							EndDate: timestamppb.New(form.LotterySchedule.EndDate),
						},
					})
				} else {
					res = append(res, &offer_item.SaveSchedule{
						Id:           form.LotterySchedule.ID,
						ScheduleType: offer_item.ScheduleType_SCHEDULE_TYPE_LOTTERY,
					})
				}
				if form.NeedsPreliminaryReview {
					res = append(res,
						&offer_item.SaveSchedule{
							Id:           form.DraftSubmissionSchedule.ID,
							ScheduleType: offer_item.ScheduleType_SCHEDULE_TYPE_DRAFT_SUBMISSION,
							OptionalStartDate: &offer_item.SaveSchedule_StartDate{
								StartDate: timestamppb.New(form.DraftSubmissionSchedule.StartDate),
							},
							OptionalEndDate: &offer_item.SaveSchedule_EndDate{
								EndDate: timestamppb.New(form.DraftSubmissionSchedule.EndDate),
							},
						},
						&offer_item.SaveSchedule{
							Id:           form.PreExaminationSchedule.ID,
							ScheduleType: offer_item.ScheduleType_SCHEDULE_TYPE_PRE_EXAMINATION,
							OptionalStartDate: &offer_item.SaveSchedule_StartDate{
								StartDate: timestamppb.New(form.PreExaminationSchedule.StartDate),
							},
							OptionalEndDate: &offer_item.SaveSchedule_EndDate{
								EndDate: timestamppb.New(form.PreExaminationSchedule.EndDate),
							},
						},
					)
				} else {
					res = append(res,
						&offer_item.SaveSchedule{
							Id:           form.DraftSubmissionSchedule.ID,
							ScheduleType: offer_item.ScheduleType_SCHEDULE_TYPE_DRAFT_SUBMISSION,
						},
						&offer_item.SaveSchedule{
							Id:           form.PreExaminationSchedule.ID,
							ScheduleType: offer_item.ScheduleType_SCHEDULE_TYPE_PRE_EXAMINATION,
						},
					)
				}
				if form.NeedsAfterReview {
					res = append(res,
						&offer_item.SaveSchedule{
							Id:           form.ExaminationSchedule.ID,
							ScheduleType: offer_item.ScheduleType_SCHEDULE_TYPE_EXAMINATION,
							OptionalStartDate: &offer_item.SaveSchedule_StartDate{
								StartDate: timestamppb.New(form.ExaminationSchedule.StartDate),
							},
							OptionalEndDate: &offer_item.SaveSchedule_EndDate{
								EndDate: timestamppb.New(form.ExaminationSchedule.EndDate),
							},
						},
					)
				} else {
					res = append(res,
						&offer_item.SaveSchedule{
							Id:           form.ExaminationSchedule.ID,
							ScheduleType: offer_item.ScheduleType_SCHEDULE_TYPE_EXAMINATION,
						},
					)
				}
				res = append(
					res,
					&offer_item.SaveSchedule{
						Id:           form.InvitationSchedule.ID,
						ScheduleType: offer_item.ScheduleType_SCHEDULE_TYPE_INVITATION,
						OptionalStartDate: &offer_item.SaveSchedule_StartDate{
							StartDate: timestamppb.New(form.InvitationSchedule.StartDate),
						},
						OptionalEndDate: &offer_item.SaveSchedule_EndDate{
							EndDate: timestamppb.New(form.InvitationSchedule.EndDate),
						},
					},
					&offer_item.SaveSchedule{
						Id:           form.ArticlePostingSchedule.ID,
						ScheduleType: offer_item.ScheduleType_SCHEDULE_TYPE_ARTICLE_POSTING,
						OptionalStartDate: &offer_item.SaveSchedule_StartDate{
							StartDate: timestamppb.New(form.ArticlePostingSchedule.StartDate),
						},
						OptionalEndDate: &offer_item.SaveSchedule_EndDate{
							EndDate: timestamppb.New(form.ArticlePostingSchedule.EndDate),
						},
					},
					&offer_item.SaveSchedule{
						Id:                form.PaymentSchedule.ID,
						ScheduleType:      offer_item.ScheduleType_SCHEDULE_TYPE_PAYMENT,
						OptionalStartDate: nil,
						OptionalEndDate: &offer_item.SaveSchedule_EndDate{
							EndDate: timestamppb.New(form.PaymentSchedule.EndDate),
						},
					},
				)
				return res
			}(),
			OptionalQuestionnaire: func() *offer_item.SaveOfferItem_Questionnaire {
				if !form.HasQuestionnaire {
					return nil
				}
				questionnaire := form.Questionnaire
				return &offer_item.SaveOfferItem_Questionnaire{
					Questionnaire: &offer_item.Questionnaire{
						Description: questionnaire.Detail,
						Questions: func() []*offer_item.Questionnaire_Question {
							res := make([]*offer_item.Questionnaire_Question, 0, len(questionnaire.Questions))
							for _, q := range questionnaire.Questions {
								res = append(res, &offer_item.Questionnaire_Question{
									Id:           q.ID,
									QuestionType: offer_item.Questionnaire_QuestionType(q.Type),
									Title:        q.Title,
									ImageUrl:     q.ImageData,
									Options: func() []string {
										switch offer_item.Questionnaire_QuestionType(q.Type) {
										case offer_item.Questionnaire_QUESTION_TYPE_RADIO:
											return q.Options
										case offer_item.Questionnaire_QUESTION_TYPE_TEXT:
											return nil
										default:
											return q.Options
										}
									}(),
								})
							}
							return res
						}(),
					},
				}
			}(),
			Assignees: func() []*offer_item.SaveAssignee {
				res := make([]*offer_item.SaveAssignee, 0)
				for _, a := range form.Assignees {
					res = append(res, &offer_item.SaveAssignee{
						AmebaId:    a.AmebaID,
						Stage:      a.Stage(),
						WritingFee: int64(a.WritingFee),
					})
				}
				return res
			}(),
		},
	}); err != nil {
		return nil, fmt.Errorf("a.offerItemHandlerClient.CreateOfferItem: %w", err)
	} else {
		id := model.OfferItemID(res.GetOfferItemId())
		return &id, nil
	}
}

// UpdateOfferItem implements adapter.OfferItemAdapter.
func (a *offerItemAdapterImpl) UpdateOfferItem(ctx context.Context, id model.OfferItemID, form dto.OfferItemForm) error {
	form.ID = id.String()
	if _, err := a.CreateOfferItem(ctx, form); err != nil {
		return fmt.Errorf("a.CreateOfferItem: %w", err)
	}
	return nil
}

func NewofferItemAdapter(config config.OfferItemGRPCServer) (adapter.OfferItemAdapter, error) {
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", config.Host, config.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//grpc.WithUnaryInterceptor(
		//	grpc_middleware.UnaryClientInterceptor("admin"),
		//),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{
			StartOptions: trace.StartOptions{
				Sampler: trace.ProbabilitySampler(0.5),
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc.Dial: %w", err)
	}
	return &offerItemAdapterImpl{
		offerItemHandlerClient: offer_item.NewOfferItemHandlerClient(conn),
	}, nil
}

func (a *offerItemAdapterImpl) ListAssignees(ctx context.Context, offerItemID model.OfferItemID, stage offer_item.Stage) ([]*offer_item.Assignee, error) {
	res, err := a.offerItemHandlerClient.ListAssignee(ctx, &offer_item.ListAssigneeRequest{
		OfferItemId: offerItemID.String(),
		Stage:       stage,
	})
	if err != nil {
		return nil, fmt.Errorf("a.offerItemHandlerClient.ListAssignee: %w", err)
	}
	if res.GetAssignees() == nil {
		return make([]*offer_item.Assignee, 0), nil
	}
	return res.GetAssignees(), nil
}

func (a *offerItemAdapterImpl) ListAssigneesUnderExamination(ctx context.Context) ([]*offer_item.Assignee, error) {
	res, err := a.offerItemHandlerClient.ListAssigneeUnderExamination(ctx, &offer_item.ListAssigneeUnderExaminationRequest{})
	if err != nil {
		return nil, fmt.Errorf("a.offerItemHandlerClient.ListAssigneeUnderExamination: %w", err)
	}
	return res.GetAssignees(), nil
}

func (a *offerItemAdapterImpl) BulkGetExaminations(ctx context.Context, offerItemID model.OfferItemID, entryType offer_item.EntryType) (map[model.AmebaID]*offer_item.Examination, error) {
	res, err := a.offerItemHandlerClient.BulkGetExaminations(ctx, &offer_item.BulkGetExaminationsRequest{
		OfferItemId: offerItemID.String(),
		EntryType:   entryType,
	})
	if err != nil {
		return nil, fmt.Errorf("a.offerItemHandlerClient.BulkGetExaminations: %w", err)
	}
	m := make(map[model.AmebaID]*offer_item.Examination)
	for k, v := range res.GetExaminationMap() {
		m[model.AmebaID(k)] = v
	}
	return m, nil
}

func (a *offerItemAdapterImpl) SaveExaminationResults(ctx context.Context, offerItemID model.OfferItemID, entryType offer_item.EntryType, forms []dto.ExaminationResultForm) error {
	_, err := a.offerItemHandlerClient.UploadExaminationResults(ctx, &offer_item.UploadExaminationResultsRequest{
		OfferItemId: offerItemID.String(),
		EntryType:   entryType,
		MapExaminationResults: func() map[string]*offer_item.ExaminationResult {
			res := make(map[string]*offer_item.ExaminationResult)
			for _, f := range forms {
				res[f.AmebaID] = &offer_item.ExaminationResult{
					IsPassed:     f.IsPassed,
					ExaminerName: f.ExaminerName,
					OptionalReason: func() *offer_item.ExaminationResult_Reason {
						if f.Reason == nil {
							return nil
						}
						return &offer_item.ExaminationResult_Reason{
							Reason: *f.Reason,
						}
					}(),
				}
			}
			return res
		}(),
	})
	if err != nil {
		return fmt.Errorf("a.offerItemHandlerClient.UploadExaminationResults: %w", err)
	}
	return nil
}

func (a *offerItemAdapterImpl) SavePaymentResults(ctx context.Context, offerItemID model.OfferItemID, amebaIDs []model.AmebaID) error {
	_, err := a.offerItemHandlerClient.PaymentCompleted(ctx, &offer_item.PaymentCompletedRequest{
		OfferItemId: offerItemID.String(),
		AmebaIds: func() []string {
			res := make([]string, 0, len(amebaIDs))
			for _, a := range amebaIDs {
				res = append(res, a.String())
			}
			return res
		}(),
	})
	if err != nil {
		return fmt.Errorf("a.offerItemHandlerClient.PaymentCompleted: %w", err)
	}
	return nil
}

func (a *offerItemAdapterImpl) BulkGetQuestionnaireQuestionAnswers(ctx context.Context, offerItemID model.OfferItemID, amebaIDs []model.AmebaID) (map[model.AmebaID]*offer_item.BulkGetQuestionnaireQuestionAnswersResponse_QuestionAnswerMap, error) {
	res, err := a.offerItemHandlerClient.BulkGetQuestionnaireQuestionAnswers(ctx, &offer_item.BulkGetQuestionnaireQuestionAnswersRequest{
		OfferItemId: offerItemID.String(),
		AmebaIds: func() []string {
			res := make([]string, 0, len(amebaIDs))
			for _, a := range amebaIDs {
				res = append(res, a.String())
			}
			return res
		}(),
	})
	if err != nil {
		if errors.Is(err, errors.New("OfferItemNotFoundError")) {
			return nil, adapter.ErrAdapterNotFound
		}
		return nil, fmt.Errorf("a.offerItemHandlerClient.BulkGetQuestionnaireQuestionAnswers: %w", err)
	}
	m := make(map[model.AmebaID]*offer_item.BulkGetQuestionnaireQuestionAnswersResponse_QuestionAnswerMap)
	for k, v := range res.GetAmebaIdAnswersMap() {
		m[model.AmebaID(k)] = v
	}
	return m, nil
}

func (a *offerItemAdapterImpl) FinishShipment(ctx context.Context, id model.OfferItemID) error {
	if _, err := a.offerItemHandlerClient.FinishedShipment(ctx, &offer_item.FinishedShipmentRequest{
		OfferItemId: id.String(),
	}); err != nil {
		return fmt.Errorf("a.offerItemHandlerClient.FinishedShipment: %w", err)
	}
	return nil
}
