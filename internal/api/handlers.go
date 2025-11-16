package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ilya2044/avito2025/internal/model"
	"github.com/ilya2044/avito2025/internal/service"
)

type Handler struct {
	Svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{Svc: svc}
}

type ErrResp struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var t model.Team
	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &t)
	if t.TeamName == "" {
		writeJSON(w, 400, map[string]string{"error": "team_name required"})
		return
	}
	err := h.Svc.CreateTeam(t)
	if err != nil {
		er := ErrResp{}
		er.Error.Code = "TEAM_EXISTS"
		er.Error.Message = err.Error()
		writeJSON(w, 400, er)
		return
	}
	writeJSON(w, 201, map[string]model.Team{"team": t})
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("team_name")
	if q == "" {
		writeJSON(w, 400, map[string]string{"error": "team_name required"})
		return
	}
	t, err := h.Svc.GetTeam(q)
	if err != nil {
		er := ErrResp{}
		er.Error.Code = "NOT_FOUND"
		er.Error.Message = err.Error()
		writeJSON(w, 404, er)
		return
	}
	writeJSON(w, 200, t)
}

func (h *Handler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.UserID == "" {
		writeJSON(w, 400, map[string]string{"error": "user_id required"})
		return
	}
	u, err := h.Svc.SetUserIsActive(req.UserID, req.IsActive)
	if err != nil {
		er := ErrResp{}
		er.Error.Code = "NOT_FOUND"
		er.Error.Message = err.Error()
		writeJSON(w, 404, er)
		return
	}
	writeJSON(w, 200, map[string]model.User{"user": u})
}

func (h *Handler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		writeJSON(w, 400, map[string]string{"error": "invalid"})
		return
	}
	pr := model.PullRequest{
		PullRequestID:   req.PullRequestID,
		PullRequestName: req.PullRequestName,
		AuthorID:        req.AuthorID,
	}
	created, err := h.Svc.CreatePullRequest(pr)
	if err != nil {
		er := ErrResp{}
		switch err {
		case service.ErrPRExists:
			er.Error.Code = "PR_EXISTS"
			er.Error.Message = err.Error()
			writeJSON(w, 409, er)
		case service.ErrTeamNotFound:
			er.Error.Code = "NOT_FOUND"
			er.Error.Message = err.Error()
			writeJSON(w, 404, er)
		default:
			er.Error.Code = "NOT_FOUND"
			er.Error.Message = err.Error()
			writeJSON(w, 404, er)
		}
		return
	}
	writeJSON(w, 201, map[string]model.PullRequest{"pr": created})
}

func (h *Handler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.PullRequestID == "" {
		writeJSON(w, 400, map[string]string{"error": "pull_request_id required"})
		return
	}
	pr, err := h.Svc.MergePullRequest(req.PullRequestID)
	if err != nil {
		er := ErrResp{}
		er.Error.Code = "NOT_FOUND"
		er.Error.Message = err.Error()
		writeJSON(w, 404, er)
		return
	}
	writeJSON(w, 200, map[string]model.PullRequest{"pr": pr})
}

func (h *Handler) Reassign(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.PullRequestID == "" || req.OldUserID == "" {
		writeJSON(w, 400, map[string]string{"error": "invalid"})
		return
	}
	pr, replacedBy, err := h.Svc.ReassignReviewer(req.PullRequestID, req.OldUserID)
	if err != nil {
		er := ErrResp{}
		switch err {
		case service.ErrPRMerged:
			er.Error.Code = "PR_MERGED"
			er.Error.Message = err.Error()
			writeJSON(w, 409, er)
		case service.ErrNotAssigned:
			er.Error.Code = "NOT_ASSIGNED"
			er.Error.Message = err.Error()
			writeJSON(w, 409, er)
		case service.ErrNoCandidate:
			er.Error.Code = "NO_CANDIDATE"
			er.Error.Message = err.Error()
			writeJSON(w, 409, er)
		default:
			er.Error.Code = "NOT_FOUND"
			er.Error.Message = err.Error()
			writeJSON(w, 404, er)
		}
		return
	}
	writeJSON(w, 200, map[string]interface{}{"pr": pr, "replaced_by": replacedBy})
}

func (h *Handler) GetReviews(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("user_id")
	if uid == "" {
		writeJSON(w, 400, map[string]string{"error": "user_id required"})
		return
	}
	list, err := h.Svc.GetPRsByReviewer(uid)
	if err != nil {
		er := ErrResp{}
		er.Error.Code = "NOT_FOUND"
		er.Error.Message = err.Error()
		writeJSON(w, 404, er)
		return
	}
	writeJSON(w, 200, map[string]interface{}{"user_id": uid, "pull_requests": list})
}

func (h *Handler) AddUserToTeam(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TeamName string     `json:"team_name"`
		User     model.User `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid"})
		return
	}
	if req.TeamName == "" || req.User.UserID == "" {
		writeJSON(w, 400, map[string]string{"error": "team_name and user_id required"})
		return
	}
	team, err := h.Svc.AddUserToTeam(req.TeamName, req.User)
	if err != nil {
		er := ErrResp{}
		er.Error.Code = "ERROR"
		er.Error.Message = err.Error()
		writeJSON(w, 400, er)
		return
	}
	writeJSON(w, 200, map[string]model.Team{"team": team})
}

func (h *Handler) RemoveUserFromTeam(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TeamName string `json:"team_name"`
		UserID   string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid"})
		return
	}
	if req.TeamName == "" || req.UserID == "" {
		writeJSON(w, 400, map[string]string{"error": "team_name and user_id required"})
		return
	}
	team, err := h.Svc.RemoveUserFromTeam(req.TeamName, req.UserID)
	if err != nil {
		er := ErrResp{}
		er.Error.Code = "ERROR"
		er.Error.Message = err.Error()
		writeJSON(w, 400, er)
		return
	}
	writeJSON(w, 200, map[string]model.Team{"team": team})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/team/add", h.AddTeam).Methods("POST")
	r.HandleFunc("/team/get", h.GetTeam).Methods("GET")
	r.HandleFunc("/users/setIsActive", h.SetIsActive).Methods("POST")
	r.HandleFunc("/pullRequest/create", h.CreatePR).Methods("POST")
	r.HandleFunc("/pullRequest/merge", h.MergePR).Methods("POST")
	r.HandleFunc("/pullRequest/reassign", h.Reassign).Methods("POST")
	r.HandleFunc("/users/getReview", h.GetReviews).Methods("GET")
	r.HandleFunc("/team/addUser", h.AddUserToTeam).Methods("POST")
	r.HandleFunc("/team/removeUser", h.RemoveUserFromTeam).Methods("POST")

	r.HandleFunc("/health", h.Health).Methods("GET")
}
