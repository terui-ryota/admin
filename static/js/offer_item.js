const stageJapaneseMap = {
    "STAGE_BEFORE_INVITATION":"募集前",
    "STAGE_INVITATION":"参加募集",
    "STAGE_LOTTERY": "選考中",
    "STAGE_LOTTERY_LOST": "選考落ち",
    "STAGE_SHIPMENT": "発送中",
    "STAGE_DRAFT_SUBMISSION": "下書き提出",
    "STAGE_PRE_EXAMINATION": "下書き審査中",
    "STAGE_PRE_REEXAMINATION": "下書き再提出",
    "STAGE_ARTICLE_POSTING": "記事投稿",
    "STAGE_EXAMINATION": "記事審査中",
    "STAGE_REEXAMINATION": "記事再提出",
    "STAGE_PAYING": "支払い中",
    "STAGE_PAYMENT_COMPLETED": "支払い完了",
    "STAGE_DONE": "終了",
};

const japaneseStageMap = {
    // e.g.
    // "募集前": "STAGE_BEFORE_INVITATION"
};
for (let key in stageJapaneseMap) {
    japaneseStageMap[stageJapaneseMap[key]] = key;
}

const stages = {
    0: "STAGE_UNKNOWN",
    1: "STAGE_BEFORE_INVITATION",
    2: "STAGE_INVITATION",
    3: "STAGE_LOTTERY",
    4: "STAGE_LOTTERY_LOST",
    5: "STAGE_SHIPMENT",
    6: "STAGE_DRAFT_SUBMISSION",
    7: "STAGE_PRE_EXAMINATION",
    8: "STAGE_PRE_REEXAMINATION",
    9: "STAGE_ARTICLE_POSTING",
    10: "STAGE_EXAMINATION",
    11: "STAGE_REEXAMINATION",
    12: "STAGE_PAYING",
    13: "STAGE_PAYMENT_COMPLETED",
    14: "STAGE_DONE",
};

function amebaIdFormatter(value) {
    return `${value}`;
}

function dateFormatter(value) {
    return `${formatDate(new Date(value), "yyyy-MM-dd HH:mm:ss")}`;
}

function questionAnswersFormatter(value, row, index, field) {
    if (!row.QuestionAnswers) {
        return "";
    }
    for (let q of row.QuestionAnswers) {
        if (q.question_id === field) {
            return q.content;
        }
    }
    return "";
}

function couponBannerIdFormatter(value, row) {
    let couponBannerId = row.OfferItem.OptionalCouponBannerId?.CouponBannerId;
    if (!couponBannerId) {
        return "クーポン Pick ID が登録されていません。編集ページより設定してください。";
    }
    let entryText = row.ExaminationEntry?.entry_text;
    if (!entryText) {
        return "記事取得に失敗しました。記事が削除された可能性があります。";
    }
    if (entryText.includes(couponBannerId)) {
        return couponBannerId;
    }
    return "";
}

function existsPrImgUrlFormatter(value, row) {
    let entryText = row.ExaminationEntry?.entry_text;
    if (!entryText) {
        return "記事取得に失敗しました。記事が削除された可能性があります。";
    }
    if (entryText.includes("https://stat100.ameba.jp/blog/img/pr_mark_201507.gif")) {
        return "有";
    }
    return "無";
}

function itemUrlFormatter(value) {
    let url = '';
    if (value.OptionalDfItem) {
        url = value.OptionalDfItem.DfItem.urls.find(o => o.platform_type === 1)?.url
    } else {
        url = value.OptionalItem.Item.urls.find(o => o.platform_type === 1)?.url
    }
    return `<a href="${url}">${url}</a>`;
}

function stageFormatter(value) {
    return stageJapaneseMap[stages[value]];
}

function reasonFormatter(value) {
    if (!value) {
        return "";
    }
    return value.OptionalReason.Reason;
}

function examinationEntryFormatter(v, row, index, field) {
    let value = row.Examination;
    if (!value) {
        return ""
    }
    const entryUrl = `${amebaBlogUrl}/${value.ameba_id}/entry-${value.OptionalEntryId.EntryId}.html`;
    return `<a href="${entryUrl}" target="_blank">${entryUrl}</a>`;
}

function sendRemindMail(offerItemId, stage) {
    if (!confirm('実行しますか？この操作は戻せません。')) {
        return;
    }
    $.ajax({
        method: "POST",
        url: `${contextPath}/api/offer_item/${offerItemId}/stage/${stage}/send_remind_mail`,
        success: function(res) {
            console.log(res)
            alert('送りました')
        },
        error: function(res) {
            console.error(res)
            alert('送れませんでした')
        },
    });
};

function parseExaminationCsvText(text, withOfferItemId, withStage) {
    let csv;
    try {
        csv = $.csv.toArrays(text);
    } catch (e) {
        alert('CSV ファイルの読み込みに失敗しました。 CSV ファイルか確認してください。');
        throw new Error(`invalid input: ${e}`)
    }
    const data = [];
    $(csv).each((idx, cols) => {
        if (idx === 0) {
            return;
        }
        let passed = undefined;
        if (cols[1] === 'OK') {
            passed = true;
        } else if (cols[1] === 'NG') {
            passed = false;
        } else {
            alert(`「審査結果」の値が不正です。 OK または、NG のみ設定できます。行数: ${idx+1}, 不正な値: [[${cols[1]}]]`)
            throw new Error(`invalid input: ${cols[1]}`);
        }
        if (!passed && !cols[2]) {
            alert('審査 NG の場合、 NG の理由を「詳細コメント」へ入力してください。')
            throw new Error(`invalid input: ${cols[2]}`);
        }
        if (!cols[3]) {
            alert('審査者名がの値が未設定です。')
            throw new Error(`invalid input: ${cols[3]}`);
        }
        if (withOfferItemId) {
            if (!cols[5] && cols[5].length !== 22) {
                alert('PR投稿案件 ID の値が不正です')
                throw new Error(`invalid input: ${cols[5]}`);
            }
        }
        let stage = null;
        if (withStage) {
            stage = japaneseStageMap[cols[6]];
            if (!stage) {
                alert(`ステージが不正です。 ${cols[6]}`)
                throw new Error("failed")
            }
        }
        data.push({
            ameba_id: cols[0],
            is_passed: passed,
            reason: cols[2] != "" ? cols[2] : null,
            examiner_name: cols[3],
            offer_item_id: withOfferItemId ? cols[5] : null,
            stage: stage,
        })
    });
    console.log(data);
    return data;
}
