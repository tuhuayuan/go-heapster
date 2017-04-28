package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"
	"zonst/qipai/gamehealthysrv/middlewares"
	"zonst/qipai/gamehealthysrv/models"
)

// CreateHeapsterReq 创建请求
type CreateHeapsterReq struct {
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Port       int           `json:"port"`
	Timeout    time.Duration `json:"timeout"`
	Interval   time.Duration `json:"interval"`
	Threshold  int           `json:"threshold"`
	Groups     []string      `json:"groups"`
	Notifiers  []string      `json:"notifiers"`
	AcceptCode []int         `json:"accept_code,omitempty"`
	Host       string        `json:"host,omitempty"`
	Location   string        `json:"location,omitempty"`
}

// MuteHeapsterReq 静音请求
type MuteHeapsterReq struct {
	ID   string `json:"id" http:"id"`
	Mute bool   `json:"mute" http:"mute"`
}

// DeleteHeapsterReq 删除请求
type DeleteHeapsterReq struct {
	ID string `json:"id" http:"id"`
}

// UpdateHeapsterReq 更新请求
type UpdateHeapsterReq struct {
	ID string `json:"id" http:"id"`

	CreateHeapsterReq
}

// FetchHeapsterReq 查询请求
type FetchHeapsterReq struct {
	ID string `json:"id,omitempty" http:"id,omitempty"`
}

// FetchHeapsterStatusReq 获取状态请求
type FetchHeapsterStatusReq struct {
	ID string `json:"id,omitempty" http:"id,omitempty"`
}

// CreateHeapsterHandler 创建
func CreateHeapsterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*CreateHeapsterReq)

	model := &models.Heapster{
		ID:         models.NewSerialNumber(),
		Name:       req.Name,
		Type:       models.CheckType(req.Type),
		Port:       req.Port,
		Timeout:    req.Timeout * time.Second,
		Interval:   req.Interval * time.Second,
		Threshold:  req.Threshold,
		Groups:     req.Groups,
		Notifiers:  req.Notifiers,
		AcceptCode: req.AcceptCode,
		Host:       req.Host,
		Location:   req.Location,
	}

	if err := model.Save(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 2, err)
		return
	}
	if err := model.Fill(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 3, err)
		return
	}
	data, err := json.Marshal(model)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 4, err)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}

// DeleteHeapsterHandler 删除
func DeleteHeapsterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*DeleteHeapsterReq)

	model := &models.Heapster{
		ID: models.SerialNumber(req.ID),
	}
	if err = model.Delete(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 2, err)
		return
	}
	middlewares.ErrorWriteOK(w)
}

// UpdateHeapsterHandler 更新
func UpdateHeapsterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*UpdateHeapsterReq)

	model := &models.Heapster{
		ID: models.SerialNumber(req.ID),
	}
	if err := model.Fill(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 2, err)
		return
	}
	model.Name = req.Name
	model.Type = models.CheckType(req.Type)
	model.Port = req.Port
	model.Timeout = req.Timeout * time.Second
	model.Interval = req.Interval * time.Second
	model.Threshold = req.Threshold
	model.Groups = req.Groups
	model.Notifiers = req.Notifiers
	model.AcceptCode = req.AcceptCode
	model.Host = req.Host
	model.Location = req.Location
	if err := model.Save(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 3, err)
		return
	}
	middlewares.ErrorWriteOK(w)
}

// FetchHeapsterHandler 查询
func FetchHeapsterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*FetchHeapsterReq)

	var hs = make(models.Heapsters, 0, 256)

	if req.ID == "" {
		hset, err := models.FetchHeapsters(ctx)
		if err != nil {
			middlewares.ErrorWrite(w, 200, 2, err)
			return
		}
		for _, ht := range hset {
			hs = append(hs, ht)
		}
	} else {
		ht := &models.Heapster{
			ID: models.SerialNumber(req.ID),
		}
		if err := ht.Fill(ctx); err != nil {
			middlewares.ErrorWrite(w, 200, 3, err)
			return
		}
		hs = append(hs, *ht)
	}
	sort.Sort(hs)
	data, err := json.Marshal(hs)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 3, err)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}

// FetchHeapsterStatusHandler 批量获取状态
func FetchHeapsterStatusHandler(w http.ResponseWriter, r *http.Request) {
	var statusList []models.HeapsterStatusSet

	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*FetchHeapsterStatusReq)

	if req.ID == "" {
		var err error
		statusList, err = models.FetchHeapsterStatus(ctx)
		if err != nil {
			middlewares.ErrorWrite(w, 200, 2, err)
			return
		}
	} else {
		model := &models.Heapster{
			ID: models.SerialNumber(req.ID),
		}
		statusList = append(statusList, models.HeapsterStatusSet{
			ID:     model.ID,
			Status: model.GetStatus(ctx),
		})
	}

	data, err := json.Marshal(statusList)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 2, err)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}

// MuteHeapsterHandler 静音
func MuteHeapsterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := middlewares.GetBindBody(ctx)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 1, err)
		return
	}
	req := body.(*MuteHeapsterReq)

	model := &models.Heapster{
		ID: models.SerialNumber(req.ID),
	}
	if err := model.Fill(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 2, err)
		return
	}

	model.Mute = req.Mute
	if err := model.Save(ctx); err != nil {
		middlewares.ErrorWrite(w, 200, 3, err)
		return
	}
	data, err := json.Marshal(model)
	if err != nil {
		middlewares.ErrorWrite(w, 200, 4, err)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}
