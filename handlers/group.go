package handlers

import (
	"encoding/json"
	"net/http"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"
)

// CreateGroupReq 创建group请求模型
type CreateGroupReq struct {
	Name      string   `json:"name"`
	Endpoints []string `json:"endpoints"`
	Excluded  []string `json:"excluded"`
	Status    string   `json:"status,omitempty"`
}

// UpdateGroupReq 修改group请求模型
type UpdateGroupReq struct {
	CreateGroupReq

	ID string `json:"id"`
}

// DeleteGroupReq 删除请求
type DeleteGroupReq struct {
	ID string `json:"id" http:"id"`
}

// FetchGroupReq 查询请求
type FetchGroupReq struct {
	ID string `json:"id,omitempty" http:"id,omitempty"`
}

// CreateGroupHandler 创建
func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*CreateGroupReq)
	eps, err := models.ParseEndpoints(req.Endpoints, true)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 2, err)
	}
	excluded, err := models.ParseEndpoints(req.Excluded, true)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 3, err)
	}
	model := models.Group{
		ID:        models.NewSerialNumber(),
		Name:      req.Name,
		Endpoints: eps,
		Excluded:  excluded,
		Status:    models.GroupStatusEnable,
	}
	if err := model.Validate(); err != nil {
		middlewares.ErrorWrite(w, 200, 4, err)
		return
	}
	if err := model.Save(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 5, err)
		return
	}
	if err = model.Fill(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 6, err)
	}
	data, err := json.Marshal(model)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 7, err)
	}
	w.WriteHeader(200)
	w.Write(data)
}

// UpdateGroupHandler 更新
func UpdateGroupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*UpdateGroupReq)

	model := &models.Group{
		ID: models.SerialNumber(req.ID),
	}
	if err := model.Fill(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 2, err)
		return
	}
	model.Name = req.Name
	model.Status = models.GroupStatus(req.Status)
	model.Endpoints, err = models.ParseEndpoints(req.Endpoints, true)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 3, err)
		return
	}
	model.Excluded, err = models.ParseEndpoints(req.Excluded, true)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 4, err)
		return
	}
	if err := model.Validate(); err != nil {
		middlewares.ErrorWrite(w, 200, 5, err)
		return
	}
	if err := model.Save(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 6, err)
		return
	}
	middlewares.ErrorWriteOK(w)
}

// DeleteGroupHandler 删除
func DeleteGroupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*DeleteGroupReq)

	model := &models.Group{
		ID: models.SerialNumber(req.ID),
	}
	if err = model.Delete(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 2, err)
		return
	}
	middlewares.ErrorWriteOK(w)
}

// FetchGroupHandler 查询
func FetchGroupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*FetchGroupReq)

	var gs models.Groups

	if req.ID == "" {
		gs, err = models.FetchGroups(ctx)
		if err != nil {
			middlewares.ErrorWrite(w, 200, 2, err)
			return
		}
	} else {
		g := &models.Group{
			ID: models.SerialNumber(req.ID),
		}
		if err = g.Fill(ctx); err != nil {
			middlewares.ErrorWrite(w, 200, 2, err)
			return
		}
		gs = models.Groups{*g}
	}
	data, err := json.Marshal(gs)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 3, err)
	}
	w.WriteHeader(200)
	w.Write(data)
}
