// Package client 鎻愪緵MongoDB宸ュ叿鍑芥暟
//
// 姝ゆ枃浠跺寘鍚玀ongoDB鎿嶄綔鐩稿叧鐨勫伐鍏峰嚱鏁?
package client

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"gateway/pkg/mongo/errors"
	"gateway/pkg/mongo/types"
)

// === ObjectID宸ュ叿鍑芥暟 ===

// ObjectID 浠庡瓧绗︿覆鍒涘缓ObjectID
// 灏嗗瓧绗︿覆杞崲涓篗ongoDB ObjectID
func ObjectID(id string) (primitive.ObjectID, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, errors.NewValidationError("invalid ObjectID format", err)
	}
	return objectID, nil
}

// NewObjectID 鐢熸垚鏂扮殑ObjectID
// 鍒涘缓涓€涓柊鐨凪ongoDB ObjectID
func NewObjectID() primitive.ObjectID {
	return primitive.NewObjectID()
}

// IsValidObjectID 妫€鏌bjectID鏄惁鏈夋晥
// 楠岃瘉瀛楃涓叉槸鍚︿负鏈夋晥鐨凮bjectID鏍煎紡
func IsValidObjectID(id string) bool {
	return primitive.IsValidObjectID(id)
}

// === BSON杞崲宸ュ叿鍑芥暟 ===

// BSON 灏咲ocument杞崲涓篵son.D
// 鐢ㄤ簬闇€瑕佹湁搴忔枃妗ｇ殑鍦烘櫙
func BSON(doc types.Document) bson.D {
	d := make(bson.D, 0, len(doc))
	for k, v := range doc {
		d = append(d, bson.E{Key: k, Value: v})
	}
	return d
}

// M 灏咲ocument杞崲涓篵son.M
// 鐢ㄤ簬鏅€氱殑鏂囨。鎿嶄綔
func M(doc types.Document) bson.M {
	return bson.M(doc)
}

// === 鏃堕棿宸ュ叿鍑芥暟 ===

// Now 鑾峰彇褰撳墠鏃堕棿
// 杩斿洖褰撳墠UTC鏃堕棿
func Now() time.Time {
	return time.Now().UTC()
}
