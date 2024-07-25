package handler

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/terui-ryota/admin/internal/app/admin-web/dto"
	"github.com/terui-ryota/admin/internal/domain/model"
	transformer "github.com/terui-ryota/admin/internal/lib/shiftjis_transformer"
	"github.com/terui-ryota/admin/internal/usecase"
	"github.com/terui-ryota/protofiles/go/offer_item"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type OfferItemHandler struct {
	offerItemUsecase usecase.OfferItemUsecase
}

func NewOfferItemHandler(
	offerItemUsecase usecase.OfferItemUsecase,
) *OfferItemHandler {
	return &OfferItemHandler{
		offerItemUsecase: offerItemUsecase,
	}
}

type SearchOfferItemParameter struct {
	ItemID     string `form:"item_id" json:"item_id"`
	DfItemID   string `form:"df_item_id" json:"df_item_id"`
	SearchText string `form:"search_text" json:"search_text"`
	Limit      uint   `form:"limit" json:"limit"`
	Offset     uint   `form:"offset" json:"offset"`
}

func (h *OfferItemHandler) GetOfferItemNewHTML(c *gin.Context) {
	data := c.Keys
	if id := c.Query("source_id"); id != "" {
		sourceID := model.OfferItemID(id)
		res, err := h.offerItemUsecase.GetOfferItemForm(c, sourceID)
		if err != nil {
			if errors.Is(err, usecase.ErrNotFound) {
				handle404HTML(c, err)
				return
			}
			handle5xxHTML(c, err)
			return
		}
		res.Name = "コピー_" + res.Name
		data["data"] = *res
	}
	c.HTML(http.StatusOK, "offer_item/edit.html", data)
}

func (h *OfferItemHandler) GetOfferItemEditHTML(c *gin.Context) {
	data := c.Keys
	id := model.OfferItemID(c.Param("offer_item_id"))
	res, err := h.offerItemUsecase.GetOfferItemForm(c, id)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			handle404HTML(c, err)
			return
		}
		handle5xxHTML(c, err)
		return
	}
	data["id"] = id
	data["data"] = *res
	c.HTML(http.StatusOK, "offer_item/edit.html", data)
}

func (h *OfferItemHandler) GetOfferItemDetailHTML(c *gin.Context) {
	data := c.Keys
	id := model.OfferItemID(c.Param("offer_item_id"))
	res, err := h.offerItemUsecase.GetOfferItem(c, id)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			handle404HTML(c, err)
			return
		}
		handle5xxHTML(c, err)
		return
	}
	m, err := h.offerItemUsecase.GetStageAssigneeCountMap(c, id)
	if err != nil {
		handle5xxHTML(c, err)
		return
	}
	// template(html) 内で扱いやすいように key を string に変換
	stageAssigneeCountMap := make(map[string]int)
	for k, v := range m {
		stageAssigneeCountMap[k.String()] = v
	}
	data["id"] = id
	data["data"] = res
	data["stageAssigneeCountMap"] = stageAssigneeCountMap
	c.HTML(http.StatusOK, "offer_item/detail.html", data)
}

func (h *OfferItemHandler) ListOfferItemsHTML(c *gin.Context) {
	data := c.Keys
	var p SearchOfferItemParameter
	if err := c.ShouldBindQuery(&p); err != nil {
		handle400JSON(c, err)
		return
	}
	data["parameter"] = p
	data["query"] = c.Request.URL.RawQuery
	data["postTargets"] = offer_item.PostTarget_name

	fmt.Println("======ここ通ってるよ======: ", data)
	c.HTML(http.StatusOK, "offer_item/index.html", data)
}

func (h *OfferItemHandler) GetOfferItemStageDetailHTML(c *gin.Context) {
	data := c.Keys
	id := model.OfferItemID(c.Param("offer_item_id"))
	stage := offer_item.Stage_name[offer_item.Stage_value[c.Param("stage")]]
	res, err := h.offerItemUsecase.GetOfferItem(c, id)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			handle404HTML(c, err)
			return
		}
		handle5xxHTML(c, err)
		return
	}
	data["id"] = id
	data["data"] = res
	data["stage"] = stage
	questionnaire, err := h.offerItemUsecase.GetQuestionnaire(c, id)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			data["questions"] = []*offer_item.Questionnaire_Question{}
		} else {
			handle5xxHTML(c, err)
			return
		}
	} else {
		data["questions"] = questionnaire.Questions
	}
	m, err := h.offerItemUsecase.GetStageAssigneeCountMap(c, id)
	if err != nil {
		handle5xxHTML(c, err)
		return
	}
	// template(html) 内で扱いやすいように key を string に変換
	stageAssigneeCountMap := make(map[string]int)
	for k, v := range m {
		stageAssigneeCountMap[k.String()] = v
	}
	data["stageAssigneeCountMap"] = stageAssigneeCountMap

	c.HTML(http.StatusOK, "offer_item/stage/"+strings.ToLower(stage)+".html", data)
}

type ListAssigneesParameter struct {
	Stage                   string `form:"stage"`
	WithOfferItem           bool   `form:"with_offer_item"`
	WithPersonalInformation bool   `form:"with_personal_information"`
	WithQuestionAnswers     bool   `form:"with_question_answers"`
	WithBlogger             bool   `form:"with_blogger"`
	WithAsID                bool   `form:"with_as_id"`
	EntryType               string `form:"entry_type"`
	WithoutFutureEntry      bool   `form:"without_future_entry"`
}

func (h *OfferItemHandler) ListAssigneesJSON(c *gin.Context) {
	var parameter ListAssigneesParameter
	if err := c.ShouldBindQuery(&parameter); err != nil {
		handle400JSON(c, err)
		return
	}
	offerItemID := model.OfferItemID(c.Param("offer_item_id"))
	var stage *offer_item.Stage
	if parameter.Stage != "" {
		s := offer_item.Stage_value[parameter.Stage]
		stage = (*offer_item.Stage)(&s)
	}
	var entryType *offer_item.EntryType
	if parameter.EntryType != "" {
		e := offer_item.EntryType_value[parameter.EntryType]
		entryType = (*offer_item.EntryType)(&e)
	}
	res, err := h.offerItemUsecase.ListAllAssignees(
		c,
		offerItemID,
		usecase.ListAssigneesParameter{
			Stage:                   stage,
			WithOfferItem:           parameter.WithOfferItem,
			WithPersonalInformation: parameter.WithPersonalInformation,
			WithQuestionAnswers:     parameter.WithQuestionAnswers,
			WithBlogger:             parameter.WithBlogger,
			WithAsID:                parameter.WithAsID,
			EntryType:               entryType,
			WithoutFutureEntry:      parameter.WithoutFutureEntry,
		},
	)
	if err != nil {
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  res,
		"total": len(res),
	})
}

func (h *OfferItemHandler) ListAssigneesUnderExaminationJSON(c *gin.Context) {
	limitStr := c.Query("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		handle400JSON(c, err)
		return
	}
	res, err := h.offerItemUsecase.ListAssigneesUnderExamination(c, limit)
	if err != nil {
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": res,
	})
}
func (h *OfferItemHandler) SearchOfferItemsJSON(c *gin.Context) {
	var p SearchOfferItemParameter
	if err := c.ShouldBindQuery(&p); err != nil {
		handle400JSON(c, err)
		return
	}
	var res []usecase.OfferItem
	var total uint
	var err error
	if p.SearchText == "" && p.ItemID == "" && p.DfItemID == "" {
		res, total, err = h.offerItemUsecase.ListOfferItems(c, p.Limit, p.Offset)
	} else {
		res, total, err = h.offerItemUsecase.SearchOfferItems(
			c,
			func() *string {
				if p.SearchText == "" {
					return nil
				}
				return &p.SearchText
			}(),
			func() *model.ItemID {
				if p.ItemID == "" {
					return nil
				}
				return (*model.ItemID)(&p.ItemID)
			}(),
			func() *model.DfItemID {
				if p.DfItemID == "" {
					return nil
				}
				return (*model.DfItemID)(&p.DfItemID)
			}(),
			p.Limit,
			p.Offset,
		)
	}
	if err != nil {
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  res,
		"total": total,
	})
}

func (h *OfferItemHandler) GetOfferItem(c *gin.Context) {
	offerItemID := model.OfferItemID(c.Param("offer_item_id"))
	offerItem, err := h.offerItemUsecase.GetOfferItem(c, offerItemID)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{})
			return
		}
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"data": offerItem,
	})
}

func (h *OfferItemHandler) CreateOfferItem(c *gin.Context) {
	var form dto.OfferItemForm
	if err := c.BindJSON(&form); err != nil {
		handle400JSON(c, err)
		return
	}
	id, err := h.offerItemUsecase.CreateOfferItem(c, form)
	if err != nil {
		if errors.Is(err, errors.New("エラー発生！")) || errors.Is(err, usecase.ErrAffiliatorNotFound) {
			handle400JSON(c, err)
			return
		}
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"data": id,
	})
}

func (h *OfferItemHandler) UpdateOfferItem(c *gin.Context) {
	id := model.OfferItemID(c.Param("offer_item_id"))
	var form dto.OfferItemForm
	if err := c.BindJSON(&form); err != nil {
		handle400JSON(c, err)
		return
	}
	if err := h.offerItemUsecase.UpdateOfferItem(c, id, form); err != nil {
		if errors.Is(err, errors.New("エラー発生！")) || errors.Is(err, usecase.ErrAffiliatorNotFound) {
			handle400JSON(c, err)
			return
		}
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{})
}

func (h *OfferItemHandler) DeleteOfferItem(c *gin.Context) {
	id := model.OfferItemID(c.Param("offer_item_id"))
	if err := h.offerItemUsecase.DeleteOfferItem(c, id); err != nil {
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}

func (h *OfferItemHandler) Invite(c *gin.Context) {
	id := model.OfferItemID(c.Param("offer_item_id"))
	if err := h.offerItemUsecase.Invite(c, id); err != nil {
		if errors.Is(err, errors.New("エラー発生！")) {
			handle400JSON(c, err)
			return
		}
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{})
}

func (h *OfferItemHandler) SendRemindMail(c *gin.Context) {
	stage, ok := offer_item.Stage_value[c.Param("stage")]
	if !ok {
		handle400JSON(c, fmt.Errorf("unknown stage: %s", c.Param("stage")))
		return
	}
	if err := h.offerItemUsecase.SendRemindMail(c, model.OfferItemID(c.Param("offer_item_id")), offer_item.Stage(stage)); err != nil {
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *OfferItemHandler) SavePreExaminationResults(c *gin.Context) {
	var body []dto.ExaminationResultForm
	if err := c.ShouldBindJSON(&body); err != nil {
		handle400JSON(c, err)
		return
	}
	id := model.OfferItemID(c.Param("offer_item_id"))
	if err := h.offerItemUsecase.SavePreExaminationResults(c, id, body); err != nil {
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *OfferItemHandler) SaveExaminationResults(c *gin.Context) {
	var body []dto.ExaminationResultForm
	if err := c.ShouldBindJSON(&body); err != nil {
		handle400JSON(c, err)
		return
	}
	id := model.OfferItemID(c.Param("offer_item_id"))
	if err := h.offerItemUsecase.SaveExaminationResults(c, id, body); err != nil {
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *OfferItemHandler) SaveLotteryResults(c *gin.Context) {
	var body []dto.LotteryResultForm
	if err := c.ShouldBindJSON(&body); err != nil {
		handle400JSON(c, err)
		return
	}
	if err := h.offerItemUsecase.SaveLotteryResults(c, model.OfferItemID(c.Param("offer_item_id")), body); err != nil {
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *OfferItemHandler) SavePaymentResults(c *gin.Context) {
	var body []model.AmebaID
	if err := c.ShouldBindJSON(&body); err != nil {
		handle400JSON(c, err)
		return
	}
	if err := h.offerItemUsecase.SavePaymentResults(c, model.OfferItemID(c.Param("offer_item_id")), body); err != nil {
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *OfferItemHandler) Close(c *gin.Context) {
	if err := h.offerItemUsecase.Close(c, model.OfferItemID(c.Param("offer_item_id"))); err != nil {
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *OfferItemHandler) FinishShipment(c *gin.Context) {
	if err := h.offerItemUsecase.FinishShipment(c, model.OfferItemID(c.Param("offer_item_id"))); err != nil {
		handle5xxJSON(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

type DownloadShipmentZipRequest struct {
	Name string `form:"name" json:"name"`
}

func (h *OfferItemHandler) DownloadShipmentPreview(c *gin.Context) {
	stage := offer_item.Stage_STAGE_SHIPMENT
	parameter := DownloadShipmentZipRequest{}
	if err := c.ShouldBindQuery(&parameter); err != nil {
		handle400JSON(c, err)
		return
	}
	offerItemID := model.OfferItemID(c.Param("offer_item_id"))
	assignees, err := h.offerItemUsecase.ListAllAssignees(
		c,
		offerItemID,
		usecase.ListAssigneesParameter{
			Stage:                   &stage,
			WithOfferItem:           true,
			WithPersonalInformation: true,
		},
	)
	if err != nil {
		handle5xxHTML(c, err)
		return
	}

	csvBuf, err := generateShipmentCSV(assignees, false)
	if err != nil {
		handle5xxHTML(c, err)
		return
	}

	c.Header("Content-Type", "application/csv")
	if _, err := io.Copy(c.Writer, csvBuf); err != nil {
		handle5xxHTML(c, err)
		return
	}
}

//func (h *OfferItemHandler) DownloadShipmentZip(c *gin.Context) {
//	// NOTE: このエンドポイントでは暗号化した個人情報をDLするため、誰でも叩けるようにします。
//	// 以下の関数を呼び出しadapterでPermission確認を行わないようにします。
//	c.Request = c.Request.WithContext(
//		common.SetPassingPersonalInformationPermissionValidation(c.Request.Context()))
//	//
//
//	stage := offer_item.Stage_STAGE_SHIPMENT
//	parameter := DownloadShipmentZipRequest{}
//	if err := c.ShouldBindQuery(&parameter); err != nil {
//		handle400JSON(c, err)
//		return
//	}
//	offerItemID := model.OfferItemID(c.Param("offer_item_id"))
//	assignees, err := h.offerItemUsecase.ListAllAssignees(
//		c,
//		offerItemID,
//		usecase.ListAssigneesParameter{
//			Stage:                   &stage,
//			WithOfferItem:           true,
//			WithPersonalInformation: true,
//		},
//	)
//	if err != nil {
//		handle5xxHTML(c, err)
//		return
//	}
//
//	csvBuf, err := generateShipmentCSV(assignees, true)
//	if err != nil {
//		handle5xxHTML(c, err)
//		return
//	}
//	c.Header("Content-Type", "application/zip")
//	if err := writeZipWithPassword(c.Writer, csvBuf, parameter.Name, len(assignees), h.shipmentZipPassword); err != nil {
//		handle5xxHTML(c, err)
//		return
//	}
//}

func generateShipmentCSV(assignees usecase.Assignees, withPersonalInfo bool) (io.Reader, error) {
	buf := new(bytes.Buffer)
	csvWriter := csv.NewWriter(buf)

	shippingDataLength := 0
	hasJanCode := false
	//for _, assignee := range assignees {
	//	a := assignee.Assignee
	//	//if a.GetOptionalJanCode() != nil {
	//	//	hasJanCode = true
	//	//}
	//	//shippingData := a.GetShippingData()
	//	if shippingData == nil {
	//		continue
	//	}
	//	// カラム数が違うことはないとは思うが、最大の長さを入れておく
	//	if shippingDataLength < len(shippingData.Data) {
	//		shippingDataLength = len(shippingData.Data)
	//	}
	//}
	makeShippingDataHeader := func(l int) []string {
		resp := []string{}
		for i := 0; i < l; i++ {
			resp = append(resp, "発送した商品")
		}
		return resp
	}
	makeJanCodeHeader := func(hasJanCode bool) []string {
		if hasJanCode {
			return []string{"ジャンコード"}
		}
		return []string{}
	}
	transformShiftJIS := func(ss []string) ([]string, error) {
		resp := make([]string, 0, len(ss))
		for _, s := range ss {
			transformer := transformer.NewTransformer(japanese.ShiftJIS, '?')
			st, _, err := transform.String(transformer, s)
			if err != nil {
				return nil, fmt.Errorf("transform.String: %w", err)
			}
			resp = append(resp, st)
		}
		return resp, nil
	}
	header := []string{
		"PR投稿案件名",
		"アメーバID",
		"ステージ",
	}
	withPersonalInfoHeader := []string{
		"姓",
		"名",
		"姓_ひらがな",
		"名_ひらがな",
		"郵便番号",
		"住所",
		"電話番号",
	}
	if withPersonalInfo { // 個人情報が必要ならカラム追加、必要なしならカラム追加しない
		header = append(header, withPersonalInfoHeader...)
	}
	header = append(header, makeShippingDataHeader(shippingDataLength)...)
	header = append(header, makeJanCodeHeader(hasJanCode)...)
	header, err := transformShiftJIS(header)
	if err != nil {
		return nil, fmt.Errorf("transformShitJIS: %w", err)
	}
	if err := csvWriter.Write(header); err != nil {
		return nil, fmt.Errorf("csvWriter.Write: %w", err)
	}

	for _, assignee := range assignees {
		body := []string{
			assignee.OfferItem.GetName(),
			assignee.Assignee.GetAmebaId(),
			"発送中", // この関数は全部これなので決め打ちにする(不要なら削除)
		}
		//if withPersonalInfo { // 個人情報が必要ならカラム追加、必要なしならカラム追加しない
		//	body = append(body,
		//		assignee.PersonalInformation.Name.GetFamilyName(),
		//		assignee.PersonalInformation.Name.GetGivenName(),
		//		assignee.PersonalInformation.Name.GetFamilyNameKana(),
		//		assignee.PersonalInformation.Name.GetGivenNameKana(),
		//		assignee.PersonalInformation.GetPostCode(),
		//		assignee.PersonalInformation.GetAddress(),
		//		assignee.PersonalInformation.GetPhoneNumber(),
		//	)
		//}
		//body = append(body, assignee.Assignee.GetShippingData().GetData()...)
		//janCode := func() []string {
		//	if janCode := assignee.Assignee.GetJanCode(); janCode != "" {
		//		return []string{janCode}
		//	}
		//	return []string{}
		//}()
		//body = append(body, janCode...)
		body, err := transformShiftJIS(body)
		if err != nil {
			return nil, fmt.Errorf("transformShitJIS: %w", err)
		}
		if err := csvWriter.Write(body); err != nil {
			return nil, fmt.Errorf("csvWriter.Write: %w", err)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("csv.Flush: %w", err)
	}

	return buf, nil
}

//func writeZipWithPassword(w io.Writer, content io.Reader, name string, n int, password string) error {
//	zipWriter := zip.NewWriter(w)
//	defer zipWriter.Close()
//
//	filename := fmt.Sprintf("【配送依頼】AmebaPick案件（%s）_%d名.csv", name, n)
//	w, err := zipWriter.Encrypt(filename, password, zip.AES256Encryption)
//	if err != nil {
//		return fmt.Errorf("zipWriter.Encrypt: %w", err)
//	}
//
//	if _, err := io.Copy(w, content); err != nil {
//		return fmt.Errorf("io.Copy: %w", err)
//	}
//
//	return nil
//}
