package dao

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/database"
	"gateway/pkg/utils/huberrors"
	"gateway/pkg/utils/random"
	"gateway/web/views/hub0040/models"
)

// AuthTokenDAO 服务端颁发的数据面访问令牌，对应 HUB_SERVICE_AUTH_TOKEN。
type AuthTokenDAO struct {
	db database.Database
}

func NewAuthTokenDAO(db database.Database) *AuthTokenDAO {
	return &AuthTokenDAO{db: db}
}

// Issue 由服务端生成不透明令牌，绑定当前中心实例与环境，明文只在返回值里给一次。
func (dao *AuthTokenDAO) Issue(ctx context.Context, tenantID, instanceName, environment, tokenName, operatorID string, expireDays int) (*models.ServiceAuthToken, string, error) {
	if tenantID == "" || instanceName == "" || environment == "" {
		return nil, "", errors.New("实例名称和环境不能为空")
	}
	plain, err := newOpaqueTokenValue()
	if err != nil {
		return nil, "", huberrors.WrapError(err, "生成令牌失败")
	}
	scope, err := json.Marshal(map[string]string{
		"instanceName": instanceName,
		"environment":  environment,
	})
	if err != nil {
		return nil, "", huberrors.WrapError(err, "写入令牌绑定失败")
	}
	now := time.Now()
	row := &models.ServiceAuthToken{
		TokenId:        random.Generate32BitRandomString(),
		TenantId:       tenantID,
		TokenValue:     plain,
		UserId:         operatorID,
		TokenName:      tokenName,
		StatusFlag:     "Y",
		AddTime:        now,
		AddWho:         operatorID,
		EditTime:       now,
		EditWho:        operatorID,
		OprSeqFlag:     random.GenerateUniqueStringWithPrefix("", 32),
		CurrentVersion: 1,
		ActiveFlag:     "Y",
		ExtProperty:    string(scope),
	}
	if expireDays > 0 {
		exp := now.Add(time.Duration(expireDays) * 24 * time.Hour)
		row.ExpireTime = &exp
	}
	if _, err := dao.db.Insert(ctx, "HUB_SERVICE_AUTH_TOKEN", row, true); err != nil {
		return nil, "", huberrors.WrapError(err, "保存访问令牌失败")
	}
	return row, plain, nil
}

// ListByInstance 列出绑定到指定中心实例的令牌，不含明文。
func (dao *AuthTokenDAO) ListByInstance(ctx context.Context, tenantID, instanceName, environment string) ([]models.ServiceAuthToken, error) {
	query := `SELECT tokenId, tenantId, tokenValue, userId, tokenName, expireTime, statusFlag,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_SERVICE_AUTH_TOKEN
		WHERE tenantId = ? AND activeFlag = 'Y'
		ORDER BY addTime DESC`
	var rows []models.ServiceAuthToken
	if err := dao.db.Query(ctx, &rows, query, []interface{}{tenantID}, true); err != nil {
		return nil, huberrors.WrapError(err, "查询访问令牌失败")
	}
	out := make([]models.ServiceAuthToken, 0, len(rows))
	for _, row := range rows {
		scope := model.ParseTokenScope(row.ExtProperty)
		if scope.InstanceName == instanceName && scope.Environment == environment {
			out = append(out, row)
		}
	}
	return out, nil
}

// GetByID 按主键读取令牌。
func (dao *AuthTokenDAO) GetByID(ctx context.Context, tenantID, tokenID string) (*models.ServiceAuthToken, error) {
	query := `SELECT tokenId, tenantId, tokenValue, userId, tokenName, expireTime, statusFlag,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_SERVICE_AUTH_TOKEN
		WHERE tenantId = ? AND tokenId = ? AND activeFlag = 'Y'`
	var row models.ServiceAuthToken
	err := dao.db.QueryOne(ctx, &row, query, []interface{}{tenantID, tokenID}, true)
	if err == database.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, huberrors.WrapError(err, "查询访问令牌失败")
	}
	return &row, nil
}

// Revoke 禁用令牌，数据面立即不再接受。
func (dao *AuthTokenDAO) Revoke(ctx context.Context, tenantID, tokenID, operatorID string) error {
	current, err := dao.GetByID(ctx, tenantID, tokenID)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("令牌不存在")
	}
	current.StatusFlag = "N"
	current.EditTime = time.Now()
	current.EditWho = operatorID
	current.OprSeqFlag = random.GenerateUniqueStringWithPrefix("", 32)
	current.CurrentVersion = current.CurrentVersion + 1
	where := "tenantId = ? AND tokenId = ? AND currentVersion = ?"
	args := []interface{}{tenantID, tokenID, current.CurrentVersion - 1}
	n, err := dao.db.Update(ctx, "HUB_SERVICE_AUTH_TOKEN", current, where, args, true, false)
	if err != nil {
		return huberrors.WrapError(err, "吊销访问令牌失败")
	}
	if n == 0 {
		return errors.New("令牌已被其他用户修改，请刷新后重试")
	}
	return nil
}

func newOpaqueTokenValue() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "sct_" + hex.EncodeToString(buf), nil
}
