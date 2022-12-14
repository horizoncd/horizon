package mock

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type HarborProject struct {
	ID           int
	Name         string
	Public       bool
	Members      []*ProjectMember
	Repositories []*ProjectRepository
}

type ProjectMember struct {
	Username string
	Role     int
}

type ProjectRepository struct {
	Name string
	Tags []string
}

type HarborServer struct {
	R         *mux.Router
	Projects  map[string]*HarborProject
	projectID int
}

func NewHarborServer() *HarborServer {
	r := mux.NewRouter()
	s := &HarborServer{
		R:         r,
		Projects:  map[string]*HarborProject{},
		projectID: 1,
	}
	r.Path("/api/v2.0/projects").Methods(http.MethodPost).HandlerFunc(s.CreateProject)
	r.Path("/api/v2.0/projects/{projectID}/members").
		Methods(http.MethodPost).HandlerFunc(s.AddSharedMemberForProject)
	r.Path("/api/v2.0/projects/{project}/repositories/{repository}").
		Methods(http.MethodDelete).HandlerFunc(s.DeleteRepository)
	r.Path("/api/v2.0/projects/{project}/repositories/{repository}/artifacts").
		Methods(http.MethodGet).HandlerFunc(s.ListImage)
	r.Path("/api/v2.0/projects/{project_name}/preheat/policies").
		Methods(http.MethodPost).HandlerFunc(s.PreheatProject)
	return s
}

func (s *HarborServer) PreheatProject(w http.ResponseWriter, r *http.Request) {
	type Preheat struct {
		Name       string `json:"name"`
		ProjectID  int    `json:"project_id"`
		ProviderID int    `json:"provider_id"`
		Filters    string `json:"filters"`
		Trigger    string `json:"trigger"`
		Enabled    bool   `json:"enabled"`
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.responseError(w, http.StatusInternalServerError, err)
		return
	}
	var preheat Preheat
	if err := json.Unmarshal(data, &preheat); err != nil {
		s.responseError(w, http.StatusBadRequest, err)
		return
	}

	for _, v := range s.Projects {
		if v.ID == preheat.ProjectID {
			w.WriteHeader(http.StatusCreated)
			return
		}
	}
	s.responseError(w, http.StatusBadRequest, fmt.Errorf("project not exist"))
}

func (s *HarborServer) CreateProject(w http.ResponseWriter, r *http.Request) {
	type Project struct {
		ProjectName string            `json:"project_name"`
		Metadata    map[string]string `json:"metadata"`
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.responseError(w, http.StatusInternalServerError, err)
		return
	}
	var p Project
	if err := json.Unmarshal(data, &p); err != nil {
		s.responseError(w, http.StatusBadRequest, err)
		return
	}
	if len(p.ProjectName) == 0 {
		s.responseError(w, http.StatusBadRequest, fmt.Errorf("project name cannot be empty"))
		return
	}
	for _, v := range s.Projects {
		if v.Name == p.ProjectName {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}
	publicBool := false
	publicStr := p.Metadata["public"]
	publicBool, err = strconv.ParseBool(publicStr)
	if err != nil {
		s.responseError(w, http.StatusBadRequest, err)
		return
	}
	project := &HarborProject{
		ID:           s.projectID,
		Name:         p.ProjectName,
		Public:       publicBool,
		Members:      nil,
		Repositories: nil,
	}
	s.Projects[strconv.Itoa(s.projectID)] = project
	w.Header().Set("Location", fmt.Sprintf("/api/projects/%d", s.projectID))
	s.projectID++
	w.WriteHeader(http.StatusCreated)
}

func (s *HarborServer) AddSharedMemberForProject(w http.ResponseWriter, r *http.Request) {
	type User struct {
		RoleID     int               `json:"role_id"`
		MemberUser map[string]string `json:"member_user"`
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.responseError(w, http.StatusInternalServerError, err)
		return
	}
	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		s.responseError(w, http.StatusBadRequest, err)
		return
	}
	vars := mux.Vars(r)
	projectID := vars["projectID"]
	project, ok := s.Projects[projectID]
	if !ok {
		s.responseError(w, http.StatusNotFound, fmt.Errorf("project with ID %s not found", projectID))
		return
	}
	username, ok := u.MemberUser["username"]
	if !ok {
		s.responseError(w, http.StatusBadRequest, fmt.Errorf("username cannot be empty"))
		return
	}
	for _, member := range project.Members {
		if member.Username == username {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}
	s.Projects[projectID].Members = append(s.Projects[projectID].Members, &ProjectMember{
		Username: username,
		Role:     u.RoleID,
	})
	w.WriteHeader(http.StatusCreated)
}

func (s *HarborServer) PushImage(projectName string, repository string, tag string) {
	if projectName == "" || repository == "" || tag == "" {
		return
	}
	projectID := ""
	for _, v := range s.Projects {
		if v.Name == projectName {
			projectID = strconv.Itoa(v.ID)
			break
		}
	}
	if projectID == "" {
		return
	}
	var index int
	var repo *ProjectRepository
	for k, v := range s.Projects[projectID].Repositories {
		if v.Name == repository {
			index = k
			repo = v
			break
		}
	}
	if repo != nil {
		s.Projects[projectID].Repositories[index].Tags = append(s.Projects[projectID].Repositories[index].Tags, tag)
	} else {
		s.Projects[projectID].Repositories = append(s.Projects[projectID].Repositories, &ProjectRepository{
			Name: repository,
			Tags: []string{tag},
		})
	}
}

func (s *HarborServer) DeleteRepository(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project, repository := vars["project"], vars["repository"]
	var projectID = ""
	for _, v := range s.Projects {
		if v.Name == project {
			projectID = strconv.Itoa(v.ID)
			break
		}
	}
	if projectID == "" {
		s.responseError(w, http.StatusNotFound, fmt.Errorf("project %s not found", project))
		return
	}
	index := -1
	for k, v := range s.Projects[projectID].Repositories {
		if v.Name == repository {
			index = k
			break
		}
	}
	if index == -1 {
		s.responseError(w, http.StatusNotFound, fmt.Errorf("repository %s not found", repository))
		return
	}
	if index == 0 {
		s.Projects[projectID].Repositories = s.Projects[projectID].Repositories[index+1:]
	} else {
		s.Projects[projectID].Repositories = append(s.Projects[projectID].Repositories[:index-1],
			s.Projects[projectID].Repositories[index+1:]...)
	}
	w.WriteHeader(http.StatusOK)
}

func (s *HarborServer) ListImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project, repository := vars["project"], vars["repository"]
	var projectID = ""
	for _, v := range s.Projects {
		if v.Name == project {
			projectID = strconv.Itoa(v.ID)
			break
		}
	}
	if projectID == "" {
		s.responseError(w, http.StatusNotFound, fmt.Errorf("project %s not found", project))
		return
	}
	index := -1
	for k, v := range s.Projects[projectID].Repositories {
		if v.Name == repository {
			index = k
			break
		}
	}
	if index == -1 {
		s.responseError(w, http.StatusNotFound, fmt.Errorf("repository %s not found", repository))
		return
	}
	type T struct{ Name string }
	type Artifacts struct {
		Tags []T `json:"tags"`
	}
	artifacts := make([]Artifacts, 0)
	for _, v := range s.Projects[projectID].Repositories[index].Tags {
		artifacts = append(artifacts, Artifacts{
			Tags: []T{
				{
					Name: v,
				},
			},
		})
	}
	b, err := json.Marshal(artifacts)
	if err != nil {
		s.responseError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func (s *HarborServer) responseError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
	}
}
