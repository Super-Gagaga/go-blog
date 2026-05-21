package repository

import (
	"github.com/yourname/go-blog/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// toggleInteraction 是 like/favorite 两张复合主键关联表的公共切换逻辑:
// 先按 (user_id, article_id) 删除,删掉说明原来已存在(取消),
// 没删掉则插入(并发冲突时忽略并保持已存在),只有真正插入成功才累加计数。
func toggleInteraction(
	tx *gorm.DB,
	newRecord func() interface{},
	countColumn string,
	userID, articleID int64,
) (bool, error) {
	deleted := tx.Where("user_id = ? AND article_id = ?", userID, articleID).Delete(newRecord())
	if deleted.Error != nil {
		return false, deleted.Error
	}
	if deleted.RowsAffected > 0 {
		return false, tx.Model(&model.Article{}).
			Where("id = ? AND "+countColumn+" > 0", articleID).
			UpdateColumn(countColumn, gorm.Expr(countColumn+" - ?", 1)).
			Error
	}

	inserted := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(newRecord())
	if inserted.Error != nil {
		return false, inserted.Error
	}
	if inserted.RowsAffected == 0 {
		return true, nil
	}
	return true, tx.Model(&model.Article{}).
		Where("id = ?", articleID).
		UpdateColumn(countColumn, gorm.Expr(countColumn+" + ?", 1)).
		Error
}
