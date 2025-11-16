package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ilya2044/avito2025/internal/model"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) ApplyMigrations(migrationSQL string) error {
	_, err := r.DB.Exec(migrationSQL)
	return err
}

func (r *Repository) CreateTeam(team model.Team) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE team_name=$1)", team.TeamName).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("team %s already exists", team.TeamName)
	}

	for _, m := range team.Members {
		var userExists bool
		err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1)", m.UserID).Scan(&userExists)
		if err != nil {
			return err
		}
		if userExists {
			return fmt.Errorf("user %s already exists", m.UserID)
		}
	}

	_, err = tx.Exec("INSERT INTO teams(team_name) VALUES($1)", team.TeamName)
	if err != nil {
		return err
	}

	for _, m := range team.Members {
		_, err = tx.Exec(`INSERT INTO users(user_id, username, team_name, is_active)
			VALUES($1,$2,$3,$4)`,
			m.UserID, m.Username, team.TeamName, m.IsActive)
		if err != nil {
			return fmt.Errorf("cannot add user %s: %w", m.UserID, err)
		}
	}

	return tx.Commit()
}

func (r *Repository) GetTeam(teamName string) (model.Team, error) {
	var t model.Team
	t.TeamName = teamName
	rows, err := r.DB.Query("SELECT user_id, username, is_active FROM users WHERE team_name=$1", teamName)
	if err != nil {
		return t, err
	}
	defer rows.Close()
	members := []model.TeamMember{}
	for rows.Next() {
		var m model.TeamMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return t, err
		}
		members = append(members, m)
	}
	t.Members = members
	return t, nil
}

func (r *Repository) SetUserIsActive(userID string, isActive bool) (model.User, error) {
	_, err := r.DB.Exec("UPDATE users SET is_active=$1 WHERE user_id=$2", isActive, userID)
	if err != nil {
		return model.User{}, err
	}
	var u model.User
	err = r.DB.QueryRow("SELECT user_id, username, team_name, is_active FROM users WHERE user_id=$1", userID).
		Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive)
	return u, err
}

func (r *Repository) GetUser(userID string) (model.User, error) {
	var u model.User
	err := r.DB.QueryRow("SELECT user_id, username, team_name, is_active FROM users WHERE user_id=$1", userID).
		Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive)
	return u, err
}

func (r *Repository) CreatePullRequest(pr model.PullRequest, assigned []string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO pull_requests(pull_request_id, pull_request_name, author_id, status) VALUES($1,$2,$3,'OPEN')`,
		pr.PullRequestID, pr.PullRequestName, pr.AuthorID)
	if err != nil {
		return err
	}
	for _, uid := range assigned {
		_, err = tx.Exec("INSERT INTO pr_reviewers(pr_id, user_id) VALUES($1,$2)", pr.PullRequestID, uid)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) GetPullRequest(prID string) (model.PullRequest, error) {
	var pr model.PullRequest
	var createdAt, mergedAt sql.NullTime
	err := r.DB.QueryRow("SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at FROM pull_requests WHERE pull_request_id=$1", prID).
		Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &createdAt, &mergedAt)
	if err != nil {
		return pr, err
	}
	if createdAt.Valid {
		t := createdAt.Time
		pr.CreatedAt = &t
	}
	if mergedAt.Valid {
		t := mergedAt.Time
		pr.MergedAt = &t
	}
	rows, err := r.DB.Query("SELECT user_id FROM pr_reviewers WHERE pr_id=$1 ORDER BY user_id", prID)
	if err != nil {
		return pr, err
	}
	defer rows.Close()
	revs := []string{}
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return pr, err
		}
		revs = append(revs, uid)
	}
	pr.AssignedReviewers = revs
	return pr, nil
}

func (r *Repository) MergePullRequest(prID string) (model.PullRequest, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return model.PullRequest{}, err
	}
	defer tx.Rollback()
	var status string
	var mergedAt sql.NullTime
	err = tx.QueryRow("SELECT status, merged_at FROM pull_requests WHERE pull_request_id=$1 FOR UPDATE", prID).Scan(&status, &mergedAt)
	if err != nil {
		return model.PullRequest{}, err
	}
	if status == "MERGED" {
		return r.GetPullRequest(prID)
	}
	now := time.Now().UTC()
	_, err = tx.Exec("UPDATE pull_requests SET status='MERGED', merged_at=$1 WHERE pull_request_id=$2", now, prID)
	if err != nil {
		return model.PullRequest{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.PullRequest{}, err
	}
	return r.GetPullRequest(prID)
}

func (r *Repository) GetActiveTeamMembers(teamName string, exclude []string) ([]model.User, error) {
	excludeMap := map[string]bool{}
	for _, e := range exclude {
		excludeMap[e] = true
	}
	rows, err := r.DB.Query("SELECT user_id, username, team_name, is_active FROM users WHERE team_name=$1 AND is_active=true", teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		if excludeMap[u.UserID] {
			continue
		}
		res = append(res, u)
	}
	return res, nil
}

func (r *Repository) IsUserAssignedToPR(prID, userID string) (bool, error) {
	var cnt int
	err := r.DB.QueryRow("SELECT COUNT(1) FROM pr_reviewers WHERE pr_id=$1 AND user_id=$2", prID, userID).Scan(&cnt)
	return cnt > 0, err
}

func (r *Repository) ReplaceReviewer(prID, oldUserID, newUserID string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM pr_reviewers WHERE pr_id=$1 AND user_id=$2", prID, oldUserID)
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO pr_reviewers(pr_id, user_id) VALUES($1,$2)", prID, newUserID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) GetPRsByReviewer(userID string) ([]model.PullRequestShort, error) {
	rows, err := r.DB.Query(`
SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
FROM pull_requests pr
JOIN pr_reviewers rr ON rr.pr_id = pr.pull_request_id
WHERE rr.user_id = $1
ORDER BY pr.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []model.PullRequestShort{}
	for rows.Next() {
		var p model.PullRequestShort
		if err := rows.Scan(&p.PullRequestID, &p.PullRequestName, &p.AuthorID, &p.Status); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}

func (r *Repository) AddUserToTeam(teamName string, u model.User) (model.Team, error) {
	var exists int
	err := r.DB.QueryRow("SELECT COUNT(1) FROM users WHERE user_id=$1", u.UserID).Scan(&exists)
	if err != nil {
		return model.Team{}, err
	}
	if exists > 0 {
		return model.Team{}, fmt.Errorf("user_id %s already exists", u.UserID)
	}
	_, err = r.DB.Exec(`INSERT INTO users(user_id, username, team_name, is_active) VALUES($1,$2,$3,$4)`,
		u.UserID, u.Username, teamName, u.IsActive)
	if err != nil {
		return model.Team{}, err
	}
	return r.GetTeam(teamName)
}

func (r *Repository) RemoveUserFromTeam(teamName, userID string) (model.Team, error) {
	res, err := r.DB.Exec("DELETE FROM users WHERE user_id=$1 AND team_name=$2", userID, teamName)
	if err != nil {
		return model.Team{}, err
	}
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return model.Team{}, fmt.Errorf("user_id %s not found in team %s", userID, teamName)
	}
	return r.GetTeam(teamName)
}
