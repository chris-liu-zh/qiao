/*
 * @Author: Chris
 * @Date: 2024-04-30 13:46:58
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-04 11:55:06
 * @Description: 请填写简介
 */
package DB

import (
	"fmt"
)

// SELECT
// 	bi_t_item_info.item_no,
// 	bi_t_item_info.item_subno,
// 	goods_details.recommend,
// 	goods_details.hot,
// 	goods_details.special_offer,
// 	goods_details.image_list,
// 	goods_details.details,
// 	goods_details.item_no
// FROM
// 	dbo.bi_t_item_info
// 	LEFT JOIN
// 	dbo.goods_details
// 	ON
// 		bi_t_item_info.item_no = goods_details.item_no
// WHERE
// 	goods_details.item_no IS NOT NULL

// join

const SelectJoin = "select ${field} from (select ${join_field} from ${table} ${join} ${where} ${order} ${group}) temp"

func (mapper *Mapper) Join(join, joinField, on string) *Mapper {
	mapper.Debris.joinField = joinField
	mapper.Debris.join = fmt.Sprintf("%s on %s", join, on)
	mapper.SqlTpl = SelectJoin
	return mapper
}
