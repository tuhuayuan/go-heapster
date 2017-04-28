package handlers

import (
	"encoding/json"
	"net/http"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"
)

// CreateNotifierReq 创建请求
type CreateNotifierReq struct {
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"Config"`
}

// DeleteNotifierReq 删除请求
type DeleteNotifierReq struct {
	ID string `json:"id" http:"id"`
}

// UpdateNotifierReq 更新请求
type UpdateNotifierReq struct {
	ID string `json:"id"`

	CreateNotifierReq
}

// FetchNotifierReq 获取请求
type FetchNotifierReq struct {
	ID string `json:"id,omitempty" http:"id,omitempty"`
}

// CreateNotifierHandler 创建
func CreateNotifierHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*CreateNotifierReq)

	model := &models.HeapsterNotifier{
		ID:     models.NewSerialNumber(),
		Type:   req.Type,
		Name:   req.Name,
		Config: req.Config,
	}

	if err := model.Save(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 2, err)
	}
	if err := model.Fill(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 3, err)
	}
	data, err := json.Marshal(model)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 4, err)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}

// DeleteNotifierHandler 删除
func DeleteNotifierHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*DeleteNotifierReq)

	model := &models.HeapsterNotifier{
		ID: models.SerialNumber(req.ID),
	}
	if err = model.Delete(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 2, err)
		return
	}
	middlewares.ErrorWriteOK(w)
}

// UpdateNotifierHandler 更新
func UpdateNotifierHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*UpdateNotifierReq)

	model := &models.HeapsterNotifier{
		ID: models.SerialNumber(req.ID),
	}
	if err := model.Fill(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 2, err)
		return
	}
	model.Type = req.Type
	model.Name = req.Name
	model.Config = req.Config
	if err := model.Save(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 3, err)
		return
	}
	middlewares.ErrorWriteOK(w)
}

// FetchNotifierHandler 查询
func FetchNotifierHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*FetchNotifierReq)

	var nts models.HeapsterNotifiers

	if req.ID == "" {
		nts, err = models.FetchHeapsterNotifiers(ctx)
		if err != nil {
			middlewares.ErrorWrite(w, 200, 2, err)
			return
		}
	} else {
		nt := &models.HeapsterNotifier{
			ID: models.SerialNumber(req.ID),
		}
		if err = nt.Fill(ctx); err != nil {
			middlewares.ErrorWrite(w, 200, 2, err)
			return
		}
		nts = models.HeapsterNotifiers{*nt}
	}

	data, err := json.Marshal(nts)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 3, err)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}
