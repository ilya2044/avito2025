package service

import (
	"errors"
	"math/rand"
	"time"

	"github.com/ilya2044/avito2025/internal/model"
	"github.com/ilya2044/avito2025/internal/storage"
)

var (
	ErrTeamNotFound = errors.New("team or author not found")
	ErrPRExists     = errors.New("pr exists")
	ErrPRMerged     = errors.New("pr merged")
	ErrNotAssigned  = errors.New("not assigned")
	ErrNoCandidate  = errors.New("no candidate")
)

type Service struct {
	Repo *storage.Repository
	Rand *rand.Rand
}

func NewService(r *storage.Repository) *Service {
	return &Service{Repo: r, Rand: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (s *Service) CreateTeam(t model.Team) error {
	return s.Repo.CreateTeam(t)
}

func (s *Service) GetTeam(name string) (model.Team, error) {
	return s.Repo.GetTeam(name)
}

func (s *Service) SetUserIsActive(userID string, isActive bool) (model.User, error) {
	return s.Repo.SetUserIsActive(userID, isActive)
}

func (s *Service) CreatePullRequest(pr model.PullRequest) (model.PullRequest, error) {
	_, err := s.Repo.GetPullRequest(pr.PullRequestID)
	if err == nil {
		return model.PullRequest{}, ErrPRExists
	}
	author, err := s.Repo.GetUser(pr.AuthorID)
	if err != nil {
		return model.PullRequest{}, ErrTeamNotFound
	}
	exclude := []string{author.UserID}
	candidates, err := s.Repo.GetActiveTeamMembers(author.TeamName, exclude)
	if err != nil {
		return model.PullRequest{}, err
	}
	assigned := pickN(candidates, s.Rand, 2)
	uids := []string{}
	for _, u := range assigned {
		uids = append(uids, u.UserID)
	}
	if err := s.Repo.CreatePullRequest(pr, uids); err != nil {
		return model.PullRequest{}, err
	}
	return s.Repo.GetPullRequest(pr.PullRequestID)
}

func pickN(cands []model.User, r *rand.Rand, n int) []model.User {
	if len(cands) == 0 {
		return []model.User{}
	}
	if len(cands) <= n {
		return cands
	}
	res := make([]model.User, n)
	perm := r.Perm(len(cands))
	for i := 0; i < n; i++ {
		res[i] = cands[perm[i]]
	}
	return res
}

func (s *Service) MergePullRequest(prID string) (model.PullRequest, error) {
	return s.Repo.MergePullRequest(prID)
}

func (s *Service) ReassignReviewer(prID, oldUserID string) (model.PullRequest, string, error) {
	pr, err := s.Repo.GetPullRequest(prID)
	if err != nil {
		return model.PullRequest{}, "", err
	}
	if pr.Status == "MERGED" {
		return model.PullRequest{}, "", ErrPRMerged
	}
	assignedMap := map[string]bool{}
	for _, a := range pr.AssignedReviewers {
		assignedMap[a] = true
	}
	if !assignedMap[oldUserID] {
		return model.PullRequest{}, "", ErrNotAssigned
	}
	oldUser, err := s.Repo.GetUser(oldUserID)
	if err != nil {
		return model.PullRequest{}, "", err
	}
	exclude := append(pr.AssignedReviewers, pr.AuthorID)
	cands, err := s.Repo.GetActiveTeamMembers(oldUser.TeamName, exclude)
	if err != nil {
		return model.PullRequest{}, "", err
	}
	if len(cands) == 0 {
		return model.PullRequest{}, "", ErrNoCandidate
	}
	new := cands[s.Rand.Intn(len(cands))].UserID
	if err := s.Repo.ReplaceReviewer(prID, oldUserID, new); err != nil {
		return model.PullRequest{}, "", err
	}
	updatedPR, err := s.Repo.GetPullRequest(prID)
	return updatedPR, new, err
}

func (s *Service) GetPRsByReviewer(userID string) ([]model.PullRequestShort, error) {
	return s.Repo.GetPRsByReviewer(userID)
}

func (s *Service) AddUserToTeam(teamName string, u model.User) (model.Team, error) {
	return s.Repo.AddUserToTeam(teamName, u)
}

func (s *Service) RemoveUserFromTeam(teamName, userID string) (model.Team, error) {
	return s.Repo.RemoveUserFromTeam(teamName, userID)
}
