package handler

import (
    "encoding/json"
    "fmt"
    "jobView-backend/internal/auth"
    "jobView-backend/internal/model"
    "jobView-backend/internal/service"
    "net/http"
    "strconv"
    "strings"

    "github.com/gorilla/mux"
)

type ResumeHandler struct{ svc *service.ResumeService }

func NewResumeHandler(s *service.ResumeService) *ResumeHandler { return &ResumeHandler{svc:s} }

// absoluteStaticURL 构造静态资源的绝对 URL
func absoluteStaticURL(r *http.Request, url string) string {
    if strings.HasPrefix(url, "/static/") {
        scheme := "http"; if r.TLS!=nil { scheme="https" }
        return scheme+"://"+r.Host+url
    }
    return url
}

func (h *ResumeHandler) GetMyResume(w http.ResponseWriter, r *http.Request) {
    uid, ok := auth.GetUserIDFromContext(r.Context()); if !ok { h.writeErrorResponse(w,http.StatusUnauthorized,"未登录",nil); return }
    resume, err := h.svc.EnsureUserResume(r.Context(), uint(uid)); if err!=nil { h.writeErrorResponse(w, http.StatusInternalServerError, "获取简历失败", err); return }
    sections, _ := h.svc.ListSections(r.Context(), uint(uid), resume.ID)
    types := []string{}
    for _, s := range sections { types = append(types, s.Type) }
    summary := model.ResumeSummary{ Resume:*resume, SectionTypes:types }
    h.writeSuccessResponse(w, http.StatusOK, "ok", summary)
}

func (h *ResumeHandler) Create(w http.ResponseWriter, r *http.Request) {
    uid, ok := auth.GetUserIDFromContext(r.Context()); if !ok { h.writeErrorResponse(w,http.StatusUnauthorized,"未登录",nil); return }
    resume, err := h.svc.EnsureUserResume(r.Context(), uint(uid)); if err!=nil { h.writeErrorResponse(w, http.StatusInternalServerError, "创建简历失败", err); return }
    h.writeSuccessResponse(w, http.StatusCreated, "created", resume)
}

func (h *ResumeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    uid, ok := auth.GetUserIDFromContext(r.Context()); if !ok { h.writeErrorResponse(w,http.StatusUnauthorized,"未登录",nil); return }
    id, _ := strconv.Atoi(mux.Vars(r)["id"])
    agg, err := h.svc.GetResumeAggregate(r.Context(), uint(uid), id); if err!=nil { h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil); return }
    h.writeSuccessResponse(w, http.StatusOK, "ok", agg)
}

func (h *ResumeHandler) Update(w http.ResponseWriter, r *http.Request) {
    uid, ok := auth.GetUserIDFromContext(r.Context()); if !ok { h.writeErrorResponse(w,http.StatusUnauthorized,"未登录",nil); return }
    id, _ := strconv.Atoi(mux.Vars(r)["id"])
    var body struct{ Title *string `json:"title"`; Privacy *string `json:"privacy"` }
    if err := json.NewDecoder(r.Body).Decode(&body); err!=nil { h.writeErrorResponse(w,http.StatusBadRequest,"请求体错误",err); return }
    res, err := h.svc.UpdateMetadata(r.Context(), uint(uid), id, body.Title, body.Privacy); if err!=nil { h.writeErrorResponse(w,http.StatusBadRequest,err.Error(),nil); return }
    h.writeSuccessResponse(w, http.StatusOK, "updated", res)
}

func (h *ResumeHandler) Delete(w http.ResponseWriter, r *http.Request) {
    // 预留：当前实现不做软删，直接 200
    h.writeSuccessResponse(w, http.StatusOK, "deleted", nil)
}

func (h *ResumeHandler) ListSections(w http.ResponseWriter, r *http.Request) {
    uid, ok := auth.GetUserIDFromContext(r.Context()); if !ok { h.writeErrorResponse(w,http.StatusUnauthorized,"未登录",nil); return }
    id, _ := strconv.Atoi(mux.Vars(r)["id"])
    list, err := h.svc.ListSections(r.Context(), uint(uid), id); if err!=nil { h.writeErrorResponse(w,http.StatusBadRequest,err.Error(),nil); return }
    h.writeSuccessResponse(w, http.StatusOK, "ok", list)
}

func (h *ResumeHandler) UpsertSection(w http.ResponseWriter, r *http.Request) {
    uid, ok := auth.GetUserIDFromContext(r.Context()); if !ok { h.writeErrorResponse(w,http.StatusUnauthorized,"未登录",nil); return }
    id, _ := strconv.Atoi(mux.Vars(r)["id"])
    typ := mux.Vars(r)["type"]
    var raw json.RawMessage
    if err := json.NewDecoder(r.Body).Decode(&raw); err!=nil { h.writeErrorResponse(w,http.StatusBadRequest,"JSON错误",err); return }
    sct, err := h.svc.UpsertSection(r.Context(), uint(uid), id, typ, raw); if err!=nil { h.writeErrorResponse(w,http.StatusBadRequest,err.Error(),nil); return }
    h.writeSuccessResponse(w, http.StatusOK, "ok", sct)
}

func (h *ResumeHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
    uid, ok := auth.GetUserIDFromContext(r.Context()); if !ok { h.writeErrorResponse(w,http.StatusUnauthorized,"未登录",nil); return }
    id, _ := strconv.Atoi(mux.Vars(r)["id"])
    if err := r.ParseMultipartForm(10<<20); err!=nil { h.writeErrorResponse(w,http.StatusBadRequest,"无法解析表单",err); return }
    file, header, err := r.FormFile("file"); if err!=nil { h.writeErrorResponse(w,http.StatusBadRequest,"缺少文件",err); return }
    defer file.Close()
    att, url, err := h.svc.UploadAttachment(r.Context(), uint(uid), id, file, header.Filename, header.Header.Get("Content-Type"))
    if err!=nil { h.writeErrorResponse(w,http.StatusBadRequest,err.Error(),nil); return }
    // 绝对 URL
    url = absoluteStaticURL(r, url)
    h.writeSuccessResponse(w, http.StatusOK, "上传成功", map[string]interface{}{"attachment": att, "url": url})
}

// ListAttachments 返回当前简历的附件列表（包含绝对URL）
func (h *ResumeHandler) ListAttachments(w http.ResponseWriter, r *http.Request) {
    uid, ok := auth.GetUserIDFromContext(r.Context()); if !ok { h.writeErrorResponse(w,http.StatusUnauthorized,"未登录",nil); return }
    id, _ := strconv.Atoi(mux.Vars(r)["id"])
    list, err := h.svc.ListAttachments(r.Context(), uint(uid), id); if err!=nil { h.writeErrorResponse(w,http.StatusBadRequest,err.Error(),nil); return }
    // 补充绝对URL
    out := make([]map[string]interface{}, 0, len(list))
    for _, a := range list {
        url := a.FilePath
        if !strings.HasPrefix(url, "/static/") { url = "/static/" + strings.TrimLeft(url, "/") }
        abs := absoluteStaticURL(r, url)
        out = append(out, map[string]interface{}{
            "id": a.ID,
            "resume_id": a.ResumeID,
            "file_name": a.FileName,
            "file_path": a.FilePath,
            "mime_type": a.MimeType,
            "file_size": a.FileSize,
            "etag": a.ETag,
            "created_at": a.CreatedAt,
            "url": abs,
        })
    }
    h.writeSuccessResponse(w, http.StatusOK, "ok", out)
}

// 复用 auth_handler 的响应函数
func (h *ResumeHandler) writeSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    resp := model.APIResponse{ Code: statusCode, Message: message, Data: data }
    _ = json.NewEncoder(w).Encode(resp)
}

func (h *ResumeHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    resp := model.APIResponse{ Code: statusCode, Message: message }
    if err!=nil && statusCode>=500 { resp.Data = map[string]string{"error": fmt.Sprintf("%v", err)} }
    _ = json.NewEncoder(w).Encode(resp)
}
