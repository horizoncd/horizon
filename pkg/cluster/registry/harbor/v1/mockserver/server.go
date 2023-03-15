package mockserver

import (
	"fmt"
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

	r.Path("/api/repositories/{project}/{repository:[0-9a-zA-Z/-]+}").
		Methods(http.MethodDelete).HandlerFunc(s.DeleteRepository)
	return s
}

func (s *HarborServer) CreateProject(projectName string, metadata map[string]string) {
	if len(projectName) == 0 {
		return
	}

	for _, v := range s.Projects {
		if v.Name == projectName {
			return
		}
	}

	publicBool := false
	if publicStr, ok := metadata["public"]; ok {
		publicBool, _ = strconv.ParseBool(publicStr)
	}

	project := &HarborProject{
		ID:           s.projectID,
		Name:         projectName,
		Public:       publicBool,
		Members:      nil,
		Repositories: nil,
	}
	s.Projects[strconv.Itoa(s.projectID)] = project
	s.projectID++
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

func (s *HarborServer) responseError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
	}
}
